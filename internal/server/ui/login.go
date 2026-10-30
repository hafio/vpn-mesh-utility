package ui

import (
	"errors"
	"image/png"
	"net/http"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Sign-in (spec S8.4): password, then the authenticator code -- or, at the
// first sign-in, enrolment of an authenticator app by QR code.

type loginData struct {
	Username string
	// Setup links the setup page while the control plane has no admin.
	Setup bool
}

func (u *ui) loginPage(w http.ResponseWriter, r *http.Request) {
	if s, err := u.Auth.Current(r.Context(), auth.CookieValue(r)); err == nil && s.TOTPOK {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	u.show(w, r, auth.Session{}, http.StatusOK, "login", "", "Sign in", page{Data: loginData{Setup: u.setupOpen(r)}})
}

func (u *ui) login(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PostFormValue("username"))
	p, err := u.Auth.Login(r.Context(), u.clientIP(r), name, r.PostFormValue("password"))
	if err != nil {
		status, msg := message(err)
		if status >= http.StatusInternalServerError {
			u.Log.Error("sign-in failed", "err", err)
		}
		u.show(w, r, auth.Session{}, status, "login", "", "Sign in", page{Error: msg, Data: loginData{Username: name, Setup: u.setupOpen(r)}})
		return
	}
	auth.SetCookie(w, p.Cookie)
	next := "/login/totp"
	if p.NeedsEnrollment {
		next = "/login/enroll"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// The first admin (amendment A52): while the control plane has no admin, the
// setup page takes the setup code its log shows, a username and a password,
// then goes on to the authenticator's enrolment like a first sign-in.

type setupData struct {
	Username string
}

// setupOpen reports whether the setup page is open; a failed check closes it.
func (u *ui) setupOpen(r *http.Request) bool {
	open, err := u.Auth.SetupOpen(r.Context())
	if err != nil {
		u.Log.Error("check for an admin", "err", err)
	}
	return open
}

func (u *ui) setupPage(w http.ResponseWriter, r *http.Request) {
	if !u.setupOpen(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	u.show(w, r, auth.Session{}, http.StatusOK, "setup", "", "Set up the first admin", page{Data: setupData{}})
}

func (u *ui) setup(w http.ResponseWriter, r *http.Request) {
	if !u.setupOpen(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	name := strings.TrimSpace(r.PostFormValue("username"))
	var (
		p   auth.Pending
		err error
	)
	if pw := r.PostFormValue("password"); pw != r.PostFormValue("password2") {
		err = &validate.Error{Field: "password", Reason: "the two passwords differ"}
	} else {
		p, err = u.Auth.FirstAdmin(r.Context(), u.clientIP(r), r.PostFormValue("code"), name, pw)
	}
	if errors.Is(err, auth.ErrSetupDone) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err != nil {
		status, msg := message(err)
		if status >= http.StatusInternalServerError {
			u.Log.Error("first admin failed", "err", err)
		}
		u.show(w, r, auth.Session{}, status, "setup", "", "Set up the first admin", page{Error: msg, Data: setupData{Username: name}})
		return
	}
	u.Log.Info("first admin created on the setup page", "user", p.Session.Username, "ip", u.clientIP(r))
	auth.SetCookie(w, p.Cookie)
	http.Redirect(w, r, "/login/enroll", http.StatusSeeOther)
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
