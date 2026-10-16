package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"
)

// Session lifetimes (spec S8.4).
const (
	// CookieName uses the __Host- prefix: the browser only accepts it over
	// https, for this host only, with Path=/.
	CookieName      = "__Host-meshctl"
	IdleTimeout     = 12 * time.Hour
	AbsoluteTimeout = 7 * 24 * time.Hour
	// TOTPWindow is how long a session that passed the password may wait for
	// its second factor.
	TOTPWindow = 5 * time.Minute
)

// CSRFField and CSRFHeader carry the session's CSRF token on every request
// that changes something: a hidden form field, or the header htmx sends.
const (
	CSRFField  = "csrf"
	CSRFHeader = "X-CSRF-Token"
)

var reCookie = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// Session is a signed-in (or half signed-in) browser.
type Session struct {
	UserID   int64
	Username string
	CSRF     string
	// TOTPOK is false until the second factor passed.
	TOTPOK    bool
	ExpiresAt time.Time
}

// CheckCSRF reports whether r carries the session's CSRF token, compared in
// constant time.
func (s Session) CheckCSRF(r *http.Request) bool {
	got := r.Header.Get(CSRFHeader)
	if got == "" {
		got = r.PostFormValue(CSRFField)
	}
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(s.CSRF)) == 1
}

// newSecret returns 32 random bytes, base64url: a cookie value or a CSRF
// token.
func newSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read never fails (Go 1.24+)
	return base64.RawURLEncoding.EncodeToString(b)
}

// hashCookie is how a session is stored: a database read cannot hijack it.
func hashCookie(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

// SetCookie hands the browser its session cookie.
func SetCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: value, Path: "/", HttpOnly: true, Secure: true,
		SameSite: http.SameSiteStrictMode, MaxAge: int(AbsoluteTimeout / time.Second)})
}

// ClearCookie removes the session cookie.
func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", HttpOnly: true, Secure: true,
		SameSite: http.SameSiteStrictMode, MaxAge: -1})
}

// CookieValue returns the request's session cookie, or "".
func CookieValue(r *http.Request) string {
	c, err := r.Cookie(CookieName)
	if err != nil || !reCookie.MatchString(c.Value) {
		return ""
	}
	return c.Value
}
