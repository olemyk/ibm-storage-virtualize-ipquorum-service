package control

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/cluster"
	"github.com/olemyk/ipquorum-emulator/internal/config"
	"github.com/olemyk/ipquorum-emulator/internal/nodes"
)

func setup(t *testing.T) (*httptest.Server, *cluster.Registry, map[string]*nodes.Set) {
	t.Helper()
	// Built directly: Parse rejects port 0, but tests need ephemeral ports.
	node := func(addr string, b config.NodeBehavior) config.Node {
		return config.Node{Address: addr, Listen: "127.0.0.1:0", Behavior: b, DropAfter: 10 * time.Second}
	}
	cfg := &config.Config{Clusters: []config.Cluster{
		{Name: "flash-a", Scenario: config.ScenarioHealthy, RESTListen: "127.0.0.1:18443", Nodes: []config.Node{
			node("10.0.0.1:1260", config.BehaviorUp),
			node("10.0.0.2:1260", config.BehaviorDown),
		}},
		{Name: "flash-b", Scenario: config.ScenarioHealthy, RESTListen: "127.0.0.1:18444", Nodes: []config.Node{
			node("10.0.1.1:1260", config.BehaviorUp),
		}},
	}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := cluster.NewRegistry(cfg)
	sets := map[string]*nodes.Set{}
	for _, cl := range reg.All() {
		set := nodes.NewSet(cl.Config(), log)
		if err := set.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(set.Close)
		sets[cl.Name()] = set
	}
	ts := httptest.NewServer(NewHandler(reg, sets, log))
	t.Cleanup(ts.Close)
	return ts, reg, sets
}

func call(t *testing.T, ts *httptest.Server, method, path string, body any) (int, []byte) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, ts.URL+path, rd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data
}

func TestListAndGet(t *testing.T) {
	ts, _, _ := setup(t)

	code, body := call(t, ts, "GET", "/clusters", nil)
	var views []ClusterView
	if code != 200 || json.Unmarshal(body, &views) != nil || len(views) != 2 {
		t.Fatalf("list = %d %s", code, body)
	}
	if views[0].Name != "flash-a" || len(views[0].Nodes) != 2 || views[1].Name != "flash-b" {
		t.Errorf("unexpected views: %+v", views)
	}

	if code, _ := call(t, ts, "GET", "/clusters/flash-b", nil); code != 200 {
		t.Errorf("get flash-b = %d", code)
	}
	if code, _ := call(t, ts, "GET", "/clusters/ghost", nil); code != 404 {
		t.Errorf("get ghost = %d, want 404", code)
	}
	if code, _ := call(t, ts, "GET", "/healthz", nil); code != 200 {
		t.Errorf("healthz = %d", code)
	}
}

func TestSetScenarioAffectsOnlyThatCluster(t *testing.T) {
	ts, reg, _ := setup(t)
	if code, body := call(t, ts, "POST", "/clusters/flash-a/scenario", map[string]string{"scenario": "auth-401"}); code != 200 {
		t.Fatalf("set scenario = %d %s", code, body)
	}
	a, _ := reg.Get("flash-a")
	b, _ := reg.Get("flash-b")
	if a.Scenario() != config.ScenarioAuth401 || b.Scenario() != config.ScenarioHealthy {
		t.Errorf("scenarios: a=%s b=%s", a.Scenario(), b.Scenario())
	}
	if code, _ := call(t, ts, "POST", "/clusters/flash-a/scenario", map[string]string{"scenario": "nope"}); code != 400 {
		t.Errorf("unknown scenario = %d, want 400", code)
	}
	if code, _ := call(t, ts, "POST", "/clusters/ghost/scenario", map[string]string{"scenario": "healthy"}); code != 404 {
		t.Errorf("unknown cluster = %d, want 404", code)
	}
}

func TestSetNodeBehavior(t *testing.T) {
	ts, _, sets := setup(t)
	n, _ := sets["flash-a"].Get("10.0.0.1:1260")
	other, _ := sets["flash-b"].Get("10.0.1.1:1260")
	addr := n.ListenAddr()

	code, body := call(t, ts, "POST", "/clusters/flash-a/nodes/10.0.0.1:1260/behavior", map[string]string{"behavior": "down"})
	if code != 200 {
		t.Fatalf("set behavior = %d %s", code, body)
	}
	if _, err := net.DialTimeout("tcp", addr, 300*time.Millisecond); err == nil {
		t.Error("node should be down")
	}
	if !other.Status().Listening {
		t.Error("node of the other cluster must be unaffected")
	}

	code, _ = call(t, ts, "POST", "/clusters/flash-a/nodes/10.0.0.1:1260/behavior", map[string]string{"behavior": "drop-after", "drop_after": "2s"})
	if code != 200 || n.Status().Behavior != config.BehaviorDropAfter || !n.Status().Listening {
		t.Errorf("drop-after = %d %+v", code, n.Status())
	}

	for _, tt := range []struct {
		path string
		body map[string]string
		want int
	}{
		{"/clusters/flash-a/nodes/10.0.0.1:1260/behavior", map[string]string{"behavior": "bogus"}, 400},
		{"/clusters/flash-a/nodes/10.0.0.1:1260/behavior", map[string]string{"behavior": "up", "drop_after": "xx"}, 400},
		{"/clusters/flash-a/nodes/9.9.9.9:1/behavior", map[string]string{"behavior": "up"}, 404},
		{"/clusters/flash-b/nodes/10.0.0.1:1260/behavior", map[string]string{"behavior": "up"}, 404}, // node of a different cluster
	} {
		if code, _ := call(t, ts, "POST", tt.path, tt.body); code != tt.want {
			t.Errorf("%s %v = %d, want %d", tt.path, tt.body, code, tt.want)
		}
	}
}

func TestReset(t *testing.T) {
	ts, reg, sets := setup(t)
	a, _ := reg.Get("flash-a")
	a.SetScenario(config.ScenarioAuth403)
	a.CreateQuorumApp(cluster.MkquorumappRequest{})
	n1, _ := sets["flash-a"].Get("10.0.0.1:1260")
	n2, _ := sets["flash-a"].Get("10.0.0.2:1260")
	n1.SetBehavior(config.BehaviorDown, 0)
	n2.SetBehavior(config.BehaviorUp, 0)

	if code, body := call(t, ts, "POST", "/clusters/flash-a/reset", nil); code != 200 {
		t.Fatalf("reset = %d %s", code, body)
	}
	if a.Scenario() != config.ScenarioHealthy || a.QuorumAppCreated() {
		t.Errorf("cluster not reset: %+v", a.Snapshot())
	}
	if n1.Status().Behavior != config.BehaviorUp || !n1.Status().Listening {
		t.Errorf("node 1 should be back up: %+v", n1.Status())
	}
	if n2.Status().Behavior != config.BehaviorDown || n2.Status().Listening {
		t.Errorf("node 2 should be down again (configured): %+v", n2.Status())
	}
}

func TestListNodes(t *testing.T) {
	ts, _, _ := setup(t)
	code, body := call(t, ts, "GET", "/clusters/flash-a/nodes", nil)
	var st []nodes.Status
	if code != 200 || json.Unmarshal(body, &st) != nil || len(st) != 2 {
		t.Fatalf("nodes = %d %s", code, body)
	}
	if st[1].Behavior != config.BehaviorDown || st[1].Listening {
		t.Errorf("node 2 status: %+v", st[1])
	}
}
