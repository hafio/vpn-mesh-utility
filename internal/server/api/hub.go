package api

import (
	"sync"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Hub fans config events, and nudges that the mesh view changed, out to the
// agents' SSE streams. Each node has at most one stream: a new subscription
// ends the previous one, so an agent that reconnected through a proxy which
// kept the old connection open does not leak streams. Events are
// latest-wins: a stream that has not consumed its pending event gets the
// newer one instead, because only the newest version matters to the agent;
// pending nudges collapse into one.
type Hub struct {
	mu     sync.Mutex
	subs   map[int64]*subscription
	closed bool
}

type subscription struct {
	events chan proto.ConfigEvent
	mesh   chan struct{}
	done   chan struct{}
	once   sync.Once
}

func (s *subscription) end() { s.once.Do(func() { close(s.done) }) }

// NewHub returns an empty hub.
func NewHub() *Hub { return &Hub{subs: map[int64]*subscription{}} }

// Subscribe registers node's stream, ending any previous one. events delivers
// config events, mesh a nudge when the mesh view changed; done closes when
// the stream must end (replaced, kicked or shutdown). cancel unregisters and
// must always be called.
func (h *Hub) Subscribe(node int64) (events <-chan proto.ConfigEvent, mesh, done <-chan struct{}, cancel func()) {
	s := &subscription{events: make(chan proto.ConfigEvent, 1), mesh: make(chan struct{}, 1), done: make(chan struct{})}
	h.mu.Lock()
	if h.closed {
		s.end()
	} else {
		if old := h.subs[node]; old != nil {
			old.end()
		}
		h.subs[node] = s
	}
	h.mu.Unlock()
	return s.events, s.mesh, s.done, func() {
		h.mu.Lock()
		if h.subs[node] == s {
			delete(h.subs, node)
		}
		h.mu.Unlock()
		s.end()
	}
}

// Notify delivers each node's event to its stream, if it has one.
func (h *Hub) Notify(events map[int64]proto.ConfigEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for node, ev := range events {
		s := h.subs[node]
		if s == nil {
			continue
		}
		// Only Notify sends, under h.mu, so draining then sending cannot
		// block.
		select {
		case <-s.events:
		default:
		}
		s.events <- ev
	}
}

// NotifyMesh nudges every stream: the mesh view changed (amendment A38).
func (h *Hub) NotifyMesh() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, s := range h.subs {
		select {
		case s.mesh <- struct{}{}:
		default: // a nudge is already pending
		}
	}
}

// Kick ends node's stream: the node was disabled or removed.
func (h *Hub) Kick(node int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s := h.subs[node]; s != nil {
		s.end()
		delete(h.subs, node)
	}
}

// Close ends every stream and refuses new ones; the server is shutting down.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	for node, s := range h.subs {
		s.end()
		delete(h.subs, node)
	}
}

// Subscribers reports how many streams are open.
func (h *Hub) Subscribers() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
}
