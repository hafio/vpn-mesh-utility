package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// Heartbeat thresholds for the dashboard colours. A phone handshakes only
// while its tunnel is on (every two minutes, with keepalives), so a quiet
// phone is idle, not failing.
const (
	lateHeartbeat   = 90 * time.Second
	silentHeartbeat = 5 * time.Minute
	phoneActive     = 3 * time.Minute
)

// Health is a node's dashboard colour (proto.HealthOK, HealthWarn,
// HealthBad or HealthOff) and label. The web dashboard and the mesh view
// both show it.
func Health(n NodeInfo, now time.Time) (string, string) {
	age := now.Sub(time.Unix(n.LastSeen, 0))
	switch {
	case n.Disabled:
		return proto.HealthOff, "disabled"
	case n.Role == string(proto.RolePhone):
		return phoneHealth(n, now)
	case !n.Enrolled:
		return proto.HealthOff, "not joined"
	case n.DownSince > 0:
		return proto.HealthBad, "down since " + time.Unix(n.DownSince, 0).In(now.Location()).Format("2006-01-02 15:04 MST") + "; its pairs relay through other hubs"
	case n.LastSeen == 0:
		return proto.HealthBad, "no heartbeat yet"
	case age > silentHeartbeat:
		return proto.HealthBad, "silent for " + age.Round(time.Second).String()
	case n.Rollback != nil:
		return proto.HealthBad, fmt.Sprintf("rolled back v%d to v%d", n.Rollback.From, n.Rollback.To)
	case n.Copy != nil && now.Sub(time.Unix(n.Copy.At, 0)) < recentCopy:
		return proto.HealthWarn, CopyLine(*n.Copy, now.Location())
	case len(n.Degraded) > 0:
		return proto.HealthWarn, "degraded: " + strings.Join(n.Degraded, ", ")
	case !n.InSync:
		return proto.HealthWarn, fmt.Sprintf("applying v%d", n.ConfigVersion)
	case age > lateHeartbeat:
		return proto.HealthWarn, "late heartbeat"
	}
	return proto.HealthOK, "in sync"
}

// recentCopy is how long the dashboard shows a second copy of a node.
const recentCopy = 24 * time.Hour

// CopyLine says what the control plane saw of a second copy of a node
// (amendment A50), with times in loc.
func CopyLine(c CopyInfo, loc *time.Location) string {
	at := time.Unix(c.At, 0).In(loc).Format("2006-01-02 15:04 MST")
	switch {
	case c.Waiting:
		return fmt.Sprintf("a second copy on %s has waited since %s while this one is asked whether it runs", c.Hostname, at)
	case c.Code == proto.CodeForeign:
		return fmt.Sprintf("a copy on another machine (%s) was refused at %s", c.Hostname, at)
	}
	return fmt.Sprintf("a second copy on %s, the same machine, was refused at %s", c.Hostname, at)
}

// withCopy adds the node's second copy, if the control plane saw one.
func withCopy(tx *store.MeshTx, info *NodeInfo) error {
	s, err := tx.AgentSession(info.ID)
	switch {
	case err == nil && s.ChallengeAt != 0 && !s.Answered:
		info.Copy = &CopyInfo{Waiting: true, Hostname: s.WaitingHost, At: s.ChallengeAt}
		return nil
	case err != nil && !errors.Is(err, store.ErrNotFound):
		return err
	}
	r, err := tx.Refusal(info.ID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil
	case err != nil:
		return err
	}
	info.Copy = &CopyInfo{Code: r.Code, Hostname: r.Hostname, SameMachine: r.SameMachine, At: r.At}
	return nil
}

// phoneHealth is Health for a phone: what its servers see of it.
func phoneHealth(n NodeInfo, now time.Time) (string, string) {
	switch {
	case !n.Enrolled:
		return proto.HealthOff, "no profile yet"
	case !n.InSync:
		return proto.HealthWarn, "profile out of date"
	case n.LastHandshake == 0:
		return proto.HealthOff, "no handshake yet"
	case now.Sub(time.Unix(n.LastHandshake, 0)) > phoneActive:
		return proto.HealthOff, "idle"
	}
	return proto.HealthOK, "tunnel on"
}

// Seen is when the node was last heard of: its heartbeat, or for a phone
// the latest handshake a server saw.
func (n NodeInfo) Seen() int64 {
	if n.Role == string(proto.RolePhone) {
		return n.LastHandshake
	}
	return n.LastSeen
}
