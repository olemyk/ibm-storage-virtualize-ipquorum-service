package nodes

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/config"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// startNode starts a node on an ephemeral port.
func startNode(t *testing.T, b config.NodeBehavior, dropAfter time.Duration) *Node {
	t.Helper()
	n := New("flash-a", config.Node{Address: "10.0.0.1:1260", Listen: "127.0.0.1:0", Behavior: b, DropAfter: dropAfter}, discard())
	if err := n.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(n.Close)
	return n
}

func dial(t *testing.T, n *Node) net.Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", n.ListenAddr(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func readLine(c net.Conn, d time.Duration) (string, error) {
	c.SetReadDeadline(time.Now().Add(d))
	return bufio.NewReader(c).ReadString('\n')
}

func TestHandshakeAndKeepalive(t *testing.T) {
	n := startNode(t, config.BehaviorUp, 0)
	c := dial(t, n)
	r := bufio.NewReader(c)

	io.WriteString(c, "HELLO protocol=3 features_bitmap=0\n")
	c.SetReadDeadline(time.Now().Add(time.Second))
	line, err := r.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "HELLO-ACK") || !strings.Contains(line, "node=10.0.0.1:1260") {
		t.Fatalf("handshake reply = %q, err %v", line, err)
	}

	io.WriteString(c, "PING\n")
	c.SetReadDeadline(time.Now().Add(time.Second))
	if line, err := r.ReadString('\n'); err != nil || line != "PONG\n" {
		t.Fatalf("keepalive reply = %q, err %v", line, err)
	}
	if s := n.Status(); s.Accepted != 1 || s.Active != 1 || !s.Listening {
		t.Errorf("status = %+v", s)
	}
}

func TestDownRefusesConnections(t *testing.T) {
	n := startNode(t, config.BehaviorDown, 0)
	if n.ListenAddr() != "" || n.Status().Listening {
		t.Fatal("down node must not listen")
	}
}

func TestHandshakeTimeoutStaysSilent(t *testing.T) {
	n := startNode(t, config.BehaviorHandshakeTimeout, 0)
	c := dial(t, n)
	io.WriteString(c, "HELLO protocol=3 features_bitmap=0\n")
	_, err := readLine(c, 300*time.Millisecond)
	ne, ok := err.(net.Error)
	if !ok || !ne.Timeout() {
		t.Fatalf("expected a read timeout (no reply, connection open), got %v", err)
	}
}

func TestDropAfterClosesConnection(t *testing.T) {
	n := startNode(t, config.BehaviorDropAfter, 200*time.Millisecond)
	c := dial(t, n)
	r := bufio.NewReader(c)
	io.WriteString(c, "HELLO protocol=3 features_bitmap=0\n")
	c.SetReadDeadline(time.Now().Add(time.Second))
	if line, err := r.ReadString('\n'); err != nil || !strings.HasPrefix(line, "HELLO-ACK") {
		t.Fatalf("expected a normal handshake first, got %q %v", line, err)
	}
	c.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := r.ReadString('\n'); err != io.EOF && !strings.Contains(errString(err), "closed") && !strings.Contains(errString(err), "reset") {
		t.Fatalf("expected the connection to be dropped, got %v", err)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestSetBehaviorAtRuntime(t *testing.T) {
	n := startNode(t, config.BehaviorUp, 0)
	c := dial(t, n)
	addr := n.ListenAddr()

	if err := n.SetBehavior(config.BehaviorDown, 0); err != nil {
		t.Fatal(err)
	}
	// existing connection is closed
	c.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := bufio.NewReader(c).ReadString('\n'); err == nil {
		t.Error("existing connection should be closed when the node goes down")
	}
	// new connections are refused
	if _, err := net.DialTimeout("tcp", addr, 300*time.Millisecond); err == nil {
		t.Error("node should refuse connections while down")
	}

	if err := n.SetBehavior(config.BehaviorUp, 0); err != nil {
		t.Fatal(err)
	}
	c2 := dial(t, n)
	io.WriteString(c2, "HELLO protocol=3\n")
	if line, err := readLine(c2, time.Second); err != nil || !strings.HasPrefix(line, "HELLO-ACK") {
		t.Errorf("after coming back: %q %v", line, err)
	}

	if err := n.SetBehavior("bogus", 0); err == nil {
		t.Error("unknown behavior must be rejected")
	}
}

func TestSetStartAndClose(t *testing.T) {
	cl := config.Cluster{Name: "flash-a", Nodes: []config.Node{
		{Address: "10.0.0.1:1260", Listen: "127.0.0.1:0", Behavior: config.BehaviorUp},
		{Address: "10.0.0.2:1260", Listen: "127.0.0.1:0", Behavior: config.BehaviorDown},
	}}
	s := NewSet(cl, discard())
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if n, ok := s.Get("10.0.0.2:1260"); !ok || n.Status().Listening {
		t.Error("second node should exist and not listen")
	}
	if _, ok := s.Get("nope"); ok {
		t.Error("unknown address must not be found")
	}
	s.Close()
	if s.Nodes()[0].Status().Listening {
		t.Error("node still listening after Close")
	}
}

func TestListenCollisionReported(t *testing.T) {
	first := startNode(t, config.BehaviorUp, 0)
	second := New("flash-b", config.Node{Address: "10.0.0.9:1260", Listen: first.ListenAddr(), Behavior: config.BehaviorUp}, discard())
	if err := second.Start(); err == nil {
		second.Close()
		t.Fatal("expected a listen error for a port already in use")
	}
}
