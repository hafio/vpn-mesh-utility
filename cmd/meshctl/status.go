package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func setupStatus(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	asJSON := fs.Bool("json", false, "print the status as JSON")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		path, err := sock()
		if err != nil {
			return err
		}
		st, err := localapi.NewClient(path).Status(ctx)
		if err != nil {
			return err
		}
		if *asJSON {
			enc := json.NewEncoder(e.stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(st)
		}
		printStatus(e.stdout, st, e.now())
		return nil
	}
}

func printStatus(w io.Writer, st proto.LocalStatus, now time.Time) {
	fmt.Fprintf(w, "node      %s (%s), mesh IP %s\n", st.Node.Name, st.Node.Role, st.MeshIP)
	fmt.Fprintf(w, "agent     %s, %s WireGuard\n", st.AgentVersion, st.WGBackend)
	fmt.Fprintf(w, "server    %s, last answered %s\n", st.Server, ago(st.LastContact, now))
	if st.AppliedVersion == 0 && st.AppliedHash == "" {
		fmt.Fprintln(w, "config    none applied yet")
	} else {
		fmt.Fprintf(w, "config    v%d, applied %s\n", st.AppliedVersion, ago(st.AppliedAt, now))
	}
	if st.Exit != "" {
		fmt.Fprintf(w, "exit      via %s\n", st.Exit)
	}
	if st.DNS != "" {
		fmt.Fprintf(w, "dns       %s\n", dnsLine(st))
	}
	if rb := st.Rollback; rb != nil {
		fmt.Fprintf(w, "rollback  v%d -> v%d: %s\n", rb.From, rb.To, rb.Reason)
	}
	for _, d := range st.Degraded {
		fmt.Fprintf(w, "degraded  %s\n", d)
	}
	for _, e := range st.Errors {
		fmt.Fprintf(w, "error     %s\n", e)
	}
	if len(st.Peers) == 0 {
		return
	}
	fmt.Fprintln(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PEER\tENDPOINT\tHANDSHAKE\tRX\tTX")
	for _, p := range st.Peers {
		ep := p.Endpoint
		if ep == "" {
			ep = "-"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", p.Name, ep, ago(p.LastHandshake, now), bytesize(p.RxBytes), bytesize(p.TxBytes))
	}
	_ = tw.Flush() // nowhere left to report a failed write
}

// ago renders a unix time relative to now; 0 is "never".
func ago(unix int64, now time.Time) string {
	if unix == 0 {
		return "never"
	}
	d := now.Sub(time.Unix(unix, 0)).Round(time.Second)
	if d < 0 {
		d = 0
	}
	return d.String() + " ago"
}

func bytesize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
