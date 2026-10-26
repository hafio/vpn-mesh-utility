package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/client"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// JoinOnStart is how an agent in a container joins (amendment A47): once
// per join token. A token it has not used yet -- the first start's, or a
// new one from `meshctl node token` -- is joined with and remembered by its
// fingerprint in the data directory, so later starts with the same token
// leave the enrollment alone. No token is fine once the node has joined.
//
// A token the server calls spent (used or expired) is remembered too, so
// restarts do not ask again; any other failure is not. Either way a node
// that has joined before keeps running on that enrollment: a bad new token
// never takes a working node down. A node that has never joined cannot run,
// and gets an error that says what to do. The service is never installed
// (the container is the service).
func JoinOnStart(ctx context.Context, o JoinOptions) error {
	return joinOnStart(ctx, o, Join)
}

// refusedPause is how long a start refused for a spent token waits before
// it exits, so a container's restart loop stays quiet.
const refusedPause = 30 * time.Second

// ErrNoToken is JoinOnStart's error for a node that has never joined and
// was given no join token.
var ErrNoToken = errors.New("this node has not joined a mesh and has no join token: set MESHCTL_TOKEN (or MESHCTL_TOKEN_FILE) " +
	"to the node's token from the web UI or `meshctl node token <name>`")

func joinOnStart(ctx context.Context, o JoinOptions, join func(context.Context, JoinOptions) (proto.JoinResponse, error)) error {
	log := o.Log
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	dir, err := state.Open(o.DataDir)
	if err != nil {
		return err
	}
	joined := true
	if _, err := dir.Enrollment(); errors.Is(err, state.ErrNotJoined) {
		joined = false
	} else if err != nil {
		return err
	}
	if o.Token == "" {
		if !joined {
			return ErrNoToken
		}
		return nil
	}
	rec, err := dir.JoinRecord()
	if err != nil {
		return err
	}
	sum := tokenFingerprint(o.Token)
	// A copy refused on another machine than the node's wiped itself and
	// kept why (amendment A50): the token it joined with is spent here, and
	// a restart loop is told so without asking the server again.
	if ref, ok, _ := dir.Refused(); !joined && ok && ref.Token == sum { // a damaged refused.json just lets the join ask
		pause := o.Pause
		if pause == nil {
			pause = time.Sleep
		}
		pause(refusedPause)
		return fmt.Errorf("the control plane refused this node on this machine (%s: %s), so its data dir was wiped and this join token "+
			"is spent here; set MESHCTL_TOKEN (or MESHCTL_TOKEN_FILE) to a new token from `meshctl node token <name>`", ref.Code, ref.Message)
	}
	if rec.Token == sum {
		if !joined {
			return errors.New("the server refused this join token on an earlier start and the node has not joined: " +
				"set MESHCTL_TOKEN (or MESHCTL_TOKEN_FILE) to a new token from `meshctl node token <name>`")
		}
		log.Debug("the join token was used on an earlier start; not joining again")
		return nil
	}
	o.Out = io.Discard
	resp, err := join(ctx, o)
	if err == nil {
		if err := dir.SaveJoinRecord(state.JoinRecord{Token: sum}); err != nil {
			return fmt.Errorf("joined, but %w", err)
		}
		log.Info("joined", "server", o.Server, "node", resp.Name, "role", resp.Role, "mesh", meshAddresses(resp), "node_id", resp.NodeID)
		return nil
	}
	refused := isRefusal(err)
	if refused {
		if serr := dir.SaveJoinRecord(state.JoinRecord{Token: sum}); serr != nil {
			log.Warn("could not remember the refused join token; the next start asks the server again", "err", serr)
		}
	}
	switch {
	case joined:
		log.Warn("joining with the new token failed; running on this node's earlier enrollment", "err", err)
		return nil
	case refused:
		return fmt.Errorf("%w; set MESHCTL_TOKEN (or MESHCTL_TOKEN_FILE) to a new token from `meshctl node token <name>`", err)
	}
	return fmt.Errorf("%w; the next start tries again", err)
}

// isRefusal is the server saying the token itself is spent: used, expired
// or unknown, which it says every time. Anything else -- a disabled node, a
// token for another name, a rate limit, a proxy's error page, no answer --
// may go another way once that is fixed, so the token is asked about again.
func isRefusal(err error) bool {
	var ce *client.Error
	return errors.As(err, &ce) && ce.Code == proto.CodeTokenInvalid
}

// tokenFingerprint is what the data directory keeps of a join token. It is
// not the sha256 the server keeps, so join.json is no lookup key there.
func tokenFingerprint(token string) string {
	sum := sha256.Sum256([]byte("meshctl start join token\x00" + token))
	return hex.EncodeToString(sum[:])
}
