// Package rest emulates the Storage Virtualize REST endpoints that
// ipquorum-download-go uses: auth, mkquorumapp and download.
package rest

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/cluster"
	"github.com/olemyk/ipquorum-emulator/internal/config"
)

// JARSource produces the ip_quorum.jar served by a cluster.
type JARSource interface {
	Build(cl config.Cluster) ([]byte, error)
}

// Handler serves the REST API of one cluster.
type Handler struct {
	cl   *cluster.Cluster
	reg  *cluster.Registry
	jars JARSource
	log  *slog.Logger
	mux  *http.ServeMux
}

// NewHandler returns the HTTP handler for one cluster.
func NewHandler(cl *cluster.Cluster, reg *cluster.Registry, jars JARSource, log *slog.Logger) *Handler {
	h := &Handler{
		cl:   cl,
		reg:  reg,
		jars: jars,
		log:  log.With("cluster", cl.Name()),
		mux:  http.NewServeMux(),
	}
	h.mux.HandleFunc("/rest/v1/auth", h.auth)
	h.mux.HandleFunc("/rest/v1/mkquorumapp", h.mkquorumapp)
	h.mux.HandleFunc("/rest/v1/download", h.download)
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
	h.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"code": status, "message": msg})
}

// requireToken checks X-Auth-Token. Tokens are scoped to this cluster.
func (h *Handler) requireToken(w http.ResponseWriter, r *http.Request) bool {
	if !h.cl.ValidToken(r.Header.Get("X-Auth-Token")) {
		h.log.Warn("rejected: missing or invalid token", "path", r.URL.Path)
		writeError(w, http.StatusForbidden, "CMMVC: invalid or missing authentication token")
		return false
	}
	return true
}

func (h *Handler) auth(w http.ResponseWriter, r *http.Request) {
	// The downloader pre-flight does GET /rest/v1/auth and accepts 405.
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	attempt := h.cl.RecordAuthAttempt()
	cfg := h.cl.Config()

	switch h.cl.Scenario() {
	case config.ScenarioAuth401:
		writeError(w, http.StatusUnauthorized, "authentication failed (scenario auth-401)")
		return
	case config.ScenarioAuth403:
		writeError(w, http.StatusForbidden, "insufficient permissions (scenario auth-403)")
		return
	case config.ScenarioRateLimit429:
		if attempt <= cfg.RateLimitCount {
			w.Header().Set("Retry-After", "1")
			writeError(w, http.StatusTooManyRequests, "too many requests (scenario rate-limit-429)")
			return
		}
	}

	if r.Header.Get("X-Auth-Username") != cfg.Username || r.Header.Get("X-Auth-Password") != cfg.Password {
		h.log.Warn("authentication failed: wrong credentials", "user", r.Header.Get("X-Auth-Username"))
		writeError(w, http.StatusForbidden, "CMMVC: authentication failed, wrong user name or password")
		return
	}

	token := h.cl.NewToken()
	h.log.Info("authenticated", "user", cfg.Username)
	writeJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *Handler) mkquorumapp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.requireToken(w, r) {
		return
	}
	if h.cl.Scenario() == config.ScenarioMkquorumappFail {
		writeError(w, http.StatusInternalServerError, "mkquorumapp failed (scenario mkquorumapp-error)")
		return
	}

	var req cluster.MkquorumappRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error())
		return
	}
	if err := h.checkPartner(req.PartnerSystem); err != nil {
		h.log.Warn("mkquorumapp rejected", "partnersystem", req.PartnerSystem, "reason", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.cl.CreateQuorumApp(req)
	h.log.Info("quorum app created", "partnersystem", req.PartnerSystem, "ip_6", req.IP6,
		"nometadata", req.NoMetadata, "partnerip6", req.PartnerIP6)
	writeJSON(w, http.StatusCreated, map[string]any{})
}

// checkPartner is strict, like a real system: a partner system must exist.
// If the cluster has a configured partner_system only that one is accepted.
func (h *Handler) checkPartner(partner string) error {
	if partner == "" {
		return nil
	}
	if partner == h.cl.Name() {
		return fmt.Errorf("CMMVC: partnersystem %q is the system itself", partner)
	}
	if configured := h.cl.Config().PartnerSystem; configured != "" {
		if partner != configured {
			return fmt.Errorf("CMMVC: partnersystem %q is not the partner of this system", partner)
		}
		return nil
	}
	if !h.reg.KnownSystem(partner) {
		return fmt.Errorf("CMMVC: partnersystem %q does not exist", partner)
	}
	return nil
}

type downloadRequest struct {
	Prefix   string `json:"prefix"`
	Filename string `json:"filename"`
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !h.requireToken(w, r) {
		return
	}

	scenario := h.cl.Scenario()
	if scenario == config.ScenarioDownloadError {
		writeError(w, http.StatusInternalServerError, "download failed (scenario download-error)")
		return
	}

	var req downloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload: "+err.Error())
		return
	}
	if req.Prefix != "/dumps" || req.Filename != "ip_quorum.jar" {
		writeError(w, http.StatusNotFound, fmt.Sprintf("file %s/%s not found", req.Prefix, req.Filename))
		return
	}
	if !h.cl.QuorumAppCreated() {
		h.log.Warn("download before mkquorumapp: ip_quorum.jar does not exist yet")
		writeError(w, http.StatusNotFound, "ip_quorum.jar does not exist, run mkquorumapp first")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if scenario == config.ScenarioDownloadEmpty {
		h.cl.RecordDownload()
		w.WriteHeader(http.StatusOK)
		return
	}

	jar, err := h.jars.Build(h.cl.Config())
	if err != nil {
		h.log.Error("building ip_quorum.jar failed", "err", err)
		writeError(w, http.StatusInternalServerError, "could not build ip_quorum.jar")
		return
	}

	w.WriteHeader(http.StatusOK)
	if scenario == config.ScenarioDownloadSlow {
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		select {
		case <-time.After(h.cl.Config().DownloadDelay):
		case <-r.Context().Done():
			return
		}
	}
	if _, err := w.Write(jar); err != nil {
		h.log.Warn("writing ip_quorum.jar failed", "err", err)
		return
	}
	h.cl.RecordDownload()
	h.log.Info("served ip_quorum.jar", "bytes", len(jar))
}

// Server is the HTTPS listener of one cluster.
type Server struct {
	srv  *http.Server
	ln   net.Listener
	name string
}

// NewServer creates the HTTPS server of a cluster. Call Start to listen.
func NewServer(cl *cluster.Cluster, reg *cluster.Registry, jars JARSource, log *slog.Logger) (*Server, error) {
	cfg := cl.Config()
	hosts := append([]string{"localhost", "127.0.0.1", "::1", cfg.Name}, cfg.TLSHosts...)
	cert, err := GenerateCert(hosts)
	if err != nil {
		return nil, fmt.Errorf("generate certificate for %s: %w", cfg.Name, err)
	}
	return &Server{
		name: cfg.Name,
		srv: &http.Server{
			Addr:              cfg.RESTListen,
			Handler:           NewHandler(cl, reg, jars, log),
			TLSConfig:         &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
			ReadHeaderTimeout: 10 * time.Second,
		},
	}, nil
}

// Start binds the listener and serves in the background.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("cluster %s: listen on %s: %w", s.name, s.srv.Addr, err)
	}
	s.ln = ln
	go func() {
		if err := s.srv.ServeTLS(ln, "", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("REST server stopped", "cluster", s.name, "err", err)
		}
	}()
	return nil
}

// Addr returns the bound address (valid after Start).
func (s *Server) Addr() string { return s.ln.Addr().String() }

// Close stops the server immediately.
func (s *Server) Close() error { return s.srv.Close() }
