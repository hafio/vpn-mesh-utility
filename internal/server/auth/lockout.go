package auth

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Login throttling (spec S8.4): 5 failures from one address within 15
// minutes lock that address out for 15 minutes; 20 failures for one username
// within an hour lock the username until the oldest of them is an hour old,
// whichever addresses they came from.
const (
	ipFailures   = 5
	ipWindow     = 15 * time.Minute
	ipLockout    = 15 * time.Minute
	userFailures = 20
	userWindow   = time.Hour
	// maxTracked bounds the lockout's memory; quiet entries are dropped
	// first.
	maxTracked = 10000
)

// LockoutError says how long a login is refused.
type LockoutError struct {
	Until time.Time
}

func (e *LockoutError) Error() string {
	return fmt.Sprintf("too many failed sign-ins; try again after %s", e.Until.Format(time.RFC3339))
}

type failures struct {
	at    []time.Time
	until time.Time
}

// Lockout counts failed sign-ins per address and per username.
type Lockout struct {
	mu    sync.Mutex
	ips   map[string]*failures
	users map[string]*failures
}

// NewLockout returns an empty lockout.
func NewLockout() *Lockout {
	return &Lockout{ips: map[string]*failures{}, users: map[string]*failures{}}
}

// Check reports a *LockoutError when ip or user is locked out at now.
func (l *Lockout) Check(ip, user string, now time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	until := time.Time{}
	for _, f := range []*failures{l.ips[ip], l.users[strings.ToLower(user)]} {
		if f != nil && now.Before(f.until) && f.until.After(until) {
			until = f.until
		}
	}
	if !until.IsZero() {
		return &LockoutError{Until: until}
	}
	return nil
}

// Fail records a failed sign-in.
func (l *Lockout) Fail(ip, user string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.ips)+len(l.users) >= maxTracked {
		l.sweep(now)
	}
	f := record(l.ips, ip, now, ipWindow)
	if len(f.at) >= ipFailures {
		f.until = now.Add(ipLockout)
	}
	u := record(l.users, strings.ToLower(user), now, userWindow)
	if len(u.at) >= userFailures {
		u.until = u.at[0].Add(userWindow)
	}
}

// Succeed clears the counters of a successful sign-in.
func (l *Lockout) Succeed(ip, user string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.ips, ip)
	delete(l.users, strings.ToLower(user))
}

func record(m map[string]*failures, key string, now time.Time, window time.Duration) *failures {
	f := m[key]
	if f == nil {
		f = &failures{}
		m[key] = f
	}
	kept := f.at[:0]
	for _, t := range f.at {
		if now.Sub(t) < window {
			kept = append(kept, t)
		}
	}
	f.at = append(kept, now)
	return f
}

// sweep drops entries with no recent failure and no live lockout.
func (l *Lockout) sweep(now time.Time) {
	for _, m := range []map[string]*failures{l.ips, l.users} {
		for k, f := range m {
			if now.After(f.until) && (len(f.at) == 0 || now.Sub(f.at[len(f.at)-1]) >= userWindow) {
				delete(m, k)
			}
		}
	}
}
