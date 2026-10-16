package ui

import (
	"image/png"
	"net/http"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
)

// Sign-in (spec S8.4): password, then the authenticator code -- or, at the
// first sign-in, enrolment of an authenticator app by QR code.

type loginData struct {
	Username string
}

func (u *ui) loginPage(w http.ResponseWriter, r *http.Request) {
	if s, err := u.Auth.Current(r.Context(), auth.CookieValue(r)); err == nil && s.TOTPOK {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u.show(w, r, auth.Session{}, http.StatusOK, "login", "", "Sign in", page{Data: loginData{}})
}

func (u *ui) login(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PostFormValue("username"))
	p, err := u.Auth.Login(r.Context(), u.clientIP(r), name, r.PostFormValue("password"))
	if err != nil {
		status, msg := message(err)
		if status >= http.StatusInternalServerError {
			u.Log.Error("sign-in failed", "err", err)
		}
		u.show(w, r, auth.Session{}, status, "login", "", "Sign in", page{Error: msg, Data: loginData{Username: name}})
		return
	}
	auth.SetCookie(w, p.Cookie)
	next := "/login/totp"
	if p.NeedsEnrollment {
		next = "/login/enroll"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (u *ui) totpPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.show(w, r, auth.Session{CSRF: s.CSRF}, http.StatusOK, "totp", "", "Authenticator code", page{})
}

func (u *ui) totp(w http.ResponseWriter, r *http.Request, s auth.Session) {
	if _, err := u.Auth.SecondFactor(r.Context(), u.clientIP(r), auth.CookieValue(r), strings.TrimSpace(r.PostFormValue("code"))); err != nil {
		status, msg := message(err)
		u.show(w, r, auth.Session{CSRF: s.CSRF}, status, "totp", "", "Authenticator code", page{Error: msg})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type enrollData struct {
	Account string
	// Secret is shown in groups of four for typing into an app by hand.
	Secret string
}

func (u *ui) enrollPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	key, err := u.Auth.BeginEnrollment(r.Context(), auth.CookieValue(r))
	if err != nil {
		u.fail(w, r, auth.Session{}, err)
		return
	}
	u.show(w, r, auth.Session{CSRF: s.CSRF}, http.StatusOK, "enroll", "", "Set up your authenticator",
		page{Data: enrollData{Account: s.Username, Secret: group(key.Secret(), 4)}})
}

// enrollQR serves the enrolment QR code as a PNG: an image from this origin,
// which the CSP allows, rather than a data: URL, which it does not.
func (u *ui) enrollQR(w http.ResponseWriter, r *http.Request, _ auth.Session) {
	key, err := u.Auth.BeginEnrollment(r.Context(), auth.CookieValue(r))
	if err != nil {
		http.Error(w, "no enrolment in progress", http.StatusNotFound)
		return
	}
	img, err := key.Image(256, 256)
	if err != nil {
		u.Log.Error("render enrolment QR code", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		u.Log.Warn("send enrolment QR code", "err", err)
	}
}

func (u *ui) enroll(w http.ResponseWriter, r *http.Request, s auth.Session) {
	if _, err := u.Auth.ConfirmEnrollment(r.Context(), u.clientIP(r), auth.CookieValue(r), strings.TrimSpace(r.PostFormValue("code"))); err != nil {
		status, msg := message(err)
		key, kerr := u.Auth.BeginEnrollment(r.Context(), auth.CookieValue(r))
		data := enrollData{Account: s.Username}
		if kerr == nil {
			data.Secret = group(key.Secret(), 4)
		}
		u.show(w, r, auth.Session{CSRF: s.CSRF}, status, "enroll", "", "Set up your authenticator", page{Error: msg, Data: data})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (u *ui) logout(w http.ResponseWriter, r *http.Request, _ auth.Session) {
	if err := u.Auth.Logout(r.Context(), auth.CookieValue(r)); err != nil {
		u.Log.Warn("sign out", "err", err)
	}
	auth.ClearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// group splits s into space-separated runs of n characters.
func group(s string, n int) string {
	var b strings.Builder
	for i := 0; i < len(s); i += n {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s[i:min(i+n, len(s))])
	}
	return b.String()
}
