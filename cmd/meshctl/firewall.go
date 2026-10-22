package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The firewall commands show and change the mesh firewall (amendment A43)
// on the running server, over its admin socket: groups, objects, zones and
// the ordered policies every server enforces. The web UI's Firewall page
// edits the same document one item at a time.

func firewallCommands() []command {
	return []command{
		{name: "apply", args: "<file>", summary: "replace the mesh firewall with a YAML file, as `firewall show --yaml` writes it", setup: setupFirewallApply},
		{name: "check", args: "<from> <to> <service>", summary: "say what the mesh firewall does with one flow, now or --at a time (from and to: nodes or addresses, IPv4 or IPv6; service: tcp/PORT, udp/PORT or icmp)", setup: setupFirewallCheck},
		{name: "show", summary: "print the mesh firewall: policies, zones and the rules every server enforces", setup: setupFirewallShow},
	}
}

// maxFirewallFile bounds what apply reads.
const maxFirewallFile = 1 << 20

func setupFirewallShow(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	asYAML := fs.Bool("yaml", false, "print the document as YAML, to edit and give to `firewall apply`")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		v, err := client().Segment(ctx)
		switch {
		case err != nil:
			return err
		case *asJSON:
			return printJSON(e.stdout, v)
		case *asYAML:
			return export.WriteSegmentation(e.stdout, v.Config)
		}
		list := func(vs []string) string {
			if len(vs) == 0 {
				return "any"
			}
			return strings.Join(vs, ",")
		}
		inactive := map[string]bool{}
		for _, i := range v.Inactive {
			inactive[v.Rules[i].Name] = true
		}
		fmt.Fprintln(e.stdout, "policies (the first that matches decides):")
		tw := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "#\tNAME\tACTION\tFROM\tTO\tSERVICES\tWHEN")
		scheduled := v.Config.TimeZone != ""
		for i, p := range v.Config.Policies {
			name := p.Name
			if p.Disabled {
				name += " (disabled)"
			}
			when := "always"
			if p.Schedule != nil {
				scheduled, when = true, p.Schedule.String()
				if inactive["policy: "+p.Name] {
					when += " (not now)"
				}
			}
			fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\t%s\n", i+1, name, p.Action, list(p.From), list(p.To), list(p.Services), when)
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		if scheduled {
			whose := ""
			if v.Config.TimeZone == "" {
				whose = ", the server's"
			}
			fmt.Fprintf(e.stdout, "schedules are read in %s%s; now %s\n", v.TimeZone, whose, v.Now)
		}
		if len(v.Zones) > 0 {
			fmt.Fprintln(e.stdout, "\nzones:")
			tw = tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tWITHIN\tMEMBERS\tADDRESSES")
			for i, z := range v.Config.Zones {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", z.Name, z.Intra, list(z.Members), strings.Join(v.Zones[i].Addresses, ","))
			}
			if err := tw.Flush(); err != nil {
				return err
			}
		}
		fmt.Fprintln(e.stdout, "\nwhat every server enforces (then drops the rest):")
		for i, r := range v.Rules {
			note := ""
			if slices.Contains(v.Inactive, i) {
				note = " (not now)"
			}
			fmt.Fprintf(e.stdout, "  %d. %s%s\n", i+1, r, note)
		}
		printWarnings(e.stdout, v.Warnings)
		return nil
	}
}

func setupFirewallApply(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 1, "<file>"); err != nil {
			return err
		}
		file, err := validate.FilePath("file", args[0])
		if err != nil {
			return usagef("%v", err)
		}
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		doc, err := export.ReadSegmentation(io.LimitReader(f, maxFirewallFile))
		_ = f.Close() // read-only; the decode error is the one that matters
		if err != nil {
			return err
		}
		ch, err := client().SetSegment(ctx, doc)
		switch {
		case err != nil:
			return err
		case *asJSON:
			return printJSON(e.stdout, ch)
		case ch.Version == 0:
			fmt.Fprintln(e.stdout, "the mesh firewall already matched the file; no config changed")
		default:
			fmt.Fprintf(e.stdout, "mesh firewall replaced as config v%d; the servers apply it now\n", ch.Version)
		}
		printWarnings(e.stdout, ch.Warnings)
		return nil
	}
}

func setupFirewallCheck(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	at := fs.String("at", "", "judge the flow at this time, in RFC 3339 form such as 2026-10-23T09:30:00+08:00 (default now)")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 3, "<from> <to> <service>"); err != nil {
			return err
		}
		if *at != "" {
			if _, err := validate.Timestamp("--at", *at); err != nil {
				return usagef("%v", err)
			}
		}
		r, err := client().Reach(ctx, api.ReachRequest{From: args[0], To: args[1], Service: args[2], At: *at})
		switch {
		case err != nil:
			return err
		case *asJSON:
			return printJSON(e.stdout, r)
		}
		verdict := "dropped"
		if r.Allowed {
			verdict = "allowed"
		}
		where := r.Rule
		if r.Index >= 0 {
			where = "rule " + strconv.Itoa(r.Index+1) + ", " + r.Rule
		}
		fmt.Fprintf(e.stdout, "%s -> %s %s: %s (%s)\n", r.From, r.To, r.Service, verdict, where)
		if len(r.Inactive) > 0 {
			fmt.Fprintf(e.stdout, "at %s, not in force (scheduled): %s\n", r.At, strings.Join(r.Inactive, ", "))
		}
		return nil
	}
}
