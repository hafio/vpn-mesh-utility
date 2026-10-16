// Package auth is the web UI's authentication (spec S8.4): argon2id
// passwords, TOTP as the second factor, server-side sessions, CSRF tokens and
// login lockout, plus the client-address and rate-limit helpers the agent API
// shares. The UI (M2b) drives the sign-in flow through Service; the `meshctl
// user` commands manage accounts through it.
//
// Sign-in is two steps. Login checks the password and opens a session that
// may only finish the second factor, for TOTPWindow. SecondFactor (or, for a
// user without TOTP yet, BeginEnrollment and ConfirmEnrollment) promotes it to
// a full session. Authorize accepts only full sessions.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// DB is the part of *store.Store auth needs.
type DB interface {
	Tx(ctx context.Context, fn func(*store.Tx) error) error
}

// Config configures a Service.
type Config struct {
	DB DB
	// Sealer opens and seals TOTP secrets; nil limits the service to account
	// management (the CLI).
	Sealer *Sealer
	Now    func() time.Time
	// Params is the argon2id cost; zero means DefaultParams. Tests use a
	// cheap one.
	Params Params
	// Issuer names the mesh in authenticator apps; "" means "meshctl".
	Issuer string
}

// Service signs users in and manages their accounts.
type Service struct {
	db     DB
	sealer *Sealer
	now    func() time.Time
	params Params
	issuer string
	lock   *Lockout
	// dummy is verified when the username does not exist, so the answer
	// takes as long as a wrong password.
	dummy string
}

// Errors a sign-in can end with. ErrBadLogin is the same for an unknown
// user, a wrong password and a disabled account, so it does not tell an
// attacker which usernames exist.
var (
	ErrBadLogin   = errors.New("wrong username or password")
	ErrNoSession  = errors.New("not signed in, or the session expired; sign in again")
	ErrNeedsTOTP  = errors.New("this account has no authenticator yet; enroll one first")
	ErrNoSealer   = errors.New("this command cannot handle TOTP secrets; use the web UI")
	ErrHasTOTP    = errors.New("this account already has an authenticator; an administrator can reset it with `meshctl user reset-totp`")
	ErrUserExists = errors.New("a user with this name already exists")
	ErrNoUser     = errors.New("user does not exist")
	// ErrWrongPassword is a password change with the wrong current password.
	ErrWrongPassword = errors.New("the current password is wrong")
	// ErrNoEnrollment is a code sent for an enrolment that was never started.
	ErrNoEnrollment = errors.New("no enrolment in progress; start it again")
)

// New returns a service.
func New(c Config) *Service {
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.Params == (Params{}) {
		c.Params = DefaultParams()
	}
	if c.Issuer == "" {
		c.Issuer = "meshctl"
	}
	return &Service{db: c.DB, sealer: c.Sealer, now: c.Now, params: c.Params, issuer: c.Issuer, lock: NewLockout(),
		dummy: HashPassword("not a real password", c.Params)}
}

// Pending is a session that passed the password and waits for the second
// factor.
type Pending struct {
	Cookie  string
	Session Session
	// NeedsEnrollment means the user has no authenticator yet: the UI shows
	// BeginEnrollment's QR code instead of asking for a code.
	NeedsEnrollment bool
}

// Login checks the first factor.
func (s *Service) Login(ctx context.Context, ip, username, password string) (Pending, error) {
	now := s.now()
	if err := s.lock.Check(ip, username, now); err != nil {
		return Pending{}, err
	}
	var (
		u      store.User
		found  bool
		ok     bool
		rehash bool
	)
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		var err error
		u, err = tx.UserByName(username)
		switch {
		case errors.Is(err, store.ErrNotFound):
			_, _, _ = VerifyPassword(s.dummy, password, s.params) // equal time for unknown users
			return nil
		case err != nil:
			return err
		}
		found = true
		if ok, rehash, err = VerifyPassword(u.PasswordHash, password, s.params); err != nil {
			return err
		}
		if ok && rehash {
			return tx.SetPassword(u.ID, HashPassword(password, s.params))
		}
		return nil
	})
	if err != nil {
		return Pending{}, err
	}
	if !found || !ok || u.Disabled {
		s.lock.Fail(ip, username, now)
		return Pending{}, ErrBadLogin
	}
	s.lock.Succeed(ip, username)
	cookie, sess, err := s.create(ctx, u, now)
	if err != nil {
		return Pending{}, err
	}
	return Pending{Cookie: cookie, Session: sess, NeedsEnrollment: !u.TOTPEnrolled}, nil
}

func (s *Service) create(ctx context.Context, u store.User, now time.Time) (string, Session, error) {
	cookie := newSecret()
	row := store.Session{IDHash: hashCookie(cookie), UserID: u.ID, CreatedAt: now.Unix(), LastSeen: now.Unix(),
		ExpiresAt: now.Add(TOTPWindow).Unix(), CSRF: newSecret()}
	if err := s.db.Tx(ctx, func(tx *store.Tx) error { return tx.InsertSession(row) }); err != nil {
		return "", Session{}, err
	}
	return cookie, session(row, u), nil
}

func session(row store.Session, u store.User) Session {
	return Session{UserID: u.ID, Username: u.Username, CSRF: row.CSRF, TOTPOK: row.TOTPOK, ExpiresAt: time.Unix(row.ExpiresAt, 0)}
}

// load reads a live session and its user inside tx.
func load(tx *store.Tx, cookie string, now time.Time) (store.Session, store.User, error) {
	if !reCookie.MatchString(cookie) {
		return store.Session{}, store.User{}, ErrNoSession
	}
	row, err := tx.SessionByHash(hashCookie(cookie))
	if errors.Is(err, store.ErrNotFound) {
		return store.Session{}, store.User{}, ErrNoSession
	}
	if err != nil {
		return store.Session{}, store.User{}, err
	}
	if now.Unix() >= row.ExpiresAt {
		return store.Session{}, store.User{}, ErrNoSession
	}
	u, err := tx.UserByID(row.UserID)
	if errors.Is(err, store.ErrNotFound) || (err == nil && u.Disabled) {
		return store.Session{}, store.User{}, ErrNoSession
	}
	return row, u, err
}

// fullExpiry is when a full session expires if it is not used again: the idle
// limit, capped by the absolute one.
func fullExpiry(row store.Session, now time.Time) int64 {
	return min(now.Add(IdleTimeout).Unix(), time.Unix(row.CreatedAt, 0).Add(AbsoluteTimeout).Unix())
}

// SecondFactor checks a TOTP code for a pending session and promotes it.
func (s *Service) SecondFactor(ctx context.Context, ip, cookie, code string) (Session, error) {
	if s.sealer == nil {
		return Session{}, ErrNoSealer
	}
	now := s.now()
	var out Session
	var failedUser string
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		row, u, err := load(tx, cookie, now)
		if err != nil {
			return err
		}
		if err := s.lock.Check(ip, u.Username, now); err != nil {
			return err
		}
		if row.TOTPOK {
			out = session(row, u)
			return nil
		}
		if !u.TOTPEnrolled {
			return ErrNeedsTOTP
		}
		secret, err := s.sealer.Open(u.ID, u.TOTPSecret)
		if err != nil {
			return err
		}
		step, err := matchStep(secret, code, now)
		if err != nil {
			failedUser = u.Username
			return err
		}
		if err := tx.UseTOTPStep(u.ID, step); err != nil {
			failedUser = u.Username
			return err
		}
		row.TOTPOK, row.LastSeen, row.ExpiresAt = true, now.Unix(), fullExpiry(row, now)
		if err := tx.UpdateSession(row); err != nil {
			return err
		}
		out = session(row, u)
		return nil
	})
	if failedUser != "" {
		s.lock.Fail(ip, failedUser, now)
	}
	if err != nil {
		return Session{}, err
	}
	s.lock.Succeed(ip, out.Username)
	return out, nil
}

// BeginEnrollment returns the authenticator secret a pending session's user
// is enrolling, creating one if there is none yet. Reloading the enrolment
// page therefore shows the same QR code the user may already have scanned.
func (s *Service) BeginEnrollment(ctx context.Context, cookie string) (*otp.Key, error) {
	if s.sealer == nil {
		return nil, ErrNoSealer
	}
	var key *otp.Key
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		_, u, err := load(tx, cookie, s.now())
		if err != nil {
			return err
		}
		if u.TOTPEnrolled {
			return ErrHasTOTP
		}
		if u.TOTPSecret != nil {
			key, err = s.keyOf(u)
			return err
		}
		if key, err = newTOTPKey(s.issuer, u.Username, ""); err != nil {
			return err
		}
		return tx.SetTOTP(u.ID, s.sealer.Seal(u.ID, key.Secret()), false)
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

// keyOf rebuilds the otpauth key of a user's sealed, unconfirmed secret.
func (s *Service) keyOf(u store.User) (*otp.Key, error) {
	secret, err := s.sealer.Open(u.ID, u.TOTPSecret)
	if err != nil {
		return nil, err
	}
	return newTOTPKey(s.issuer, u.Username, secret)
}

// ConfirmEnrollment checks the first code from the new authenticator,
// completes the enrolment and promotes the session.
func (s *Service) ConfirmEnrollment(ctx context.Context, ip, cookie, code string) (Session, error) {
	if s.sealer == nil {
		return Session{}, ErrNoSealer
	}
	now := s.now()
	var out Session
	var failedUser string
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		row, u, err := load(tx, cookie, now)
		if err != nil {
			return err
		}
		if err := s.lock.Check(ip, u.Username, now); err != nil {
			return err
		}
		if u.TOTPEnrolled {
			return ErrHasTOTP
		}
		if u.TOTPSecret == nil {
			return ErrNoEnrollment
		}
		secret, err := s.sealer.Open(u.ID, u.TOTPSecret)
		if err != nil {
			return err
		}
		step, err := matchStep(secret, code, now)
		if err != nil {
			failedUser = u.Username
			return err
		}
		if err := tx.SetTOTP(u.ID, u.TOTPSecret, true); err != nil {
			return err
		}
		if err := tx.UseTOTPStep(u.ID, step); err != nil {
			return err
		}
		row.TOTPOK, row.LastSeen, row.ExpiresAt = true, now.Unix(), fullExpiry(row, now)
		if err := tx.UpdateSession(row); err != nil {
			return err
		}
		u.TOTPEnrolled = true
		out = session(row, u)
		return tx.Audit("user:"+u.Username, u.ID, "user.totp_enrolled", u.Username, nil)
	})
	if failedUser != "" {
		s.lock.Fail(ip, failedUser, now)
	}
	if err != nil {
		return Session{}, err
	}
	return out, nil
}

// Current returns the live session behind a cookie -- full or still waiting
// for its second factor -- without extending it. The sign-in pages use it.
func (s *Service) Current(ctx context.Context, cookie string) (Session, error) {
	var out Session
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		row, u, err := load(tx, cookie, s.now())
		if err != nil {
			return err
		}
		out = session(row, u)
		return nil
	})
	return out, err
}

// Peek returns the full session behind a cookie without extending it, for
// requests a page makes on its own (the dashboard's refresh): only what the
// user does keeps a session alive.
func (s *Service) Peek(ctx context.Context, cookie string) (Session, error) {
	out, err := s.Current(ctx, cookie)
	if err == nil && !out.TOTPOK {
		return Session{}, ErrNoSession
	}
	return out, err
}

// Authorize returns the full session behind a cookie and slides its idle
// expiry. A pending, expired or unknown session, or a disabled user, is
// ErrNoSession.
func (s *Service) Authorize(ctx context.Context, cookie string) (Session, error) {
	now := s.now()
	var out Session
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		row, u, err := load(tx, cookie, now)
		if err != nil {
			return err
		}
		if !row.TOTPOK {
			return ErrNoSession
		}
		row.LastSeen, row.ExpiresAt = now.Unix(), fullExpiry(row, now)
		if err := tx.UpdateSession(row); err != nil {
			return err
		}
		out = session(row, u)
		return nil
	})
	return out, err
}

// Logout ends the session behind a cookie.
func (s *Service) Logout(ctx context.Context, cookie string) error {
	if !reCookie.MatchString(cookie) {
		return nil
	}
	return s.db.Tx(ctx, func(tx *store.Tx) error { return tx.DeleteSession(hashCookie(cookie)) })
}

// AddUser creates an account; the user enrolls an authenticator at their
// first sign-in.
func (s *Service) AddUser(ctx context.Context, actor, username, password string) (store.User, error) {
	name, err := validate.Username("username", username)
	if err != nil {
		return store.User{}, err
	}
	if err := CheckPassword(password); err != nil {
		return store.User{}, err
	}
	u := store.User{Username: name, PasswordHash: HashPassword(password, s.params)}
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		if err := tx.InsertUser(&u); err != nil {
			if errors.Is(err, store.ErrConflict) {
				return fmt.Errorf("%w: %q", ErrUserExists, name)
			}
			return err
		}
		return tx.Audit(actor, 0, "user.add", name, nil)
	})
	if err != nil {
		return store.User{}, err
	}
	return u, nil
}

// ChangePassword replaces the signed-in user's password after checking the
// current one, and ends all of their sessions, this one included.
func (s *Service) ChangePassword(ctx context.Context, who Session, current, next string) error {
	if err := CheckPassword(next); err != nil {
		return err
	}
	var ok bool
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		u, err := tx.UserByID(who.UserID)
		if err != nil {
			return err
		}
		ok, _, err = VerifyPassword(u.PasswordHash, current, s.params)
		return err
	})
	if err != nil {
		return err
	}
	if !ok {
		return ErrWrongPassword
	}
	return s.SetPassword(ctx, "user:"+who.Username, who.Username, next)
}

// Users lists the accounts.
func (s *Service) Users(ctx context.Context) ([]store.User, error) {
	var out []store.User
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		var err error
		out, err = tx.Users()
		return err
	})
	return out, err
}

// SetPassword replaces a user's password and ends their sessions.
func (s *Service) SetPassword(ctx context.Context, actor, username, password string) error {
	if err := CheckPassword(password); err != nil {
		return err
	}
	return s.updateUser(ctx, actor, username, "user.passwd", func(tx *store.Tx, u store.User) error {
		return tx.SetPassword(u.ID, HashPassword(password, s.params))
	})
}

// ResetTOTP removes a user's authenticator (they enroll a new one at their
// next sign-in) and ends their sessions.
func (s *Service) ResetTOTP(ctx context.Context, actor, username string) error {
	return s.updateUser(ctx, actor, username, "user.reset_totp", func(tx *store.Tx, u store.User) error {
		return tx.SetTOTP(u.ID, nil, false)
	})
}

// SetDisabled disables or enables a user; disabling ends their sessions.
func (s *Service) SetDisabled(ctx context.Context, actor, username string, disabled bool) error {
	action := "user.enable"
	if disabled {
		action = "user.disable"
	}
	return s.updateUser(ctx, actor, username, action, func(tx *store.Tx, u store.User) error {
		return tx.SetUserDisabled(u.ID, disabled)
	})
}

func (s *Service) updateUser(ctx context.Context, actor, username, action string, change func(*store.Tx, store.User) error) error {
	name, err := validate.Username("username", username)
	if err != nil {
		return err
	}
	return s.db.Tx(ctx, func(tx *store.Tx) error {
		u, err := tx.UserByName(name)
		if errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("%w: %q", ErrNoUser, name)
		}
		if err != nil {
			return err
		}
		if err := change(tx, u); err != nil {
			return err
		}
		if err := tx.DeleteUserSessions(u.ID); err != nil {
			return err
		}
		return tx.Audit(actor, 0, action, name, nil)
	})
}
