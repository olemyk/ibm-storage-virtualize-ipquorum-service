package client

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/olemyk/ipquorum-go/internal/errors"
	"github.com/olemyk/ipquorum-go/pkg/config"
	"github.com/olemyk/ipquorum-go/pkg/logger"
)

// newTestClient starts an HTTPS test server with the given handler and returns
// a Client that talks to it through the normal endpoint + port configuration.
func newTestClient(t *testing.T, h http.Handler, mutate func(*config.Config)) (*Client, *config.Config) {
	t.Helper()
	ts := httptest.NewTLSServer(h)
	t.Cleanup(ts.Close)

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portStr)

	cfg := config.NewConfig()
	cfg.APIEndpoint = host
	cfg.APIPort = port
	cfg.Username = "superuser"
	cfg.Password = "secret"
	cfg.PartnerSystem = "cluster02"
	cfg.OutputFile = filepath.Join(t.TempDir(), "ip_quorum.jar")
	cfg.BaseDelay = 0
	cfg.MaxRetries = 3
	if mutate != nil {
		mutate(cfg)
	}
	return NewClient(cfg, logger.New(false)), cfg
}

// authedClient returns a client that already holds a token.
func authedClient(t *testing.T, h http.Handler, mutate func(*config.Config)) (*Client, *config.Config) {
	t.Helper()
	c, cfg := newTestClient(t, h, mutate)
	c.token = "tok"
	return c, cfg
}

func TestNewClientBaseURLUsesPort(t *testing.T) {
	cfg := config.NewConfig()
	cfg.APIEndpoint = "10.1.2.3"
	if got := NewClient(cfg, logger.New(false)).baseURL; got != "https://10.1.2.3:7443" {
		t.Errorf("default baseURL = %q", got)
	}
	cfg.APIPort = 8443
	if got := NewClient(cfg, logger.New(false)).baseURL; got != "https://10.1.2.3:8443" {
		t.Errorf("custom port baseURL = %q", got)
	}
}

func TestPreflightCheck(t *testing.T) {
	for _, status := range []int{200, 401, 404, 405, 500} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			var method string
			c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method = r.Method
				w.WriteHeader(status)
			}), nil)
			if err := c.PreflightCheck(); err != nil {
				t.Errorf("PreflightCheck() with %d = %v, want nil", status, err)
			}
			if method != "GET" {
				t.Errorf("method = %s, want GET", method)
			}
		})
	}
}

func TestPreflightCheckUnreachable(t *testing.T) {
	cfg := config.NewConfig()
	cfg.APIEndpoint = "127.0.0.1"
	cfg.APIPort = 1 // nothing listens here
	err := NewClient(cfg, logger.New(false)).PreflightCheck()
	if _, ok := err.(*errors.NetworkError); !ok {
		t.Errorf("err = %T %v, want *NetworkError", err, err)
	}
}

func TestPreflightCheckTLSVerification(t *testing.T) {
	// A self-signed server must be rejected when verification is on.
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), func(cfg *config.Config) {
		cfg.VerifySSL = true
	})
	err := c.PreflightCheck()
	if _, ok := err.(*errors.NetworkError); !ok {
		t.Errorf("err = %T %v, want *NetworkError for an untrusted certificate", err, err)
	}
}

func TestAuthenticateTokenSources(t *testing.T) {
	tests := []struct {
		name string
		resp func(w http.ResponseWriter)
		want string
	}{
		{"X-Auth-Token header", func(w http.ResponseWriter) {
			w.Header().Set("X-Auth-Token", "hdr-token")
			w.WriteHeader(201)
		}, "hdr-token"},
		{"JSON token", func(w http.ResponseWriter) {
			w.WriteHeader(201)
			io.WriteString(w, `{"token":"json-token"}`)
		}, "json-token"},
		{"JSON access_token", func(w http.ResponseWriter) {
			w.WriteHeader(200)
			io.WriteString(w, `{"access_token":"at"}`)
		}, "at"},
		{"JSON nested data.token", func(w http.ResponseWriter) {
			w.WriteHeader(200)
			io.WriteString(w, `{"data":{"token":"nested"}}`)
		}, "nested"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.resp(w)
			}), nil)
			if err := c.Authenticate(); err != nil {
				t.Fatal(err)
			}
			if c.GetToken() != tt.want {
				t.Errorf("token = %q, want %q", c.GetToken(), tt.want)
			}
		})
	}
}

func TestAuthenticateSendsCredentials(t *testing.T) {
	var gotUser, gotPass, gotMethod, gotPath string
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass = r.Header.Get("X-Auth-Username"), r.Header.Get("X-Auth-Password")
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("X-Auth-Token", "t")
		w.WriteHeader(201)
	}), nil)
	if err := c.Authenticate(); err != nil {
		t.Fatal(err)
	}
	if gotUser != "superuser" || gotPass != "secret" || gotMethod != "POST" || gotPath != "/rest/v1/auth" {
		t.Errorf("request = %s %s user=%q pass=%q", gotMethod, gotPath, gotUser, gotPass)
	}
}

func TestAuthenticateFailsFast(t *testing.T) {
	for _, status := range []int{401, 403} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			var calls int32
			c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&calls, 1)
				w.WriteHeader(status)
			}), nil)
			err := c.Authenticate()
			ae, ok := err.(*errors.AuthenticationError)
			if !ok {
				t.Fatalf("err = %T %v, want *AuthenticationError", err, err)
			}
			if ae.StatusCode != status {
				t.Errorf("StatusCode = %d, want %d", ae.StatusCode, status)
			}
			if n := atomic.LoadInt32(&calls); n != 1 {
				t.Errorf("made %d attempts, want exactly 1 (fail fast)", n)
			}
		})
	}
}

func TestAuthenticateRetriesOnRateLimit(t *testing.T) {
	var calls int32
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			w.Header().Set("Retry-After", "0") // keep the test fast
			w.WriteHeader(429)
			return
		}
		w.Header().Set("X-Auth-Token", "after-retry")
		w.WriteHeader(201)
	}), nil)
	if err := c.Authenticate(); err != nil {
		t.Fatal(err)
	}
	if c.GetToken() != "after-retry" || atomic.LoadInt32(&calls) != 3 {
		t.Errorf("token=%q calls=%d, want after-retry and 3", c.GetToken(), calls)
	}
}

func TestAuthenticateGivesUpAfterMaxRetries(t *testing.T) {
	var calls int32
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(429)
	}), func(cfg *config.Config) { cfg.MaxRetries = 2 })

	err := c.Authenticate()
	if _, ok := err.(*errors.AuthenticationError); !ok {
		t.Fatalf("err = %T %v, want *AuthenticationError", err, err)
	}
	if n := atomic.LoadInt32(&calls); n != 2 {
		t.Errorf("made %d attempts, want MaxRetries=2", n)
	}
}

func TestAuthenticateNoTokenInSuccessResponse(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		io.WriteString(w, `{}`)
	}), func(cfg *config.Config) { cfg.MaxRetries = 1 })
	if err := c.Authenticate(); err == nil {
		t.Fatal("expected an error when no token is returned")
	}
	if c.GetToken() != "" {
		t.Errorf("token = %q, want empty", c.GetToken())
	}
}

func TestCreateQuorumApp(t *testing.T) {
	t.Run("sends the payload", func(t *testing.T) {
		var got QuorumAppPayload
		var token, path string
		c, _ := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, path = r.Header.Get("X-Auth-Token"), r.URL.Path
			json.NewDecoder(r.Body).Decode(&got)
			w.WriteHeader(201)
		}), func(cfg *config.Config) {
			cfg.IP6, cfg.NoMetadata, cfg.PartnerIP6 = true, true, true
		})
		if err := c.CreateQuorumApp(); err != nil {
			t.Fatal(err)
		}
		want := QuorumAppPayload{IP6: true, NoMetadata: true, PartnerSystem: "cluster02", PartnerIP6: true}
		if got != want || token != "tok" || path != "/rest/v1/mkquorumapp" {
			t.Errorf("payload=%+v token=%q path=%q", got, token, path)
		}
	})

	t.Run("accepts 200", func(t *testing.T) {
		c, _ := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }), nil)
		if err := c.CreateQuorumApp(); err != nil {
			t.Errorf("err = %v", err)
		}
	})

	for _, status := range []int{400, 403, 500} {
		t.Run("error "+strconv.Itoa(status), func(t *testing.T) {
			c, _ := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }), nil)
			err := c.CreateQuorumApp()
			ae, ok := err.(*errors.APIError)
			if !ok || ae.StatusCode != status {
				t.Errorf("err = %T %v, want *APIError with %d", err, err, status)
			}
		})
	}

	t.Run("requires authentication", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("no request expected without a token")
		}), nil)
		if _, ok := c.CreateQuorumApp().(*errors.APIError); !ok {
			t.Error("expected an APIError when not authenticated")
		}
	})
}

func TestDownloadJAR(t *testing.T) {
	t.Run("writes the file", func(t *testing.T) {
		var got DownloadPayload
		var path string
		c, cfg := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path = r.URL.Path
			json.NewDecoder(r.Body).Decode(&got)
			w.WriteHeader(200)
			io.WriteString(w, "PK-fake-jar-bytes")
		}), nil)
		if err := c.DownloadJAR(); err != nil {
			t.Fatal(err)
		}
		if got != (DownloadPayload{Prefix: "/dumps", Filename: "ip_quorum.jar"}) || path != "/rest/v1/download" {
			t.Errorf("payload=%+v path=%q", got, path)
		}
		data, err := readFile(cfg.OutputFile)
		if err != nil || data != "PK-fake-jar-bytes" {
			t.Errorf("file = %q, err %v", data, err)
		}
	})

	t.Run("empty body is a warning, not an error", func(t *testing.T) {
		c, cfg := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }), nil)
		if err := c.DownloadJAR(); err != nil {
			t.Fatal(err)
		}
		if data, _ := readFile(cfg.OutputFile); data != "" {
			t.Errorf("file = %q, want empty", data)
		}
	})

	for _, status := range []int{403, 404, 500} {
		t.Run("error "+strconv.Itoa(status), func(t *testing.T) {
			c, cfg := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }), nil)
			err := c.DownloadJAR()
			ae, ok := err.(*errors.APIError)
			if !ok || ae.StatusCode != status {
				t.Errorf("err = %T %v, want *APIError with %d", err, err, status)
			}
			if _, statErr := readFile(cfg.OutputFile); statErr == nil {
				t.Error("no output file must be created on a failed download")
			}
		})
	}

	t.Run("unwritable output path", func(t *testing.T) {
		c, _ := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			io.WriteString(w, "x")
		}), func(cfg *config.Config) { cfg.OutputFile = "/nonexistent-dir/ip_quorum.jar" })
		if _, ok := c.DownloadJAR().(*errors.APIError); !ok {
			t.Error("expected an APIError when the output file cannot be created")
		}
	})

	t.Run("requires authentication", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("no request expected without a token")
		}), nil)
		if _, ok := c.DownloadJAR().(*errors.APIError); !ok {
			t.Error("expected an APIError when not authenticated")
		}
	})
}

// Two systems must be fully independent: each client talks only to its own
// system and receives that system's JAR.
func TestTwoSystemsAreIndependent(t *testing.T) {
	system := func(name string) http.Handler {
		mux := http.NewServeMux()
		mux.HandleFunc("/rest/v1/auth", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Auth-Token", "token-of-"+name)
			w.WriteHeader(201)
		})
		mux.HandleFunc("/rest/v1/mkquorumapp", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Auth-Token") != "token-of-"+name {
				w.WriteHeader(403)
				return
			}
			w.WriteHeader(201)
		})
		mux.HandleFunc("/rest/v1/download", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Auth-Token") != "token-of-"+name {
				w.WriteHeader(403)
				return
			}
			io.WriteString(w, "jar-of-"+name)
		})
		return mux
	}

	var results []string
	for _, name := range []string{"flash-a", "flash-b"} {
		c, cfg := newTestClient(t, system(name), nil)
		if err := c.Authenticate(); err != nil {
			t.Fatal(err)
		}
		if err := c.CreateQuorumApp(); err != nil {
			t.Fatal(err)
		}
		if err := c.DownloadJAR(); err != nil {
			t.Fatal(err)
		}
		data, _ := readFile(cfg.OutputFile)
		results = append(results, data)
	}
	if results[0] != "jar-of-flash-a" || results[1] != "jar-of-flash-b" {
		t.Errorf("jars = %v", results)
	}
}

func TestDownloadTimeoutIsSeparateFromAPITimeout(t *testing.T) {
	// A body slower than the 30s API timeout must still be accepted; here we
	// only verify that a slow-but-short body completes.
	c, cfg := authedClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		time.Sleep(200 * time.Millisecond)
		io.WriteString(w, "slow-jar")
	}), nil)
	if err := c.DownloadJAR(); err != nil {
		t.Fatal(err)
	}
	if data, _ := readFile(cfg.OutputFile); data != "slow-jar" {
		t.Errorf("file = %q", data)
	}
}
