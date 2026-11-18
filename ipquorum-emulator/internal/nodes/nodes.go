// Package nodes implements the mock quorum nodes: TCP listeners that the fake
// ip_quorum.jar connects to.
//
// The wire protocol is a simplified, line based protocol of our own. The real
// IP quorum protocol is not publicly documented and is NOT emulated.
//
//	client -> node   HELLO protocol=3 features_bitmap=0\n
//	node   -> client HELLO-ACK protocol=3 node=<address>\n
//	client -> node   PING\n          (repeated)
//	node   -> client PONG\n
package nodes

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/config"
)

// Node is one mock quorum node. Its behavior can be changed at runtime.
type Node struct {
	cluster string
	cfg     config.Node
	log     *slog.Logger

	mu        sync.Mutex
	behavior  config.NodeBehavior
	dropAfter time.Duration
	ln        net.Listener
	conns     map[net.Conn]struct{}
	accepted  int
	wg        sync.WaitGroup
	closed    bool
}

// Status is a point-in-time view of a node.
type Status struct {
	Address   string              `json:"address"`
	Listen    string              `json:"listen"`
	Behavior  config.NodeBehavior `json:"behavior"`
	Listening bool                `json:"listening"`
	Accepted  int                 `json:"accepted_connections"`
	Active    int                 `json:"active_connections"`
}

// New creates a node. Call Start to apply its behavior.
func New(cluster string, cfg config.Node, log *slog.Logger) *Node {
	return &Node{
		cluster:   cluster,
		cfg:       cfg,
		log:       log.With("cluster", cluster, "node", cfg.Address),
		behavior:  cfg.Behavior,
		dropAfter: cfg.DropAfter,
		conns:     map[net.Conn]struct{}{},
	}
}

// Address returns the node address embedded in the generated JAR.
func (n *Node) Address() string { return n.cfg.Address }

// Start applies the configured behavior (listens unless behavior is "down").
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.applyLocked()
}

// SetBehavior changes the behavior at runtime. Existing connections are closed
// so that the change is visible to connected clients.
func (n *Node) SetBehavior(b config.NodeBehavior, dropAfter time.Duration) error {
	if !b.Valid() {
		return fmt.Errorf("unknown behavior %q", b)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return fmt.Errorf("node is closed")
	}
	n.behavior = b
	if dropAfter > 0 {
		n.dropAfter = dropAfter
	}
	n.closeConnsLocked()
	return n.applyLocked()
}

// applyLocked makes the listener match the behavior. Caller holds n.mu.
func (n *Node) applyLocked() error {
	if n.behavior == config.BehaviorDown {
		n.stopListenerLocked()
		n.log.Info("node is down (not listening)")
		return nil
	}
	if n.ln != nil {
		n.log.Info("node behavior changed", "behavior", n.behavior)
		return nil
	}
	ln, err := net.Listen("tcp", n.cfg.Listen)
	if err != nil {
		return fmt.Errorf("cluster %s node %s: listen on %s: %w", n.cluster, n.cfg.Address, n.cfg.Listen, err)
	}
	n.ln = ln
	n.wg.Add(1)
	go n.acceptLoop(ln)
	n.log.Info("node listening", "listen", ln.Addr().String(), "behavior", n.behavior)
	return nil
}

func (n *Node) stopListenerLocked() {
	if n.ln != nil {
		_ = n.ln.Close()
		n.ln = nil
	}
	n.closeConnsLocked()
}

func (n *Node) closeConnsLocked() {
	for c := range n.conns {
		_ = c.Close()
	}
}

// Close stops the node and waits for its goroutines.
func (n *Node) Close() {
	n.mu.Lock()
	n.closed = true
	n.stopListenerLocked()
	n.mu.Unlock()
	n.wg.Wait()
}

// Status returns the current state.
func (n *Node) Status() Status {
	n.mu.Lock()
	defer n.mu.Unlock()
	return Status{
		Address:   n.cfg.Address,
		Listen:    n.cfg.Listen,
		Behavior:  n.behavior,
		Listening: n.ln != nil,
		Accepted:  n.accepted,
		Active:    len(n.conns),
	}
}

// ListenAddr returns the bound address, or "" when the node is down.
func (n *Node) ListenAddr() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.ln == nil {
		return ""
	}
	return n.ln.Addr().String()
}

func (n *Node) acceptLoop(ln net.Listener) {
	defer n.wg.Done()
	for {
		c, err := ln.Accept()
		if err != nil {
			return // listener closed
		}
		n.mu.Lock()
		if n.closed || n.ln != ln {
			n.mu.Unlock()
			_ = c.Close()
			continue
		}
		n.conns[c] = struct{}{}
		n.accepted++
		behavior, dropAfter := n.behavior, n.dropAfter
		n.wg.Add(1)
		n.mu.Unlock()
		go n.serve(c, behavior, dropAfter)
	}
}

func (n *Node) serve(c net.Conn, behavior config.NodeBehavior, dropAfter time.Duration) {
	defer n.wg.Done()
	defer func() {
		_ = c.Close()
		n.mu.Lock()
		delete(n.conns, c)
		n.mu.Unlock()
	}()
	log := n.log.With("remote", c.RemoteAddr().String())
	log.Info("connection accepted", "behavior", behavior)

	if behavior == config.BehaviorDropAfter {
		timer := time.AfterFunc(dropAfter, func() {
			log.Info("dropping connection", "after", dropAfter)
			_ = c.Close()
		})
		defer timer.Stop()
	}

	r := bufio.NewReader(c)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			log.Info("connection closed", "reason", err)
			return
		}
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "HELLO"):
			if behavior == config.BehaviorHandshakeTimeout {
				log.Info("handshake received, not answering (handshake-timeout)")
				continue // keep the connection open and stay silent
			}
			if _, err := fmt.Fprintf(c, "HELLO-ACK protocol=3 node=%s\n", n.cfg.Address); err != nil {
				return
			}
		case line == "PING":
			if behavior == config.BehaviorHandshakeTimeout {
				continue
			}
			if _, err := fmt.Fprint(c, "PONG\n"); err != nil {
				return
			}
		default:
			log.Warn("unexpected message", "line", line)
		}
	}
}

// Set is the group of mock nodes of one cluster.
type Set struct {
	cluster string
	nodes   []*Node
}

// NewSet creates the nodes of a cluster. Nodes are not started.
func NewSet(cl config.Cluster, log *slog.Logger) *Set {
	s := &Set{cluster: cl.Name}
	for _, nc := range cl.Nodes {
		s.nodes = append(s.nodes, New(cl.Name, nc, log))
	}
	return s
}

// Nodes returns the nodes in configuration order.
func (s *Set) Nodes() []*Node { return s.nodes }

// Get returns the node with the given address.
func (s *Set) Get(address string) (*Node, bool) {
	for _, n := range s.nodes {
		if n.Address() == address {
			return n, true
		}
	}
	return nil, false
}

// Start starts every node. On error already started nodes are closed.
func (s *Set) Start() error {
	for _, n := range s.nodes {
		if err := n.Start(); err != nil {
			s.Close()
			return err
		}
	}
	return nil
}

// Close stops every node.
func (s *Set) Close() {
	for _, n := range s.nodes {
		n.Close()
	}
}
