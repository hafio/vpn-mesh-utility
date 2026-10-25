package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// User is one row of users: a web UI account (spec S8.4).
type User struct {
	ID       int64
	Username string
	// PasswordHash is an argon2id PHC string (internal/server/auth).
	PasswordHash string
	// TOTPSecret is the sealed TOTP secret; nil until enrolment starts.
	TOTPSecret   []byte
	TOTPEnrolled bool
	// TOTPLastStep is the last accepted TOTP time step; a code for it or an
	// earlier step is a replay.
	TOTPLastStep int64
	Disabled     bool
	CreatedAt    int64
	// Admin manages the control plane: every mesh, the meshes and the
	// users (amendment A49); any other user only the meshes granted to them.
	Admin bool
}

// ErrTOTPReplay means a TOTP step at or before the last accepted one.
var ErrTOTPReplay = errors.New("this TOTP code was already used; wait for the next one")

const userCols = `id, username, password_hash, totp_secret, totp_enrolled, totp_last_step, disabled, created_at, is_admin`

func scanUser(r scanner) (User, error) {
	var (
		u                         User
		enrolled, disabled, admin int
	)
	if err := r.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.TOTPSecret, &enrolled, &u.TOTPLastStep, &disabled, &u.CreatedAt, &admin); err != nil {
		return User{}, err
	}
	u.TOTPEnrolled, u.Disabled, u.Admin = enrolled == 1, disabled == 1, admin == 1
	return u, nil
}

// InsertUser creates a user and sets u.ID and u.CreatedAt. A taken username
// is ErrConflict.
func (t *Tx) InsertUser(u *User) error {
	u.CreatedAt = t.now.Unix()
	res, err := t.exec(`INSERT INTO users (username, password_hash, totp_secret, totp_enrolled, totp_last_step, disabled, created_at, is_admin)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.TOTPSecret, boolInt(u.TOTPEnrolled), u.TOTPLastStep, boolInt(u.Disabled), u.CreatedAt, boolInt(u.Admin))
	if err != nil {
		return fmt.Errorf("insert user %q: %w", u.Username, err)
	}
	if u.ID, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("insert user %q: %w", u.Username, err)
	}
	return nil
}

// UserByName returns one user or ErrNotFound.
func (t *Tx) UserByName(name string) (User, error) {
	u, err := scanUser(t.queryRow(`SELECT `+userCols+` FROM users WHERE username = ?`, name))
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, fmt.Errorf("user %q: %w", name, ErrNotFound)
	}
	if err != nil {
		return User{}, fmt.Errorf("read user %q: %w", name, err)
	}
	return u, nil
}

// UserByID returns one user or ErrNotFound.
func (t *Tx) UserByID(id int64) (User, error) {
	u, err := scanUser(t.queryRow(`SELECT `+userCols+` FROM users WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, fmt.Errorf("user %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return User{}, fmt.Errorf("read user %d: %w", id, err)
	}
	return u, nil
}

// Users lists every user by name.
func (t *Tx) Users() ([]User, error) {
	rows, err := t.query(`SELECT ` + userCols + ` FROM users ORDER BY username`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return out, nil
}

// SetPassword replaces a user's password hash.
func (t *Tx) SetPassword(id int64, hash string) error {
	res, err := t.exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hash, id)
	if err != nil {
		return fmt.Errorf("set password of user %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("user %d", id))
}

// SetTOTP stores a sealed secret (nil clears it) and the enrolment flag. The
// replay counter starts over with a new secret.
func (t *Tx) SetTOTP(id int64, sealed []byte, enrolled bool) error {
	res, err := t.exec(`UPDATE users SET totp_secret = ?, totp_enrolled = ?, totp_last_step = 0 WHERE id = ?`,
		sealed, boolInt(enrolled), id)
	if err != nil {
		return fmt.Errorf("set TOTP of user %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("user %d", id))
}

// UseTOTPStep records step as used. The write only happens when step is newer
// than the last accepted one, so two requests racing with the same code
// cannot both succeed: the loser gets ErrTOTPReplay.
func (t *Tx) UseTOTPStep(id, step int64) error {
	res, err := t.exec(`UPDATE users SET totp_last_step = ? WHERE id = ? AND totp_last_step < ?`, step, id, step)
	if err != nil {
		return fmt.Errorf("record TOTP step of user %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("record TOTP step of user %d: %w", id, err)
	}
	if n == 0 {
		return ErrTOTPReplay
	}
	return nil
}

// SetUserDisabled disables or enables a user.
func (t *Tx) SetUserDisabled(id int64, disabled bool) error {
	res, err := t.exec(`UPDATE users SET disabled = ? WHERE id = ?`, boolInt(disabled), id)
	if err != nil {
		return fmt.Errorf("disable user %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("user %d", id))
}

// SetUserAdmin makes a user an admin or takes it away.
func (t *Tx) SetUserAdmin(id int64, admin bool) error {
	res, err := t.exec(`UPDATE users SET is_admin = ? WHERE id = ?`, boolInt(admin), id)
	if err != nil {
		return fmt.Errorf("set admin of user %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("user %d", id))
}

// Admins counts the enabled admins.
func (t *Tx) Admins() (int, error) {
	var n int
	if err := t.queryRow(`SELECT COUNT(*) FROM users WHERE is_admin = 1 AND disabled = 0`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return n, nil
}

// GrantMesh lets a user manage a mesh and reports whether that changed
// anything: granting it twice is no change.
func (t *Tx) GrantMesh(userID, meshID int64) (bool, error) {
	res, err := t.exec(`INSERT INTO user_meshes (user_id, mesh_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, userID, meshID)
	if err != nil {
		return false, fmt.Errorf("grant mesh %d to user %d: %w", meshID, userID, err)
	}
	return changed(res, fmt.Sprintf("grant mesh %d to user %d", meshID, userID))
}

// RevokeMesh takes a mesh away from a user and reports whether that changed
// anything: one not granted is no change.
func (t *Tx) RevokeMesh(userID, meshID int64) (bool, error) {
	res, err := t.exec(`DELETE FROM user_meshes WHERE user_id = ? AND mesh_id = ?`, userID, meshID)
	if err != nil {
		return false, fmt.Errorf("revoke mesh %d from user %d: %w", meshID, userID, err)
	}
	return changed(res, fmt.Sprintf("revoke mesh %d from user %d", meshID, userID))
}

func changed(res sql.Result, what string) (bool, error) {
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", what, err)
	}
	return n > 0, nil
}

// UserMeshes are the ids of the meshes granted to a user, in order.
func (t *Tx) UserMeshes(userID int64) ([]int64, error) {
	rows, err := t.query(`SELECT mesh_id FROM user_meshes WHERE user_id = ? ORDER BY mesh_id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list the meshes of user %d: %w", userID, err)
	}
	var out []int64
	err = scanAll(rows, "user meshes", func(r *sql.Rows) error {
		var id int64
		if err := r.Scan(&id); err != nil {
			return err
		}
		out = append(out, id)
		return nil
	})
	return out, err
}

// Grants are the names of the meshes granted to each user, by user id, in
// mesh order.
func (t *Tx) Grants() (map[int64][]string, error) {
	rows, err := t.query(`SELECT um.user_id, m.name FROM user_meshes um JOIN meshes m ON m.id = um.mesh_id ORDER BY m.id`)
	if err != nil {
		return nil, fmt.Errorf("list grants: %w", err)
	}
	out := map[int64][]string{}
	err = scanAll(rows, "grants", func(r *sql.Rows) error {
		var (
			id   int64
			name string
		)
		if err := r.Scan(&id, &name); err != nil {
			return err
		}
		out[id] = append(out[id], name)
		return nil
	})
	return out, err
}

// Session is one row of sessions. IDHash is sha256 of the cookie value, so a
// database read cannot hijack a session.
type Session struct {
	IDHash    string
	UserID    int64
	CreatedAt int64
	LastSeen  int64
	ExpiresAt int64
	CSRF      string
	// TOTPOK is false between the password and the second factor.
	TOTPOK bool
}

// InsertSession creates a session.
func (t *Tx) InsertSession(s Session) error {
	if _, err := t.exec(`INSERT INTO sessions (id_hash, user_id, created_at, last_seen, expires_at, csrf, totp_ok)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.IDHash, s.UserID, s.CreatedAt, s.LastSeen, s.ExpiresAt, s.CSRF, boolInt(s.TOTPOK)); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// SessionByHash returns one session or ErrNotFound.
func (t *Tx) SessionByHash(h string) (Session, error) {
	var (
		s  Session
		ok int
	)
	err := t.queryRow(`SELECT id_hash, user_id, created_at, last_seen, expires_at, csrf, totp_ok FROM sessions WHERE id_hash = ?`, h).
		Scan(&s.IDHash, &s.UserID, &s.CreatedAt, &s.LastSeen, &s.ExpiresAt, &s.CSRF, &ok)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, fmt.Errorf("session: %w", ErrNotFound)
	}
	if err != nil {
		return Session{}, fmt.Errorf("read session: %w", err)
	}
	s.TOTPOK = ok == 1
	return s, nil
}

// UpdateSession writes a session's activity, expiry and second-factor state.
func (t *Tx) UpdateSession(s Session) error {
	res, err := t.exec(`UPDATE sessions SET last_seen = ?, expires_at = ?, totp_ok = ? WHERE id_hash = ?`,
		s.LastSeen, s.ExpiresAt, boolInt(s.TOTPOK), s.IDHash)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return mustAffect(res, "session")
}

// DeleteSession ends one session; an absent one is not an error.
func (t *Tx) DeleteSession(h string) error {
	if _, err := t.exec(`DELETE FROM sessions WHERE id_hash = ?`, h); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteUserSessions ends every session of a user (password change, TOTP
// reset, disable).
func (t *Tx) DeleteUserSessions(userID int64) error {
	if _, err := t.exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("delete sessions of user %d: %w", userID, err)
	}
	return nil
}

// PurgeSessions deletes sessions that expired before now and reports how
// many.
func (t *Tx) PurgeSessions(now int64) (int64, error) {
	res, err := t.exec(`DELETE FROM sessions WHERE expires_at <= ?`, now)
	if err != nil {
		return 0, fmt.Errorf("purge sessions: %w", err)
	}
	return res.RowsAffected()
}
