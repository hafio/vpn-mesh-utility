package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"github.com/hafio/vpn-mesh-utility/internal/server"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// The user commands open the database directly (amendment A9), so they work
// with the server stopped -- the recovery path when nobody can sign in. In
// the container: `docker exec -it <container> /app/meshctl user add admin`.

func userCommands() []command {
	return []command{
		{name: "add", args: "<name>", summary: "create a web UI user; they enroll an authenticator at their first sign-in", setup: setupUser(addUser)},
		{name: "passwd", args: "<name>", summary: "set a user's password and end their sessions", setup: setupUser(passwdUser)},
		{name: "reset-totp", args: "<name>", summary: "remove a user's authenticator (they enroll a new one at their next sign-in) and end their sessions",
			setup: setupUser(resetTOTP)},
	}
}

// userAction is one user command once flags and the database are ready.
type userAction func(ctx context.Context, e *env, svc *auth.Service, name string) error

func setupUser(act userAction) func(fs *flag.FlagSet, e *env) action {
	return func(fs *flag.FlagSet, e *env) action {
		data := fs.String("data", "/data", "the server's data directory")
		return func(ctx context.Context, args []string) error {
			if err := wantArgs(args, 1, "<name>"); err != nil {
				return err
			}
			st, err := store.Open(ctx, filepath.Join(*data, server.DBFile), store.Options{Now: e.now})
			if err != nil {
				return err
			}
			defer st.Close()
			return act(ctx, e, auth.New(auth.Config{DB: st, Now: e.now}), args[0])
		}
	}
}

func addUser(ctx context.Context, e *env, svc *auth.Service, name string) error {
	pw, err := newPassword(e, name)
	if err != nil {
		return err
	}
	if _, err := svc.AddUser(ctx, actorCLI, name, pw); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "added user %s; they enroll an authenticator app at their first sign-in\n", name)
	return nil
}

func passwdUser(ctx context.Context, e *env, svc *auth.Service, name string) error {
	pw, err := newPassword(e, name)
	if err != nil {
		return err
	}
	if err := svc.SetPassword(ctx, actorCLI, name, pw); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "changed the password of %s and ended their sessions\n", name)
	return nil
}

func resetTOTP(ctx context.Context, e *env, svc *auth.Service, name string) error {
	if err := svc.ResetTOTP(ctx, actorCLI, name); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "removed the authenticator of %s; they enroll a new one at their next sign-in\n", name)
	return nil
}

// actorCLI is the audit actor of commands run on the server host.
const actorCLI = "cli"

func newPassword(e *env, name string) (string, error) {
	pw, err := e.password(fmt.Sprintf("new password for %s: ", name))
	if err != nil {
		return "", err
	}
	if err := auth.CheckPassword(pw); err != nil {
		return "", err
	}
	again, err := e.password("repeat it: ")
	if err != nil {
		return "", err
	}
	if again != pw {
		return "", errors.New("the two passwords differ; nothing was changed")
	}
	return pw, nil
}

// ttyPassword reads a password from the terminal without echo. Anything but
// a terminal is refused, so a password never comes from a pipe, a file or a
// command line (shell history, process lists).
func ttyPassword(in *os.File, prompt io.Writer) func(string) (string, error) {
	return func(msg string) (string, error) {
		fd := int(in.Fd())
		if !term.IsTerminal(fd) {
			return "", errors.New("passwords are read from a terminal only; run the command interactively (docker exec -it ...)")
		}
		fmt.Fprint(prompt, msg)
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(prompt)
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(b), nil
	}
}
