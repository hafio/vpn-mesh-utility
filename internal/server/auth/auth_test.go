package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

const (
	ip       = "192.0.2.1"
	password = "correct horse battery"
)

type rig struct {
	svc   *Service
	st    *store.Store
	clock *testfix.Clock
}

func newRig(t *testing.T) *rig {
	t.Helper()
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "mesh.db"), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	sealer, err := NewSealer(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	r := &rig{svc: New(Config{DB: st, Sealer: sealer, Now: clock.Now, Params: cheap}), st: st, clock: clock}
	if _, err := r.svc.AddUser(context.Background(), "cli", "admin", password); err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *rig) user(t *testing.T, name string) store.User {
	t.Helper()
	var u store.User
	if err := r.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		u, err = tx.UserByName(name)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return u
}

// enroll signs admin in and completes TOTP enrolment; it returns the
// session cookie and the authenticator secret.
func (r *rig) enroll(t *testing.T) (string, string) {
	t.Helper()
	ctx := context.Background()
	p, err := r.svc.Login(ctx, ip, "admin", password)
	if err != nil || !p.NeedsEnrollment || p.Session.TOTPOK {
		t.Fatalf("Login = %+v, %v", p, err)
	}
	key, err := r.svc.BeginEnrollment(ctx, p.Cookie)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key.URL(), "otpauth://totp/meshctl:admin?") {
		t.Fatalf("otpauth URL %q", key.URL())
	}
	s, err := r.svc.ConfirmEnrollment(ctx, ip, p.Cookie, code(t, key.Secret(), r.clock.Now()))
	if err != nil || !s.TOTPOK || s.Username != "admin" || s.CSRF == "" {
		t.Fatalf("ConfirmEnrollment = %+v, %v", s, err)
	}
	return p.Cookie, key.Secret()
}

func TestSignInWithEnrolment(t *testing.T) {
	r := newRig(t)
	cookie, _ := r.enroll(t)
	s, err := r.svc.Authorize(context.Background(), cookie)
	if err != nil || s.Username != "admin" {
		t.Fatalf("Authorize = %+v, %v", s, err)
	}
	if u := r.user(t, "admin"); !u.TOTPEnrolled || bytes.Contains(u.TOTPSecret, []byte("otpauth")) {
		t.Fatalf("user after enrolment %+v", u)
	}
	if _, err := r.svc.BeginEnrollment(context.Background(), cookie); !errors.Is(err, ErrHasTOTP) {
		t.Fatalf("a second enrolment = %v", err)
	}
}

func TestSecondFactorAndReplay(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	_, secret := r.enroll(t)
	r.clock.Advance(time.Minute)
	p, err := r.svc.Login(ctx, ip, "admin", password)
	if err != nil || p.NeedsEnrollment {
		t.Fatalf("Login = %+v, %v", p, err)
	}
	if _, err := r.svc.Authorize(ctx, p.Cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("a pending session must not authorize: %v", err)
	}
	if _, err := r.svc.SecondFactor(ctx, ip, p.Cookie, "000000"); !errors.Is(err, ErrBadCode) {
		t.Fatalf("wrong code = %v", err)
	}
	c := code(t, secret, r.clock.Now())
	if _, err := r.svc.SecondFactor(ctx, ip, p.Cookie, c); err != nil {
		t.Fatal(err)
	}
	// The same code on another login is a replay.
	p2, err := r.svc.Login(ctx, ip, "admin", password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.SecondFactor(ctx, ip, p2.Cookie, c); !errors.Is(err, store.ErrTOTPReplay) {
		t.Fatalf("a reused code = %v, want ErrTOTPReplay", err)
	}
}

func TestLoginRefusals(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	if err := r.svc.SetDisabled(ctx, "cli", "admin", true); err != nil {
		t.Fatal(err)
	}
	for name, attempt := range map[string][2]string{
		"unknown user":   {"nobody", password},
		"wrong password": {"admin", "wrong horse battery"},
		"disabled user":  {"admin", password},
	} {
		if _, err := r.svc.Login(ctx, ip, attempt[0], attempt[1]); !errors.Is(err, ErrBadLogin) {
			t.Errorf("%s: Login = %v, want the one ErrBadLogin", name, err)
		}
	}
}

func TestLoginLocksOutAfterFiveFailures(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	for i := 0; i < ipFailures; i++ {
		if _, err := r.svc.Login(ctx, ip, "admin", "wrong horse battery"); !errors.Is(err, ErrBadLogin) {
			t.Fatalf("attempt %d = %v", i+1, err)
		}
	}
	var le *LockoutError
	if _, err := r.svc.Login(ctx, ip, "admin", password); !errors.As(err, &le) {
		t.Fatalf("the right password while locked out = %v", err)
	}
	r.clock.Advance(ipLockout)
	if _, err := r.svc.Login(ctx, ip, "admin", password); err != nil {
		t.Fatalf("after the lockout: %v", err)
	}
}

func TestLoginUpgradesAnOldHash(t *testing.T) {
	r := newRig(t)
	old := HashPassword(password, Params{Time: 2, MemoryKiB: 32, Threads: 1})
	// Read the id first: the store has one connection, so a lookup inside
	// the transaction below would wait for it forever.
	id := r.user(t, "admin").ID
	if err := r.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.SetPassword(id, old) }); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Login(context.Background(), ip, "admin", password); err != nil {
		t.Fatal(err)
	}
	h := r.user(t, "admin").PasswordHash
	if h == old || !strings.Contains(h, "m=64,t=1,p=1") {
		t.Fatalf("hash after login %q, want it re-made with the service's parameters", h)
	}
}

func TestSessionExpiry(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	cookie, _ := r.enroll(t)

	// Idle: each use slides the expiry, 12 h of silence ends it.
	r.clock.Advance(IdleTimeout - time.Minute)
	if _, err := r.svc.Authorize(ctx, cookie); err != nil {
		t.Fatalf("within the idle window: %v", err)
	}
	r.clock.Advance(IdleTimeout)
	if _, err := r.svc.Authorize(ctx, cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("after 12 h idle: %v", err)
	}

	// Absolute: even a session used every hour ends after 7 days.
	cookie, _ = r.secondLogin(t)
	start := r.clock.Now()
	for r.clock.Now().Sub(start) < AbsoluteTimeout-time.Hour {
		r.clock.Advance(time.Hour)
		if _, err := r.svc.Authorize(ctx, cookie); err != nil {
			t.Fatalf("active session ended early at %s: %v", r.clock.Now().Sub(start), err)
		}
	}
	r.clock.Advance(time.Hour)
	if _, err := r.svc.Authorize(ctx, cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("after 7 days: %v", err)
	}
}

// secondLogin signs admin in again with the enrolled authenticator.
func (r *rig) secondLogin(t *testing.T) (string, Session) {
	t.Helper()
	ctx := context.Background()
	u := r.user(t, "admin")
	secret, err := r.svc.sealer.Open(u.ID, u.TOTPSecret)
	if err != nil {
		t.Fatal(err)
	}
	r.clock.Advance(time.Minute) // a fresh TOTP step
	p, err := r.svc.Login(ctx, ip, "admin", password)
	if err != nil {
		t.Fatal(err)
	}
	s, err := r.svc.SecondFactor(ctx, ip, p.Cookie, code(t, secret, r.clock.Now()))
	if err != nil {
		t.Fatal(err)
	}
	return p.Cookie, s
}

func TestTheSecondFactorWindowIsFiveMinutes(t *testing.T) {
	r := newRig(t)
	p, err := r.svc.Login(context.Background(), ip, "admin", password)
	if err != nil {
		t.Fatal(err)
	}
	r.clock.Advance(TOTPWindow)
	if _, err := r.svc.BeginEnrollment(context.Background(), p.Cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("after the window: %v", err)
	}
}

func TestAccountChangesEndSessions(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	cookie, _ := r.enrollOrLogin(t)
	if err := r.svc.ResetTOTP(ctx, "cli", "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Authorize(ctx, cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("a session survived a TOTP reset: %v", err)
	}
	if u := r.user(t, "admin"); u.TOTPEnrolled || u.TOTPSecret != nil {
		t.Fatalf("ResetTOTP left %+v", u)
	}
	cookie, _ = r.enrollOrLogin(t)
	if err := r.svc.SetPassword(ctx, "cli", "admin", "another horse battery"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Authorize(ctx, cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("a session survived a password change: %v", err)
	}
	var audit []store.AuditEntry
	if err := r.st.Tx(ctx, func(tx *store.Tx) error {
		var err error
		audit, err = tx.AuditLog(10, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(audit) != 5 || audit[0].Action != "user.passwd" || audit[2].Action != "user.reset_totp" {
		t.Fatalf("audit %+v", audit)
	}
	for _, a := range audit {
		if strings.Contains(a.Detail, "horse") {
			t.Fatalf("a password reached the audit log: %+v", a)
		}
	}
}

// enrollOrLogin returns a full session whatever the account's state.
func (r *rig) enrollOrLogin(t *testing.T) (string, Session) {
	t.Helper()
	if !r.user(t, "admin").TOTPEnrolled {
		cookie, _ := r.enroll(t)
		s, err := r.svc.Authorize(context.Background(), cookie)
		if err != nil {
			t.Fatal(err)
		}
		return cookie, s
	}
	return r.secondLogin(t)
}

func TestAccountManagementValidates(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	if _, err := r.svc.AddUser(ctx, "cli", "admin", password); !errors.Is(err, ErrUserExists) {
		t.Fatalf("duplicate user = %v", err)
	}
	for _, name := range []string{"Admin", "a b", "x;rm", ""} {
		if _, err := r.svc.AddUser(ctx, "cli", name, password); err == nil {
			t.Errorf("AddUser(%q) accepted", name)
		}
	}
	if _, err := r.svc.AddUser(ctx, "cli", "ops", "short"); err == nil {
		t.Fatal("a short password was accepted")
	}
	if err := r.svc.ResetTOTP(ctx, "cli", "ghost"); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("ResetTOTP of an unknown user = %v", err)
	}
	cli := New(Config{DB: r.st, Now: r.clock.Now, Params: cheap})
	if _, err := cli.SecondFactor(ctx, ip, "x", "123456"); !errors.Is(err, ErrNoSealer) {
		t.Fatalf("a service without a sealer = %v", err)
	}
}

func TestCookieAndCSRF(t *testing.T) {
	w := httptest.NewRecorder()
	SetCookie(w, strings.Repeat("a", 43))
	c := w.Result().Cookies()[0]
	if c.Name != "__Host-meshctl" || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode || c.Path != "/" || c.Domain != "" {
		t.Fatalf("cookie %+v", c)
	}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(c)
	if CookieValue(r) != c.Value {
		t.Fatal("CookieValue did not read the cookie")
	}
	r.Header.Set("Cookie", CookieName+"=bad;value")
	if CookieValue(r) != "" {
		t.Fatal("a malformed cookie was accepted")
	}
	w = httptest.NewRecorder()
	ClearCookie(w)
	if w.Result().Cookies()[0].MaxAge >= 0 {
		t.Fatal("ClearCookie must expire the cookie")
	}

	s := Session{CSRF: "token-123"}
	form := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{CSRFField: {"token-123"}}.Encode()))
	form.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	header := httptest.NewRequest(http.MethodPost, "/", nil)
	header.Header.Set(CSRFHeader, "token-123")
	wrong := httptest.NewRequest(http.MethodPost, "/", nil)
	wrong.Header.Set(CSRFHeader, "token-124")
	missing := httptest.NewRequest(http.MethodPost, "/", nil)
	for name, c := range map[string]struct {
		r    *http.Request
		want bool
	}{"form": {form, true}, "header": {header, true}, "wrong": {wrong, false}, "missing": {missing, false}} {
		if got := s.CheckCSRF(c.r); got != c.want {
			t.Errorf("%s: CheckCSRF = %v", name, got)
		}
	}
}

func TestLogout(t *testing.T) {
	r := newRig(t)
	cookie, _ := r.enroll(t)
	if err := r.svc.Logout(context.Background(), cookie); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Authorize(context.Background(), cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("after logout: %v", err)
	}
	if err := r.svc.Logout(context.Background(), "not a cookie"); err != nil {
		t.Fatalf("logout of garbage = %v", err)
	}
}

func TestEnrolmentReusesItsSecretAndPeekDoesNotSlide(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	p, err := r.svc.Login(ctx, ip, "admin", password)
	if err != nil {
		t.Fatal(err)
	}
	if cur, err := r.svc.Current(ctx, p.Cookie); err != nil || cur.TOTPOK || cur.Username != "admin" {
		t.Fatalf("Current of a pending session = %+v, %v", cur, err)
	}
	if _, err := r.svc.Peek(ctx, p.Cookie); !errors.Is(err, ErrNoSession) {
		t.Fatalf("Peek of a pending session = %v", err)
	}
	k1, err := r.svc.BeginEnrollment(ctx, p.Cookie)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := r.svc.BeginEnrollment(ctx, p.Cookie)
	if err != nil || k2.Secret() != k1.Secret() {
		t.Fatal("a reloaded enrolment page must show the same secret")
	}
	if _, err := r.svc.ConfirmEnrollment(ctx, ip, p.Cookie, code(t, k1.Secret(), r.clock.Now())); err != nil {
		t.Fatal(err)
	}
	before, _ := r.svc.Current(ctx, p.Cookie)
	r.clock.Advance(time.Hour)
	if _, err := r.svc.Peek(ctx, p.Cookie); err != nil {
		t.Fatal(err)
	}
	if after, _ := r.svc.Current(ctx, p.Cookie); !after.ExpiresAt.Equal(before.ExpiresAt) {
		t.Fatal("Peek extended the session")
	}
	if _, err := r.svc.Authorize(ctx, p.Cookie); err != nil {
		t.Fatal(err)
	}
	if after, _ := r.svc.Current(ctx, p.Cookie); !after.ExpiresAt.After(before.ExpiresAt) {
		t.Fatal("Authorize did not extend the session")
	}
}

func TestChangePassword(t *testing.T) {
	r := newRig(t)
	ctx := context.Background()
	cookie, s := r.enrollOrLogin(t)
	if err := r.svc.ChangePassword(ctx, s, "wrong horse battery", "another horse battery"); err == nil {
		t.Fatal("a wrong current password was accepted")
	}
	if err := r.svc.ChangePassword(ctx, s, password, "short"); err == nil {
		t.Fatal("a short new password was accepted")
	}
	if err := r.svc.ChangePassword(ctx, s, password, "another horse battery"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Authorize(ctx, cookie); !errors.Is(err, ErrNoSession) {
		t.Fatal("the changing session must end too")
	}
	if _, err := r.svc.Login(ctx, ip, "admin", "another horse battery"); err != nil {
		t.Fatalf("sign-in with the new password: %v", err)
	}
	if users, err := r.svc.Users(ctx); err != nil || len(users) != 1 {
		t.Fatalf("Users = %v, %v", users, err)
	}
}
