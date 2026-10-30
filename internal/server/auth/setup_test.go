package auth

import (
	"context"
	"errors"
	"path/filepath"
	"regexp"
	"slices"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// emptyService is a control plane without any user.
func emptyService(t *testing.T) (*Service, *store.Store) {
	t.Helper()
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "mesh.db"), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(Config{DB: st, Now: clock.Now, Params: cheap}), st
}

// The first admin from the web UI (amendment A52): only with the latest
// setup code, only while no admin exists; wrong codes are throttled per
// address and nothing is stored for them.
func TestFirstAdmin(t *testing.T) {
	ctx := context.Background()
	svc, st := emptyService(t)
	count := func() int {
		us, err := svc.Users(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return len(us)
	}

	if open, err := svc.SetupOpen(ctx); err != nil || !open {
		t.Fatalf("SetupOpen on an empty control plane = %v, %v", open, err)
	}
	if _, err := svc.FirstAdmin(ctx, ip, "abcd-efgh-ijkl-mnop-qrst-uvwx-yz23-4567", "boss", password); !errors.Is(err, ErrSetupCode) {
		t.Fatalf("FirstAdmin before any code = %v", err)
	}
	old, err := svc.SetupCode(ctx)
	if err != nil {
		t.Fatal(err)
	}
	code, err := svc.SetupCode(ctx)
	if err != nil || code == old || !regexp.MustCompile(`^([a-z2-7]{4}-){7}[a-z2-7]{4}$`).MatchString(code) {
		t.Fatalf("SetupCode = %q (before %q), %v", code, old, err)
	}

	for _, c := range []struct {
		name, code, user, pw string
		wrongCode            bool // else a *validate.Error
	}{
		{"a code from an earlier start", old, "boss", password, true},
		{"not a code", "nope", "boss", password, false},
		{"a bad username", code, "Boss!", password, false},
		{"a short password", code, "boss", "short", false},
	} {
		_, err := svc.FirstAdmin(ctx, ip, c.code, c.user, c.pw)
		var ve *validate.Error
		if c.wrongCode && !errors.Is(err, ErrSetupCode) || !c.wrongCode && !errors.As(err, &ve) {
			t.Errorf("%s: FirstAdmin = %v", c.name, err)
		}
		if n := count(); n != 0 {
			t.Fatalf("%s: %d users stored", c.name, n)
		}
	}

	// Five wrong codes lock that address out, the right code included;
	// another address is not.
	const other = "198.51.100.9"
	for i := 1; i <= 5; i++ {
		if _, err := svc.FirstAdmin(ctx, other, old, "boss", password); !errors.Is(err, ErrSetupCode) {
			t.Fatalf("wrong code %d = %v", i, err)
		}
	}
	var le *LockoutError
	if _, err := svc.FirstAdmin(ctx, other, code, "boss", password); !errors.As(err, &le) {
		t.Fatalf("the right code from a locked address = %v", err)
	}

	p, err := svc.FirstAdmin(ctx, ip, " "+code+" ", "boss", password)
	if err != nil || !p.NeedsEnrollment || p.Session.TOTPOK || p.Session.Username != "boss" || !p.Session.Admin || p.Cookie == "" {
		t.Fatalf("FirstAdmin = %+v, %v", p, err)
	}
	var log []store.AuditEntry
	if err := st.Tx(ctx, func(tx *store.Tx) error {
		var err error
		log, err = tx.AuditLog(10, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(log, func(e store.AuditEntry) bool {
		return e.Actor == ActorSetup && e.Action == "user.add" && e.Target == "boss"
	}) {
		t.Fatalf("audit = %+v", log)
	}

	// Done: the code is gone, the page is closed, no code is made.
	if _, err := svc.FirstAdmin(ctx, ip, code, "second", password); !errors.Is(err, ErrSetupCode) || count() != 1 {
		t.Fatalf("the code again = %v, %d users", err, count())
	}
	if open, err := svc.SetupOpen(ctx); err != nil || open {
		t.Fatalf("SetupOpen with an admin = %v, %v", open, err)
	}
	if c, err := svc.SetupCode(ctx); err != nil || c != "" {
		t.Fatalf("SetupCode with an admin = %q, %v", c, err)
	}
}

// An admin made with `meshctl user add` while the setup page is open closes
// it: the logged code makes no second admin.
func TestFirstAdminAfterTheCLI(t *testing.T) {
	ctx := context.Background()
	svc, _ := emptyService(t)
	code, err := svc.SetupCode(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddUser(ctx, "cli", "admin", password, Access{}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.FirstAdmin(ctx, ip, code, "boss", password); !errors.Is(err, ErrSetupDone) {
		t.Fatalf("FirstAdmin after the CLI = %v", err)
	}
	if us, err := svc.Users(ctx); err != nil || len(us) != 1 {
		t.Fatalf("users = %+v, %v", us, err)
	}
}
