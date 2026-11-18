package config

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const minimal = `
clusters:
  - name: flash-a
    password: pw
    nodes:
      - address: "10.0.0.1:1260"
`

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse([]byte(minimal))
	if err != nil {
		t.Fatal(err)
	}
	cl := cfg.Clusters[0]
	if cl.Username != DefaultUsername || cl.RESTListen != DefaultRESTListen ||
		cl.Scenario != ScenarioHealthy || cl.RateLimitCount != DefaultRateLimitCount ||
		cl.DownloadDelay != DefaultDownloadDelay {
		t.Errorf("defaults not applied: %+v", cl)
	}
	n := cl.Nodes[0]
	if n.Behavior != BehaviorUp || n.Listen != ":1260" || n.DropAfter != DefaultDropAfter {
		t.Errorf("node defaults not applied: %+v", n)
	}
}

func TestParseDurations(t *testing.T) {
	cfg, err := Parse([]byte(`
clusters:
  - name: a
    password: pw
    download_delay: 1500ms
    nodes:
      - address: "h:1"
        behavior: drop-after
        drop_after: 7s
`))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Clusters[0].DownloadDelay; got != 1500*time.Millisecond {
		t.Errorf("download_delay = %v", got)
	}
	if got := cfg.Clusters[0].Nodes[0].DropAfter; got != 7*time.Second {
		t.Errorf("drop_after = %v", got)
	}
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"no clusters", "clusters: []", "at least one cluster"},
		{"unknown field", minimal + "\nbogus: 1", "bogus"},
		{"missing name", `
clusters:
  - password: pw
    nodes: [{address: "h:1"}]`, "needs a name"},
		{"duplicate name", `
clusters:
  - {name: a, password: pw, rest_listen: ":1", nodes: [{address: "h:10"}]}
  - {name: a, password: pw, rest_listen: ":2", nodes: [{address: "h:11"}]}`, "duplicate cluster name"},
		{"missing password", `
clusters:
  - {name: a, nodes: [{address: "h:1"}]}`, "password is required"},
		{"unknown scenario", `
clusters:
  - {name: a, password: pw, scenario: nope, nodes: [{address: "h:1"}]}`, "unknown scenario"},
		{"no nodes", `
clusters:
  - {name: a, password: pw}`, "at least one node"},
		{"unknown behavior", `
clusters:
  - {name: a, password: pw, nodes: [{address: "h:1", behavior: nope}]}`, "unknown behavior"},
		{"bad node address", `
clusters:
  - {name: a, password: pw, nodes: [{address: "nohost"}]}`, "invalid address"},
		{"node address without host", `
clusters:
  - {name: a, password: pw, nodes: [{address: ":1260"}]}`, "needs a host"},
		{"duplicate node", `
clusters:
  - {name: a, password: pw, nodes: [{address: "h:1"}, {address: "h:1"}]}`, "duplicate node address"},
		{"unknown partner (strict)", `
clusters:
  - {name: a, password: pw, partner_system: ghost, nodes: [{address: "h:1"}]}`, "not a known cluster"},
		{"partner is self", `
clusters:
  - {name: a, password: pw, partner_system: a, nodes: [{address: "h:1"}]}`, "not be the cluster itself"},
		{"rest port collision", `
clusters:
  - {name: a, password: pw, rest_listen: ":7443", nodes: [{address: "h:1"}]}
  - {name: b, password: pw, rest_listen: ":7443", nodes: [{address: "h:2"}]}`, "listener collision"},
		{"node/rest collision", `
clusters:
  - {name: a, password: pw, rest_listen: "127.0.0.1:9", nodes: [{address: "h:9", listen: "127.0.0.1:9"}]}`, "listener collision"},
		{"control collision", `
control_listen: ":7443"
clusters:
  - {name: a, password: pw, nodes: [{address: "h:1"}]}`, "listener collision"},
		{"bad port", `
clusters:
  - {name: a, password: pw, rest_listen: ":99999", nodes: [{address: "h:1"}]}`, "invalid port"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestSameNodePortOnDifferentHostsAllowed(t *testing.T) {
	_, err := Parse([]byte(`
clusters:
  - name: a
    password: pw
    rest_listen: "127.0.0.1:8443"
    nodes:
      - {address: "10.0.0.1:1260", listen: "127.0.0.1:1260"}
      - {address: "10.0.0.2:1260", listen: "127.0.0.2:1260"}
`))
	if err != nil {
		t.Fatalf("distinct bind hosts must not collide: %v", err)
	}
}

func TestPartnerReferences(t *testing.T) {
	_, err := Parse([]byte(`
external_systems: [remote]
clusters:
  - {name: a, password: pw, partner_system: b, rest_listen: ":1", nodes: [{address: "h:10"}]}
  - {name: b, password: pw, partner_system: remote, rest_listen: ":2", nodes: [{address: "h:11"}]}
`))
	if err != nil {
		t.Fatalf("cluster and external partner references must be valid: %v", err)
	}
}

// Every example under scenarios/ must stay loadable.
func TestExampleScenariosLoad(t *testing.T) {
	files, err := filepath.Glob("../../scenarios/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("no scenario files found: %v", err)
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			if _, err := Load(f); err != nil {
				t.Fatal(err)
			}
		})
	}
}
