package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// `meshctl exit` shows or switches a spoke's internet exit (spec S7.3). The
// local agent relays the change to the server, which owns every AllowedIP
// (amendment A10); the command then waits until the agent has applied the
// config that carries it.
func setupExit(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	wait := fs.Duration("wait", 30*time.Second, "how long to wait for the agent to apply the change (0 returns once the server accepted it)")
	return func(ctx context.Context, args []string) error {
		c := localapi.NewClient(sock())
		if len(args) == 0 {
			st, err := c.Status(ctx)
			if err != nil {
				return err
			}
			printExit(e.stdout, st)
			return nil
		}
		if err := wantArgs(args, 1, "<server> or off"); err != nil {
			return err
		}
		via, what := args[0], "exit via "+args[0]
		if via == "off" {
			via, what = "", "exit off"
		} else if _, err := validate.Name("exit", via); err != nil {
			return usagef("%v", err)
		}
		resp, err := c.Prefs(ctx, proto.PrefsRequest{Exit: &via})
		if err != nil {
			return err
		}
		if *wait <= 0 {
			fmt.Fprintf(e.stdout, "%s: the server made config v%d; the agent applies it in a moment\n", what, resp.ConfigVersion)
			return nil
		}
		if err := waitApplied(ctx, c, resp.ConfigHash, *wait); err != nil {
			return err
		}
		fmt.Fprintf(e.stdout, "%s: config v%d applied\n", what, resp.ConfigVersion)
		return nil
	}
}

// waitApplied polls the agent until the config with hash is on the
// interface.
func waitApplied(ctx context.Context, c *localapi.Client, hash string, within time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, within)
	defer cancel()
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	var last proto.LocalStatus
	for {
		if st, err := c.Status(ctx); err == nil {
			if st.AppliedHash == hash {
				return nil
			}
			last = st
		}
		select {
		case <-ctx.Done():
			why := "it is still working on it"
			if len(last.Errors) > 0 {
				why = last.Errors[0]
			}
			return fmt.Errorf("the server accepted the change, but the agent has not applied it within %s (%s); see `meshctl status`", within, why)
		case <-tick.C:
		}
	}
}

func printExit(w io.Writer, st proto.LocalStatus) {
	if st.Exit != "" {
		fmt.Fprintf(w, "exit via %s\n", st.Exit)
	} else {
		fmt.Fprintln(w, "exit off")
	}
	if len(st.ExitChoices) == 0 {
		fmt.Fprintln(w, "no server of this mesh can be an exit; an operator allows one with `meshctl node set <server> --can-exit`")
		return
	}
	fmt.Fprintf(w, "choices: %s\n", strings.Join(st.ExitChoices, ", "))
}
