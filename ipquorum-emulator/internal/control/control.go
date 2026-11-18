// Package control is the test-only runtime control API of the emulator.
// It has no authentication and must never be exposed outside a test setup.
//
//	GET  /clusters                                    list clusters, state and nodes
//	GET  /clusters/{name}                             one cluster
//	POST /clusters/{name}/scenario                    {"scenario":"auth-401"}
//	POST /clusters/{name}/reset                       back to the configured state
//	GET  /clusters/{name}/nodes                       list nodes
//	POST /clusters/{name}/nodes/{address}/behavior    {"behavior":"down","drop_after":"5s"}
//	GET  /healthz
package control

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/cluster"
	"github.com/olemyk/ipquorum-emulator/internal/config"
	"github.com/olemyk/ipquorum-emulator/internal/nodes"
)

// Handler serves the control API.
type Handler struct {
	reg   *cluster.Registry
	nodes map[string]*nodes.Set
	log   *slog.Logger
	mux   *http.ServeMux
}

// ClusterView is the JSON representation of a cluster.
type ClusterView struct {
	cluster.Snapshot
	RESTListen string         `json:"rest_listen"`
	Nodes      []nodes.Status `json:"nodes"`
}

// NewHandler creates the control API. nodeSets is keyed by cluster name.
func NewHandler(reg *cluster.Registry, nodeSets map[string]*nodes.Set, log *slog.Logger) *Handler {
	h := &Handler{reg: reg, nodes: nodeSets, log: log, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	h.mux.HandleFunc("GET /clusters", h.list)
	h.mux.HandleFunc("GET /clusters/{name}", h.get)
	h.mux.HandleFunc("POST /clusters/{name}/scenario", h.setScenario)
	h.mux.HandleFunc("POST /clusters/{name}/reset", h.reset)
	h.mux.HandleFunc("GET /clusters/{name}/nodes", h.listNodes)
	h.mux.HandleFunc("POST /clusters/{name}/nodes/{address}/behavior", h.setBehavior)
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Info("control request", "method", r.Method, "path", r.URL.Path)
	h.mux.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) view(cl *cluster.Cluster) ClusterView {
	v := ClusterView{Snapshot: cl.Snapshot(), RESTListen: cl.Config().RESTListen, Nodes: []nodes.Status{}}
	if set, ok := h.nodes[cl.Name()]; ok {
		for _, n := range set.Nodes() {
			v.Nodes = append(v.Nodes, n.Status())
		}
	}
	return v
}

func (h *Handler) lookup(w http.ResponseWriter, r *http.Request) (*cluster.Cluster, bool) {
	name := r.PathValue("name")
	cl, ok := h.reg.Get(name)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown cluster "+name)
	}
	return cl, ok
}

func (h *Handler) list(w http.ResponseWriter, _ *http.Request) {
	out := []ClusterView{}
	for _, cl := range h.reg.All() {
		out = append(out, h.view(cl))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	if cl, ok := h.lookup(w, r); ok {
		writeJSON(w, http.StatusOK, h.view(cl))
	}
}

func (h *Handler) setScenario(w http.ResponseWriter, r *http.Request) {
	cl, ok := h.lookup(w, r)
	if !ok {
		return
	}
	var req struct {
		Scenario config.Scenario `json:"scenario"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if !req.Scenario.Valid() {
		writeError(w, http.StatusBadRequest, "unknown scenario "+string(req.Scenario))
		return
	}
	cl.SetScenario(req.Scenario)
	h.log.Info("scenario changed", "cluster", cl.Name(), "scenario", req.Scenario)
	writeJSON(w, http.StatusOK, h.view(cl))
}

func (h *Handler) reset(w http.ResponseWriter, r *http.Request) {
	cl, ok := h.lookup(w, r)
	if !ok {
		return
	}
	cl.Reset()
	if set, ok := h.nodes[cl.Name()]; ok {
		for i, n := range set.Nodes() {
			nc := cl.Config().Nodes[i]
			if err := n.SetBehavior(nc.Behavior, nc.DropAfter); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
	}
	h.log.Info("cluster reset", "cluster", cl.Name())
	writeJSON(w, http.StatusOK, h.view(cl))
}

func (h *Handler) listNodes(w http.ResponseWriter, r *http.Request) {
	cl, ok := h.lookup(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, h.view(cl).Nodes)
}

func (h *Handler) setBehavior(w http.ResponseWriter, r *http.Request) {
	cl, ok := h.lookup(w, r)
	if !ok {
		return
	}
	set, ok := h.nodes[cl.Name()]
	if !ok {
		writeError(w, http.StatusNotFound, "cluster has no nodes")
		return
	}
	// The address is host:port; allow a trailing-slash-free path value only.
	address := strings.TrimSpace(r.PathValue("address"))
	n, ok := set.Get(address)
	if !ok {
		writeError(w, http.StatusNotFound, "unknown node "+address)
		return
	}
	var req struct {
		Behavior  config.NodeBehavior `json:"behavior"`
		DropAfter string              `json:"drop_after"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	var dropAfter time.Duration
	if req.DropAfter != "" {
		d, err := time.ParseDuration(req.DropAfter)
		if err != nil || d < 0 {
			writeError(w, http.StatusBadRequest, "invalid drop_after "+req.DropAfter)
			return
		}
		dropAfter = d
	}
	if err := n.SetBehavior(req.Behavior, dropAfter); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.log.Info("node behavior changed", "cluster", cl.Name(), "node", address, "behavior", req.Behavior)
	writeJSON(w, http.StatusOK, n.Status())
}
