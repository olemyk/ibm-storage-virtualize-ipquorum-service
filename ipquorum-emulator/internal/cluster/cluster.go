// Package cluster holds the runtime state of the emulated FlashSystem
// clusters: scenario, auth tokens and whether the quorum app was created.
// All state is per cluster and safe for concurrent use.
package cluster

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/olemyk/ipquorum-emulator/internal/config"
)

// MkquorumappRequest is the payload of POST /rest/v1/mkquorumapp.
type MkquorumappRequest struct {
	IP6           bool   `json:"ip_6"`
	NoMetadata    bool   `json:"nometadata"`
	PartnerSystem string `json:"partnersystem"`
	PartnerIP6    bool   `json:"partnerip6"`
}

// Snapshot is a point-in-time copy of a cluster's mutable state.
type Snapshot struct {
	Name             string              `json:"name"`
	Scenario         config.Scenario     `json:"scenario"`
	PartnerSystem    string              `json:"partner_system,omitempty"`
	QuorumAppCreated bool                `json:"quorum_app_created"`
	AuthAttempts     int                 `json:"auth_attempts"`
	ActiveTokens     int                 `json:"active_tokens"`
	Downloads        int                 `json:"downloads"`
	LastMkquorumapp  *MkquorumappRequest `json:"last_mkquorumapp,omitempty"`
}

// Cluster is one emulated FlashSystem.
type Cluster struct {
	cfg config.Cluster

	mu           sync.Mutex
	scenario     config.Scenario
	authAttempts int
	tokens       map[string]struct{}
	quorumApp    bool
	downloads    int
	lastMk       *MkquorumappRequest
}

// New creates a cluster in the state described by cfg.
func New(cfg config.Cluster) *Cluster {
	return &Cluster{
		cfg:       cfg,
		scenario:  cfg.Scenario,
		tokens:    map[string]struct{}{},
		quorumApp: cfg.QuorumAppExists,
	}
}

// Config returns the static configuration of the cluster.
func (c *Cluster) Config() config.Cluster { return c.cfg }

// Name returns the cluster name.
func (c *Cluster) Name() string { return c.cfg.Name }

// Scenario returns the current REST scenario.
func (c *Cluster) Scenario() config.Scenario {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.scenario
}

// SetScenario switches the REST scenario and restarts the auth attempt
// counter, so rate-limit-429 starts over.
func (c *Cluster) SetScenario(s config.Scenario) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scenario = s
	c.authAttempts = 0
}

// RecordAuthAttempt counts an auth request and returns its 1-based number.
func (c *Cluster) RecordAuthAttempt() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authAttempts++
	return c.authAttempts
}

// NewToken issues an auth token that is only valid for this cluster.
func (c *Cluster) NewToken() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		panic("cluster: crypto/rand failed: " + err.Error())
	}
	token := hex.EncodeToString(buf)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[token] = struct{}{}
	return token
}

// ValidToken reports whether token was issued by this cluster.
func (c *Cluster) ValidToken(token string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.tokens[token]
	return token != "" && ok
}

// QuorumAppCreated reports whether ip_quorum.jar can be downloaded.
func (c *Cluster) QuorumAppCreated() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.quorumApp
}

// CreateQuorumApp records a successful mkquorumapp call.
func (c *Cluster) CreateQuorumApp(req MkquorumappRequest) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.quorumApp = true
	c.lastMk = &req
}

// RecordDownload counts a served ip_quorum.jar.
func (c *Cluster) RecordDownload() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.downloads++
}

// Reset returns the cluster to its configured initial state.
func (c *Cluster) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scenario = c.cfg.Scenario
	c.authAttempts = 0
	c.tokens = map[string]struct{}{}
	c.quorumApp = c.cfg.QuorumAppExists
	c.downloads = 0
	c.lastMk = nil
}

// Snapshot returns a copy of the current state.
func (c *Cluster) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := Snapshot{
		Name:             c.cfg.Name,
		Scenario:         c.scenario,
		PartnerSystem:    c.cfg.PartnerSystem,
		QuorumAppCreated: c.quorumApp,
		AuthAttempts:     c.authAttempts,
		ActiveTokens:     len(c.tokens),
		Downloads:        c.downloads,
	}
	if c.lastMk != nil {
		mk := *c.lastMk
		s.LastMkquorumapp = &mk
	}
	return s
}

// Registry is the set of clusters of one emulator process.
type Registry struct {
	clusters []*Cluster
	byName   map[string]*Cluster
	external map[string]bool
}

// NewRegistry builds a registry from a validated configuration.
func NewRegistry(cfg *config.Config) *Registry {
	r := &Registry{byName: map[string]*Cluster{}, external: map[string]bool{}}
	for _, name := range cfg.ExternalSystems {
		r.external[name] = true
	}
	for _, cc := range cfg.Clusters {
		cl := New(cc)
		r.clusters = append(r.clusters, cl)
		r.byName[cc.Name] = cl
	}
	return r
}

// All returns the clusters in configuration order.
func (r *Registry) All() []*Cluster { return r.clusters }

// Get returns a cluster by name.
func (r *Registry) Get(name string) (*Cluster, bool) {
	cl, ok := r.byName[name]
	return cl, ok
}

// KnownSystem reports whether name is an emulated cluster or an external system.
func (r *Registry) KnownSystem(name string) bool {
	_, ok := r.byName[name]
	return ok || r.external[name]
}
