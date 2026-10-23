package agent

import (
	"context"
	"net/netip"
	"slices"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/names"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The container runtime socket choices of Deps.RuntimeSocket.
const (
	RuntimeAuto = "auto"
	RuntimeOff  = "off"
)

// Container names timing (amendment A38).
const (
	// namesResync lists the containers again although the runtime said
	// nothing: its event stream may have dropped one.
	namesResync = 2 * time.Minute
	// namesRetry is how soon a report the server did not take is sent
	// again.
	namesRetry = 15 * time.Second
	// The runtime's event stream reconnects with backoff between these.
	watchBackoffMin = 5 * time.Second
	watchBackoffMax = time.Minute
)

// nameServer is the DNS step's resolver, which answers the containers'
// names (dns.Service).
type nameServer interface {
	SetNames(names []proto.MeshName)
}

// setNames hands the mesh view's container names to the resolver, and
// their IPv6 addresses to the engine (M8): when those change, proxyLoop
// applies the config again, so the containers' side answers neighbour
// discovery for them.
func (r *runner) setNames(ns []proto.MeshName) {
	if s, ok := r.d.DNS.(nameServer); ok {
		s.SetNames(ns)
	}
	if r.engine.SetMeshNames(ns) {
		kick(r.proxyKick)
	}
}

// proxyLoop applies the config again whenever the mesh's containers' IPv6
// addresses change. A pass that fails is the drift check's to retry.
func (r *runner) proxyLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.proxyKick:
		}
		res, err := r.reapply(ctx, r.engine.Apply)
		switch {
		case ctx.Err() != nil:
			return
		case err != nil:
			r.log.Error("the containers' side could not follow the mesh's containers; the drift check tries again", "err", err)
		case res.Changes > 0:
			r.log.Info("the containers' side follows the mesh's containers", "changes", res.Changes)
		}
	}
}

// openRuntime picks the container runtime the names are read from, unless
// a test gave one or the operator turned it off.
func (r *runner) openRuntime() {
	if r.d.Runtime != nil || r.d.RuntimeSocket == RuntimeOff {
		return
	}
	sock := r.d.RuntimeSocket
	if sock == "" || sock == RuntimeAuto {
		if sock = names.Detect(names.Sockets()); sock == "" {
			r.log.Debug("no container runtime socket found; containers' names are not reported", "looked_at", names.Sockets())
			return
		}
	}
	r.d.Runtime = names.New(sock)
	r.log.Info("reading the containers' names from the container runtime", "socket", sock)
}

// containerSubnets are the applied config's container network and its
// IPv6 side (M8); invalid while the node's containers are not on the mesh,
// or have no IPv6.
func (r *runner) containerSubnets() (netip.Prefix, netip.Prefix) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.applied == nil || r.applied.Containers == nil {
		return netip.Prefix{}, netip.Prefix{}
	}
	p, _ := netip.ParsePrefix(r.applied.Containers.Subnet)
	p6, _ := netip.ParsePrefix(r.applied.Containers.Subnet6)
	return p, p6
}

// namesLoop keeps the server's copy of this node's container names
// (amendment A38). It lists the runtime's containers on the mesh network
// once things settle after the runtime reports a change, after a config
// arrives (the network may have moved) and every namesResync, and puts the
// names whenever they differ from what the server last took.
func (r *runner) namesLoop(ctx context.Context) {
	if r.d.Runtime == nil {
		return
	}
	changed := make(chan struct{}, 1)
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		r.watchRuntime(ctx, changed)
	}()
	defer func() { <-watched }()
	resync := time.NewTicker(namesResync)
	defer resync.Stop()
	var (
		sent    []proto.ContainerName
		synced  bool
		lastErr string
	)
	kick(changed)
	for {
		select {
		case <-ctx.Done():
			return
		case <-changed:
		case <-r.namesKick:
		case <-resync.C:
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(r.d.NamesSettle):
		}
		drain(changed)
		drain(r.namesKick)
		subnet, subnet6 := r.containerSubnets()
		if !subnet.IsValid() {
			continue
		}
		cs, err := r.d.Runtime.List(ctx)
		if err != nil {
			// The same failure every minute is logged once.
			if ctx.Err() == nil && err.Error() != lastErr {
				r.log.Warn("containers' names not read; the next change, or the resync, tries again", "err", err)
			}
			lastErr = err.Error()
			continue
		}
		lastErr = ""
		want, over := names.Of(cs, subnet, subnet6)
		if over > 0 {
			r.log.Warn("more containers' names than one report carries; the rest are left out", "left_out", over, "max", proto.MaxNames)
		}
		if synced && slices.Equal(want, sent) {
			continue
		}
		resp, err := r.cl.Names(ctx, proto.NamesRequest{Names: want})
		if err != nil {
			r.serverError("report the containers' names", err)
			synced = false
			time.AfterFunc(namesRetry, func() { kick(r.namesKick) })
			continue
		}
		r.contact()
		sent, synced = want, true
		r.log.Info("containers' names reported", "names", len(want), "stored", resp.Stored)
		for _, d := range resp.Dropped {
			r.log.Warn("container name not published", "name", d.Name, "ip", d.IP, "reason", d.Reason)
		}
	}
}

// watchRuntime follows the runtime's event stream, kicking changed for
// every event, and reconnects with backoff; a stream that comes back is
// itself a change, since events may have been missed.
func (r *runner) watchRuntime(ctx context.Context, changed chan struct{}) {
	backoff, warned := watchBackoffMin, false
	for ctx.Err() == nil {
		opened := time.Now()
		err := r.d.Runtime.Watch(ctx, func() { kick(changed) })
		if ctx.Err() != nil {
			return
		}
		if time.Since(opened) > 2*watchBackoffMax {
			backoff, warned = watchBackoffMin, false
		}
		if !warned {
			r.log.Warn("container runtime events stopped; the names are listed every 2 minutes until they are back", "err", err)
			warned = true
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(2*backoff, watchBackoffMax)
		kick(changed)
	}
}

func drain(ch chan struct{}) {
	select {
	case <-ch:
	default:
	}
}
