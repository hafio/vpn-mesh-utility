package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The first admin from the web UI (amendment A52). While the control plane
// has no enabled admin, each start makes a one-time setup code, which the
// control plane logs; the setup page takes it with the first admin's
// username and password. The Service keeps only the code's hash, in memory,
// and forgets it once the first admin exists.

// ActorSetup is the audit actor of a user made on the setup page.
const ActorSetup = "setup"

var (
	// ErrSetupCode is a setup code that is not the latest one logged.
	ErrSetupCode = errors.New("the setup code is wrong or from an earlier start; use the latest one in the control plane's log")
	// ErrSetupDone refuses the setup page once an admin exists.
	ErrSetupDone = errors.New("the control plane has an admin already; sign in, or ask an admin for an account")
)

// SetupOpen reports whether the control plane has no enabled admin yet, so
// its setup page is open.
func (s *Service) SetupOpen(ctx context.Context) (bool, error) {
	var n int
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		var err error
		n, err = tx.Admins()
		return err
	})
	return n == 0 && err == nil, err
}

// SetupCode makes a new setup code while the setup page is open, replacing
// the one before, and returns it in groups of four; "" once an admin exists.
func (s *Service) SetupCode(ctx context.Context) (string, error) {
	open, err := s.SetupOpen(ctx)
	if err != nil || !open {
		return "", err
	}
	b := make([]byte, 20)
	_, _ = rand.Read(b) // crypto/rand.Read never fails (Go 1.24+)
	code := strings.ToLower(base32.StdEncoding.EncodeToString(b))
	sum := sha256.Sum256([]byte(code))
	s.setupMu.Lock()
	s.setup = sum[:]
	s.setupMu.Unlock()
	groups := make([]string, 0, len(code)/4)
	for i := 0; i < len(code); i += 4 {
		groups = append(groups, code[i:i+4])
	}
	return strings.Join(groups, "-"), nil
}

// FirstAdmin makes the control plane's first admin with the setup code and
// starts their sign-in: the answer is a session waiting for the
// authenticator's enrolment, as the first sign-in of a user made by
// `meshctl user add`. Wrong codes count as failed sign-ins from ip; once the
// first admin exists no code is right.
func (s *Service) FirstAdmin(ctx context.Context, ip, code, username, password string) (Pending, error) {
	now := s.now()
	key := "(setup)" + ip // never a username; keeps one address's guesses from locking out another's
	if err := s.lock.Check(ip, key, now); err != nil {
		return Pending{}, err
	}
	c, err := validate.SetupCode("setup_code", code)
	if err != nil {
		return Pending{}, err
	}
	if !s.setupMatches(c) {
		s.lock.Fail(ip, key, now)
		return Pending{}, ErrSetupCode
	}
	name, err := validate.Username("username", username)
	if err != nil {
		return Pending{}, err
	}
	if err := CheckPassword(password); err != nil {
		return Pending{}, err
	}
	u := store.User{Username: name, PasswordHash: HashPassword(password, s.params), Admin: true}
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		n, err := tx.Admins()
		if err != nil {
			return err
		}
		if n > 0 {
			return ErrSetupDone // `meshctl user add` was quicker
		}
		return insertUser(tx, ActorSetup, &u)
	})
	if err != nil {
		return Pending{}, err
	}
	s.lock.Succeed(ip, key)
	s.setupMu.Lock()
	s.setup = nil
	s.setupMu.Unlock()
	return s.Login(ctx, ip, name, password)
}

// setupMatches compares a canonical code with the kept one in constant time.
func (s *Service) setupMatches(code string) bool {
	sum := sha256.Sum256([]byte(code))
	s.setupMu.Lock()
	defer s.setupMu.Unlock()
	return s.setup != nil && subtle.ConstantTimeCompare(sum[:], s.setup) == 1
}
