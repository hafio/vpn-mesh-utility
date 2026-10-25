package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

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
		{name: "add", args: "<name>", summary: "create a web UI user; they enroll an authenticator at their first sign-in", setup: setupUserAdd},
		{name: "grant", args: "<name>", summary: "let a user manage a mesh (--mesh), or make them an admin (--admin)", setup: setupUserGrant(true)},
		{name: "list", summary: "list the web UI users, the admins and the meshes granted to each", setup: setupUserList},
		{name: "passwd", args: "<name>", summary: "set a user's password and end their sessions", setup: setupUser(passwdUser)},
		{name: "reset-totp", args: "<name>", summary: "remove a user's authenticator (they enroll a new one at their next sign-in) and end their sessions",
			setup: setupUser(resetTOTP)},
		{name: "revoke", args: "<name>", summary: "take a mesh away from a user (--mesh), or their admin role (--admin)", setup: setupUserGrant(false)},
	}
}

// userAction is one user command once flags and the database are ready.
type userAction func(ctx context.Context, e *env, svc *auth.Service, name string) error

func setupUser(act userAction) func(fs *flag.FlagSet, e *env) action {
	return func(fs *flag.FlagSet, e *env) action {
		data := fs.String("data", "/data", "the server's data directory")
		return userRun(data, e, act)
	}
}

// userRun opens the database and runs act on the one positional name.
func userRun(data *string, e *env, act userAction) action {
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

// setupUserAdd creates a user (amendment A49): an admin with --admin, one
// who manages a mesh with --mesh. The first user is always an admin.
func setupUserAdd(fs *flag.FlagSet, e *env) action {
	data := fs.String("data", "/data", "the server's data directory")
	admin := fs.Bool("admin", false, "an admin manages every mesh, the meshes themselves and the users (the first user always is one)")
	mesh := fs.String("mesh", "", "the mesh the user manages, alone (not for the first user, an admin; meshctl user grant adds more)")
	run := userRun(data, e, func(ctx context.Context, e *env, svc *auth.Service, name string) error {
		pw, err := newPassword(e, name)
		if err != nil {
			return err
		}
		u, err := svc.AddUser(ctx, actorCLI, name, pw, auth.Access{Admin: *admin, Mesh: *mesh})
		if err != nil {
			return err
		}
		role := "; grant them meshes with `meshctl user grant " + name + " --mesh <mesh>`"
		switch {
		case u.Admin:
			role = " as an admin"
		case *mesh != "":
			role = ", who manages mesh " + *mesh
		}
		fmt.Fprintf(e.stdout, "added user %s%s; they enroll an authenticator app at their first sign-in\n", name, role)
		return nil
	})
	return func(ctx context.Context, args []string) error {
		if *admin && *mesh != "" {
			return usagef("an admin manages every mesh; give --admin or --mesh, not both")
		}
		return run(ctx, args)
	}
}

// setupUserGrant is user grant (on) and user revoke: one mesh, or the admin
// role.
func setupUserGrant(on bool) func(fs *flag.FlagSet, e *env) action {
	return func(fs *flag.FlagSet, e *env) action {
		data := fs.String("data", "/data", "the server's data directory")
		mesh := fs.String("mesh", "", "the mesh to grant or take away")
		admin := fs.Bool("admin", false, "the admin role instead of a mesh")
		run := userRun(data, e, func(ctx context.Context, e *env, svc *auth.Service, name string) error {
			if *admin {
				if err := svc.SetAdmin(ctx, actorCLI, name, on); err != nil {
					return err
				}
				fmt.Fprintf(e.stdout, "%s %s an admin\n", name, map[bool]string{true: "is now", false: "is no longer"}[on])
				return nil
			}
			if on {
				if err := svc.Grant(ctx, actorCLI, name, *mesh); err != nil {
					return err
				}
				fmt.Fprintf(e.stdout, "%s manages mesh %s\n", name, *mesh)
				return nil
			}
			if err := svc.Revoke(ctx, actorCLI, name, *mesh); err != nil {
				return err
			}
			fmt.Fprintf(e.stdout, "%s no longer manages mesh %s\n", name, *mesh)
			return nil
		})
		return func(ctx context.Context, args []string) error {
			if (*mesh == "") == !*admin {
				return usagef("give --mesh <mesh> or --admin, not both")
			}
			return run(ctx, args)
		}
	}
}

func setupUserList(fs *flag.FlagSet, e *env) action {
	data := fs.String("data", "/data", "the server's data directory")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		st, err := store.Open(ctx, filepath.Join(*data, server.DBFile), store.Options{Now: e.now, ReadOnly: true})
		if err != nil {
			return err
		}
		defer st.Close()
		svc := auth.New(auth.Config{DB: st, Now: e.now})
		users, err := svc.Users(ctx)
		if err != nil {
			return err
		}
		grants, err := svc.Grants(ctx)
		if err != nil {
			return err
		}
		tw := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tMANAGES\tAUTHENTICATOR\tSTATE")
		for _, u := range users {
			manages := cmp.Or(strings.Join(grants[u.ID], ", "), "-")
			if u.Admin {
				manages = "everything (admin)"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", u.Username, manages, map[bool]string{true: "enrolled", false: "at next sign-in"}[u.TOTPEnrolled],
				map[bool]string{true: "disabled", false: "active"}[u.Disabled])
		}
		return tw.Flush()
	}
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
