package ui

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/boombuler/barcode/qr"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
)

// A phone's profile is shown as a QR code drawn inline as SVG, beside the
// same profile as text. The page is the only place the phone's private key
// ever is (amendment A6), so the code cannot be an image fetched by a second
// request: the server would have to keep the key until it came.

// qrCode is a QR code as one SVG path: View is the viewBox (the symbol and
// its quiet zone), Path draws the dark modules.
type qrCode struct {
	View string
	Path string
}

// quietZone is the light border a scanner needs, in modules (ISO 18004).
const quietZone = 4

// newQR draws text as a QR code with medium error correction. Byte mode is
// asked for directly: the automatic mode's error quotes the content, which
// here is a private key.
func newQR(text string) (qrCode, error) {
	code, err := qr.Encode(text, qr.M, qr.Unicode)
	if err != nil {
		return qrCode{}, fmt.Errorf("draw the QR code of a %d-byte profile: %w", len(text), err)
	}
	n := code.Bounds().Dx()
	dark := func(x, y int) bool {
		r, _, _, _ := code.At(x, y).RGBA()
		return r < 0x8000
	}
	var b strings.Builder
	for y := 0; y < n; y++ {
		for x := 0; x < n; {
			if !dark(x, y) {
				x++
				continue
			}
			run := 1
			for x+run < n && dark(x+run, y) {
				run++
			}
			fmt.Fprintf(&b, "M%d %dh%dv1h-%dz", x, y, run, run)
			x += run
		}
	}
	side := n + 2*quietZone
	return qrCode{View: fmt.Sprintf("-%d -%d %d %d", quietZone, quietZone, side, side), Path: b.String()}, nil
}

type phoneData struct {
	Node     string
	Exit     string
	QR       qrCode
	Profile  string
	Warnings []string
}

// phoneProfile makes a phone's profile with a new key pair and shows it
// once.
func (u *ui) phoneProfile(w http.ResponseWriter, r *http.Request, s auth.Session) {
	p, err := u.Nodes.Phone(r.Context(), actor(s), r.PathValue("name"), api.PhoneRequest{Exit: r.PostFormValue("exit")})
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderNode(w, r, s, status, page{Error: msg})
		return
	}
	d := phoneData{Node: p.Node.Name, Exit: p.Exit, Profile: p.Profile, Warnings: p.Warnings}
	pg := page{Flash: changed("New key for "+p.Node.Name, api.NodeChange{Version: p.Version})}
	if d.QR, err = newQR(p.Profile); err != nil {
		u.Log.Warn("phone profile has no QR code", "node", p.Node.Name, "err", err)
		pg.Error = "This profile is too long for a QR code; import the text below as a file instead."
	}
	pg.Data = d
	u.show(w, r, s, http.StatusOK, "phone", "nodes", "Profile for "+p.Node.Name, pg)
}

// phoneState says on a phone's page what its profile is.
func phoneState(n api.NodeInfo) string {
	switch {
	case !n.Enrolled:
		return "none yet: make one below"
	case !n.InSync:
		return fmt.Sprintf("out of date: made for config v%d, which has changed since (v%d); make a new one", n.AppliedVersion, n.ConfigVersion)
	}
	return fmt.Sprintf("current (config v%d)", n.AppliedVersion)
}
