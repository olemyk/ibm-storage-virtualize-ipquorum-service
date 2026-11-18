package rest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/cluster"
	"github.com/olemyk/ipquorum-emulator/internal/config"
)

type stubJARs struct{}

func (stubJARs) Build(cl config.Cluster) ([]byte, error) {
	return []byte("JAR:" + cl.Name), nil
}

func newRegistry(t *testing.T, yaml string) *cluster.Registry {
	t.Helper()
	cfg, err := config.Parse([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	return cluster.NewRegistry(cfg)
}

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// serve starts an HTTPS test server for the named cluster.
func serve(t *testing.T, reg *cluster.Registry, name string) (*httptest.Server, *cluster.Cluster) {
	t.Helper()
	cl, ok := reg.Get(name)
	if !ok {
		t.Fatalf("no cluster %s", name)
	}
	ts := httptest.NewTLSServer(NewHandler(cl, reg, stubJARs{}, discard()))
	t.Cleanup(ts.Close)
	return ts, cl
}

func do(t *testing.T, ts *httptest.Server, method, path, token string, headers map[string]string, body any) (*http.Response, []byte) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, ts.URL+path, rd)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-Auth-Token", token)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := ts.Client()
	client.Transport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return resp, data
}

func login(t *testing.T, ts *httptest.Server, user, pass string) (int, string) {
	t.Helper()
	resp, body := do(t, ts, "POST", "/rest/v1/auth", "", map[string]string{
		"X-Auth-Username": user, "X-Auth-Password": pass,
	}, nil)
	var out struct{ Token string }
	_ = json.Unmarshal(body, &out)
	return resp.StatusCode, out.Token
}

const twoClusters = `
clusters:
  - name: flash-a
    password: pw-a
    rest_listen: ":8443"
    partner_system: flash-b
    nodes: [{address: "h:1"}]
  - name: flash-b
    password: pw-b
    rest_listen: ":8444"
    partner_system: flash-a
    nodes: [{address: "h:2"}]
`

func TestPreflightGetAuth(t *testing.T) {
	ts, _ := serve(t, newRegistry(t, twoClusters), "flash-a")
	resp, _ := do(t, ts, "GET", "/rest/v1/auth", "", nil, nil)
	if resp.StatusCode != 405 {
		t.Errorf("GET /auth = %d, want 405", resp.StatusCode)
	}
}

func TestAuth(t *testing.T) {
	ts, _ := serve(t, newRegistry(t, twoClusters), "flash-a")

	if code, tok := login(t, ts, "superuser", "pw-a"); code != 201 || tok == "" {
		t.Errorf("valid login = %d token=%q", code, tok)
	}
	if code, tok := login(t, ts, "superuser", "wrong"); code != 403 || tok != "" {
		t.Errorf("wrong password = %d token=%q, want 403 and no token", code, tok)
	}
	if code, _ := login(t, ts, "someone", "pw-a"); code != 403 {
		t.Errorf("wrong user = %d, want 403", code)
	}
}

func TestTokenScopedToCluster(t *testing.T) {
	reg := newRegistry(t, twoClusters)
	tsA, _ := serve(t, reg, "flash-a")
	tsB, _ := serve(t, reg, "flash-b")

	_, tokA := login(t, tsA, "superuser", "pw-a")
	resp, _ := do(t, tsB, "POST", "/rest/v1/mkquorumapp", tokA, nil, cluster.MkquorumappRequest{PartnerSystem: "flash-a"})
	if resp.StatusCode != 403 {
		t.Errorf("token of A used on B = %d, want 403", resp.StatusCode)
	}
	// A's credentials must not work on B either.
	if code, _ := login(t, tsB, "superuser", "pw-a"); code != 403 {
		t.Errorf("A credentials on B = %d, want 403", code)
	}
}

func TestMkquorumappThenDownload(t *testing.T) {
	reg := newRegistry(t, twoClusters)
	ts, cl := serve(t, reg, "flash-a")
	_, tok := login(t, ts, "superuser", "pw-a")
	dl := map[string]string{"prefix": "/dumps", "filename": "ip_quorum.jar"}

	// download before mkquorumapp: the file does not exist yet
	if resp, _ := do(t, ts, "POST", "/rest/v1/download", tok, nil, dl); resp.StatusCode != 404 {
		t.Errorf("download before mkquorumapp = %d, want 404", resp.StatusCode)
	}

	mk := cluster.MkquorumappRequest{PartnerSystem: "flash-b", IP6: true}
	if resp, _ := do(t, ts, "POST", "/rest/v1/mkquorumapp", tok, nil, mk); resp.StatusCode != 201 {
		t.Fatalf("mkquorumapp = %d, want 201", resp.StatusCode)
	}
	if s := cl.Snapshot(); !s.QuorumAppCreated || s.LastMkquorumapp == nil || !s.LastMkquorumapp.IP6 {
		t.Errorf("state after mkquorumapp: %+v", s)
	}

	resp, body := do(t, ts, "POST", "/rest/v1/download", tok, nil, dl)
	if resp.StatusCode != 200 || string(body) != "JAR:flash-a" {
		t.Errorf("download = %d %q", resp.StatusCode, body)
	}
	if cl.Snapshot().Downloads != 1 {
		t.Errorf("downloads = %d, want 1", cl.Snapshot().Downloads)
	}
}

func TestStatePerCluster(t *testing.T) {
	reg := newRegistry(t, twoClusters)
	tsA, clA := serve(t, reg, "flash-a")
	_, clB := serve(t, reg, "flash-b")
	_, tok := login(t, tsA, "superuser", "pw-a")
	do(t, tsA, "POST", "/rest/v1/mkquorumapp", tok, nil, cluster.MkquorumappRequest{PartnerSystem: "flash-b"})

	if !clA.QuorumAppCreated() || clB.QuorumAppCreated() {
		t.Errorf("quorum app state leaked between clusters: A=%v B=%v", clA.QuorumAppCreated(), clB.QuorumAppCreated())
	}
}

func TestPartnerSystemStrict(t *testing.T) {
	reg := newRegistry(t, `
external_systems: [remote]
clusters:
  - {name: flash-a, password: pw, partner_system: flash-b, rest_listen: ":8441", nodes: [{address: "h:1"}]}
  - {name: flash-b, password: pw, rest_listen: ":8442", nodes: [{address: "h:2"}]}
  - {name: flash-c, password: pw, rest_listen: ":8443", nodes: [{address: "h:3"}]}
`)
	tests := []struct {
		cluster, partner string
		want             int
	}{
		{"flash-a", "flash-b", 201}, // the configured partner
		{"flash-a", "flash-c", 400}, // exists, but is not this cluster's partner
		{"flash-a", "ghost", 400},   // does not exist
		{"flash-a", "flash-a", 400}, // itself
		{"flash-b", "flash-c", 201}, // no configured partner: any known cluster
		{"flash-b", "remote", 201},  // external system counts as known
		{"flash-b", "ghost", 400},
	}
	for _, tt := range tests {
		t.Run(tt.cluster+"->"+tt.partner, func(t *testing.T) {
			ts, _ := serve(t, reg, tt.cluster)
			_, tok := login(t, ts, "superuser", "pw")
			resp, _ := do(t, ts, "POST", "/rest/v1/mkquorumapp", tok, nil, cluster.MkquorumappRequest{PartnerSystem: tt.partner})
			if resp.StatusCode != tt.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestNoTokenRejected(t *testing.T) {
	ts, _ := serve(t, newRegistry(t, twoClusters), "flash-a")
	for _, p := range []string{"/rest/v1/mkquorumapp", "/rest/v1/download"} {
		if resp, _ := do(t, ts, "POST", p, "", nil, map[string]string{}); resp.StatusCode != 403 {
			t.Errorf("%s without token = %d, want 403", p, resp.StatusCode)
		}
		if resp, _ := do(t, ts, "POST", p, "bogus", nil, map[string]string{}); resp.StatusCode != 403 {
			t.Errorf("%s with bogus token = %d, want 403", p, resp.StatusCode)
		}
	}
}

func scenarioCluster(scenario string) string {
	return `
clusters:
  - name: flash-a
    password: pw
    scenario: ` + scenario + `
    download_delay: 300ms
    rate_limit_count: 2
    quorum_app_exists: true
    nodes: [{address: "h:1"}]
`
}

func TestScenarios(t *testing.T) {
	dl := map[string]string{"prefix": "/dumps", "filename": "ip_quorum.jar"}

	t.Run("auth-401", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("auth-401")), "flash-a")
		if code, _ := login(t, ts, "superuser", "pw"); code != 401 {
			t.Errorf("= %d, want 401", code)
		}
	})
	t.Run("auth-403", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("auth-403")), "flash-a")
		if code, _ := login(t, ts, "superuser", "pw"); code != 403 {
			t.Errorf("= %d, want 403", code)
		}
	})
	t.Run("rate-limit-429", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("rate-limit-429")), "flash-a")
		for i := 1; i <= 2; i++ {
			resp, _ := do(t, ts, "POST", "/rest/v1/auth", "", map[string]string{"X-Auth-Username": "superuser", "X-Auth-Password": "pw"}, nil)
			if resp.StatusCode != 429 || resp.Header.Get("Retry-After") == "" {
				t.Errorf("attempt %d = %d retry-after=%q, want 429 with Retry-After", i, resp.StatusCode, resp.Header.Get("Retry-After"))
			}
		}
		if code, tok := login(t, ts, "superuser", "pw"); code != 201 || tok == "" {
			t.Errorf("attempt 3 = %d, want success", code)
		}
	})
	t.Run("mkquorumapp-error", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("mkquorumapp-error")), "flash-a")
		_, tok := login(t, ts, "superuser", "pw")
		if resp, _ := do(t, ts, "POST", "/rest/v1/mkquorumapp", tok, nil, cluster.MkquorumappRequest{}); resp.StatusCode != 500 {
			t.Errorf("= %d, want 500", resp.StatusCode)
		}
	})
	t.Run("download-error", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("download-error")), "flash-a")
		_, tok := login(t, ts, "superuser", "pw")
		if resp, _ := do(t, ts, "POST", "/rest/v1/download", tok, nil, dl); resp.StatusCode != 500 {
			t.Errorf("= %d, want 500", resp.StatusCode)
		}
	})
	t.Run("download-empty", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("download-empty")), "flash-a")
		_, tok := login(t, ts, "superuser", "pw")
		resp, body := do(t, ts, "POST", "/rest/v1/download", tok, nil, dl)
		if resp.StatusCode != 200 || len(body) != 0 {
			t.Errorf("= %d with %d bytes, want 200 and empty body", resp.StatusCode, len(body))
		}
	})
	t.Run("download-slow", func(t *testing.T) {
		ts, _ := serve(t, newRegistry(t, scenarioCluster("download-slow")), "flash-a")
		_, tok := login(t, ts, "superuser", "pw")
		start := time.Now()
		resp, body := do(t, ts, "POST", "/rest/v1/download", tok, nil, dl)
		if resp.StatusCode != 200 || string(body) != "JAR:flash-a" {
			t.Errorf("= %d %q", resp.StatusCode, body)
		}
		if elapsed := time.Since(start); elapsed < 250*time.Millisecond {
			t.Errorf("download took %v, want at least the 300ms delay", elapsed)
		}
	})
	t.Run("switch scenario at runtime", func(t *testing.T) {
		ts, cl := serve(t, newRegistry(t, scenarioCluster("healthy")), "flash-a")
		if code, _ := login(t, ts, "superuser", "pw"); code != 201 {
			t.Fatalf("healthy login = %d", code)
		}
		cl.SetScenario(config.ScenarioAuth401)
		if code, _ := login(t, ts, "superuser", "pw"); code != 401 {
			t.Errorf("after switch = %d, want 401", code)
		}
	})
}

func TestGenerateCert(t *testing.T) {
	cert, err := GenerateCert([]string{"localhost", "10.0.0.1", "flash-a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cert.Certificate) != 1 {
		t.Fatalf("got %d certs", len(cert.Certificate))
	}
	if !strings.Contains(string(cert.Certificate[0]), "ipquorum-emulator") {
		t.Error("certificate subject missing")
	}
}
