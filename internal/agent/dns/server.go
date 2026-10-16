package dns

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
	"time"

	mdns "github.com/miekg/dns"
)

// forwardTimeout bounds each upstream try.
const forwardTimeout = 2 * time.Second

// resolver serves one address over UDP and TCP.
type resolver struct {
	// want is the address asked for, addr the one bound (they differ only
	// in a test's port 0).
	want, addr netip.AddrPort
	udp, tcp   *mdns.Server
	pc         net.PacketConn
	ln         net.Listener
	stopped    chan struct{} // closed once either server stops serving
	once       sync.Once
}

// serve binds want and starts answering with h. The sockets are opened
// here, so a bind failure (the port taken, the address gone) is reported
// at once rather than from a goroutine.
func serve(want netip.AddrPort, h mdns.Handler) (*resolver, error) {
	pc, ln, addr, err := bindPair(want)
	if err != nil {
		return nil, err
	}
	r := &resolver{want: want, addr: addr, pc: pc, ln: ln, stopped: make(chan struct{})}
	started := make(chan struct{}, 2)
	notify := func() { started <- struct{}{} }
	r.udp = &mdns.Server{PacketConn: pc, Handler: h, NotifyStartedFunc: notify}
	r.tcp = &mdns.Server{Listener: ln, Handler: h, NotifyStartedFunc: notify, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	for _, s := range []*mdns.Server{r.udp, r.tcp} {
		go func() {
			_ = s.ActivateAndServe() // returns once shut down, or when its socket fails
			r.once.Do(func() { close(r.stopped) })
		}()
	}
	for range 2 {
		select {
		case <-started:
		case <-r.stopped:
			r.close()
			return nil, fmt.Errorf("the resolver on %s stopped as it started", addr)
		}
	}
	return r, nil
}

// bindPair opens want over UDP and TCP. For port 0 (tests) TCP takes the
// port the system gave UDP; when TCP may not have it -- Windows reserves
// port ranges for TCP alone (`netsh int ipv4 show excludedportrange
// protocol=tcp`) -- another port is tried.
func bindPair(want netip.AddrPort) (net.PacketConn, net.Listener, netip.AddrPort, error) {
	for try := 1; ; try++ {
		pc, err := net.ListenPacket("udp", want.String())
		if err != nil {
			return nil, nil, netip.AddrPort{}, err
		}
		addr := want
		if want.Port() == 0 {
			addr = netip.AddrPortFrom(want.Addr(), uint16(pc.LocalAddr().(*net.UDPAddr).Port))
		}
		ln, err := net.Listen("tcp", addr.String())
		if err == nil {
			return pc, ln, addr, nil
		}
		_ = pc.Close() // nothing was served on it
		if want.Port() != 0 || try == 10 {
			return nil, nil, netip.AddrPort{}, err
		}
	}
}

// dead reports whether a server stopped serving on its own.
func (r *resolver) dead() bool {
	select {
	case <-r.stopped:
		return true
	default:
		return false
	}
}

func (r *resolver) close() {
	// A server that never started refuses Shutdown; closing its socket
	// ends it either way.
	_ = r.udp.Shutdown()
	_ = r.tcp.Shutdown()
	_ = r.pc.Close()
	_ = r.ln.Close()
}

// exchange forwards req to each upstream in turn, over the transport the
// query came in on (so a truncated UDP answer sends the client to TCP, as
// it would without us), and answers SERVFAIL when none replies.
func exchange(req *mdns.Msg, upstream []string, network string) (*mdns.Msg, []error) {
	c := &mdns.Client{Net: network, Timeout: forwardTimeout}
	var errs []error
	for _, u := range upstream {
		resp, _, err := c.Exchange(req, u)
		if err == nil {
			return resp, errs
		}
		errs = append(errs, fmt.Errorf("%s: %w", u, err))
	}
	return new(mdns.Msg).SetRcode(req, mdns.RcodeServerFailure), errs
}
