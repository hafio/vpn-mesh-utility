package auth

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestLockoutPerAddress(t *testing.T) {
	l := NewLockout()
	now := testfix.Epoch
	for i := 0; i < ipFailures; i++ {
		if err := l.Check("192.0.2.1", "admin", now); err != nil {
			t.Fatalf("attempt %d refused: %v", i+1, err)
		}
		l.Fail("192.0.2.1", "admin", now)
	}
	var le *LockoutError
	if err := l.Check("192.0.2.1", "someone-else", now); !errors.As(err, &le) || !le.Until.Equal(now.Add(ipLockout)) {
		t.Fatalf("after %d failures = %v", ipFailures, err)
	}
	if err := l.Check("192.0.2.2", "admin", now); err != nil {
		t.Fatalf("another address must not be locked out: %v", err)
	}
	if err := l.Check("192.0.2.1", "admin", now.Add(ipLockout)); err != nil {
		t.Fatalf("the lockout must end: %v", err)
	}
}

func TestFailuresOutsideTheWindowAreForgotten(t *testing.T) {
	l := NewLockout()
	now := testfix.Epoch
	for i := 0; i < ipFailures-1; i++ {
		l.Fail("192.0.2.1", "admin", now)
	}
	l.Fail("192.0.2.1", "admin", now.Add(ipWindow))
	if err := l.Check("192.0.2.1", "admin", now.Add(ipWindow)); err != nil {
		t.Fatalf("old failures were counted: %v", err)
	}
	l.Succeed("192.0.2.1", "admin")
}

func TestLockoutPerUsernameAcrossAddresses(t *testing.T) {
	l := NewLockout()
	now := testfix.Epoch
	for i := 0; i < userFailures; i++ {
		l.Fail(fmt.Sprintf("198.51.100.%d", i), "Admin", now.Add(time.Duration(i)*time.Second))
	}
	if err := l.Check("203.0.113.1", "admin", now.Add(time.Minute)); err == nil {
		t.Fatal("20 failures for one username from many addresses must lock it")
	}
	if err := l.Check("203.0.113.1", "admin", now.Add(userWindow)); err != nil {
		t.Fatalf("the username lockout must end an hour after the oldest failure: %v", err)
	}
	l.Succeed("203.0.113.1", "ADMIN")
	if err := l.Check("203.0.113.1", "admin", now.Add(time.Minute)); err != nil {
		t.Fatalf("a successful sign-in clears the username: %v", err)
	}
}

func TestLockoutStaysBounded(t *testing.T) {
	l := NewLockout()
	now := testfix.Epoch
	for i := 0; i < maxTracked+10; i++ {
		l.Fail(fmt.Sprintf("10.%d.%d.%d", i>>16&255, i>>8&255, i&255), "u", now.Add(-2*userWindow))
	}
	l.Fail("192.0.2.1", "u", now)
	if n := len(l.ips) + len(l.users); n > maxTracked+2 {
		t.Fatalf("lockout tracks %d entries", n)
	}
}
