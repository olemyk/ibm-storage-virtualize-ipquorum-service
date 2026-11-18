// Package config defines the YAML model that describes the emulated
// FlashSystem clusters, their nodes and their failure scenarios.
package config

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Scenario selects how a cluster's REST API behaves.
type Scenario string

const (
	ScenarioHealthy         Scenario = "healthy"
	ScenarioAuth401         Scenario = "auth-401"          // every auth attempt returns 401
	ScenarioAuth403         Scenario = "auth-403"          // every auth attempt returns 403
	ScenarioRateLimit429    Scenario = "rate-limit-429"    // first rate_limit_count auth attempts return 429
	ScenarioMkquorumappFail Scenario = "mkquorumapp-error" // mkquorumapp returns 500
	ScenarioDownloadError   Scenario = "download-error"    // download returns 500
	ScenarioDownloadEmpty   Scenario = "download-empty"    // download returns 200 with an empty body
	ScenarioDownloadSlow    Scenario = "download-slow"     // download body is delayed by download_delay
)

// Scenarios lists every valid scenario.
var Scenarios = []Scenario{
	ScenarioHealthy, ScenarioAuth401, ScenarioAuth403, ScenarioRateLimit429,
	ScenarioMkquorumappFail, ScenarioDownloadError, ScenarioDownloadEmpty, ScenarioDownloadSlow,
}

// Valid reports whether s is a known scenario.
func (s Scenario) Valid() bool {
	for _, v := range Scenarios {
		if v == s {
			return true
		}
	}
	return false
}

// NodeBehavior selects how a mock quorum node (TCP listener) behaves.
type NodeBehavior string

const (
	BehaviorUp               NodeBehavior = "up"                // accepts and answers normally
	BehaviorDown             NodeBehavior = "down"              // not listening: connection refused
	BehaviorHandshakeTimeout NodeBehavior = "handshake-timeout" // accepts, never answers the handshake
	BehaviorDropAfter        NodeBehavior = "drop-after"        // answers normally, closes after drop_after
)

// Behaviors lists every valid node behavior.
var Behaviors = []NodeBehavior{BehaviorUp, BehaviorDown, BehaviorHandshakeTimeout, BehaviorDropAfter}

// Valid reports whether b is a known behavior.
func (b NodeBehavior) Valid() bool {
	for _, v := range Behaviors {
		if v == b {
			return true
		}
	}
	return false
}

// Default values applied by Load.
const (
	DefaultRESTListen     = ":7443"
	DefaultUsername       = "superuser"
	DefaultRateLimitCount = 2
	DefaultDownloadDelay  = 3 * time.Second
	DefaultDropAfter      = 10 * time.Second
)

// Config is the root of the emulator configuration file.
type Config struct {
	// ControlListen is the address of the test-only control API. Empty disables it.
	ControlListen string `yaml:"control_listen"`
	// ExternalSystems are names of systems that exist outside this config and
	// may be used as partner_system (for one-cluster-per-process setups).
	ExternalSystems []string  `yaml:"external_systems"`
	Clusters        []Cluster `yaml:"clusters"`
}

// Cluster describes one emulated FlashSystem.
type Cluster struct {
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	// RESTListen is where the HTTPS REST API listens (default ":7443").
	RESTListen string `yaml:"rest_listen"`
	// PartnerSystem is the only partnersystem value mkquorumapp accepts.
	PartnerSystem string `yaml:"partner_system"`
	// QuorumAppExists is the initial state: the ip_quorum.jar is already
	// available for download without calling mkquorumapp first.
	QuorumAppExists bool     `yaml:"quorum_app_exists"`
	Scenario        Scenario `yaml:"scenario"`
	// RateLimitCount is how many auth attempts get 429 in the rate-limit-429 scenario.
	RateLimitCount int `yaml:"rate_limit_count"`
	// DownloadDelay is the body delay in the download-slow scenario.
	DownloadDelay time.Duration `yaml:"download_delay"`
	// TLSHosts are extra DNS names/IPs for the generated self-signed certificate.
	TLSHosts []string `yaml:"tls_hosts"`
	Nodes    []Node   `yaml:"nodes"`
}

// Node describes one mock quorum node.
type Node struct {
	// Address is host:port as embedded in the generated ip_quorum.jar. This is
	// where the fake JAR connects.
	Address string `yaml:"address"`
	// Listen is where the mock binds. Defaults to ":<port of Address>".
	Listen    string        `yaml:"listen"`
	Behavior  NodeBehavior  `yaml:"behavior"`
	DropAfter time.Duration `yaml:"drop_after"`
}

// Load reads, defaults and validates a configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return Parse(data)
}

// Parse decodes, defaults and validates configuration YAML.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	for i := range c.Clusters {
		cl := &c.Clusters[i]
		if cl.Username == "" {
			cl.Username = DefaultUsername
		}
		if cl.RESTListen == "" {
			cl.RESTListen = DefaultRESTListen
		}
		if cl.Scenario == "" {
			cl.Scenario = ScenarioHealthy
		}
		if cl.RateLimitCount == 0 {
			cl.RateLimitCount = DefaultRateLimitCount
		}
		if cl.DownloadDelay == 0 {
			cl.DownloadDelay = DefaultDownloadDelay
		}
		for j := range cl.Nodes {
			n := &cl.Nodes[j]
			if n.Behavior == "" {
				n.Behavior = BehaviorUp
			}
			if n.DropAfter == 0 {
				n.DropAfter = DefaultDropAfter
			}
			if n.Listen == "" {
				if _, port, err := net.SplitHostPort(n.Address); err == nil {
					n.Listen = ":" + port
				}
			}
		}
	}
}

// listener is a bound address used for collision detection.
type listener struct {
	what string
	host string
	port int
}

func parseListener(what, addr string) (listener, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return listener{}, fmt.Errorf("%s: invalid address %q (want host:port or :port)", what, addr)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return listener{}, fmt.Errorf("%s: invalid port in %q", what, addr)
	}
	if host == "0.0.0.0" || host == "::" {
		host = ""
	}
	return listener{what: what, host: host, port: port}, nil
}

func (a listener) conflicts(b listener) bool {
	if a.port != b.port {
		return false
	}
	return a.host == "" || b.host == "" || a.host == b.host
}

// Validate checks the whole configuration.
func (c *Config) Validate() error {
	if len(c.Clusters) == 0 {
		return fmt.Errorf("config: at least one cluster is required")
	}

	known := map[string]bool{}
	for _, name := range c.ExternalSystems {
		if name == "" {
			return fmt.Errorf("config: external_systems contains an empty name")
		}
		known[name] = true
	}
	names := map[string]bool{}
	for _, cl := range c.Clusters {
		if cl.Name == "" {
			return fmt.Errorf("config: every cluster needs a name")
		}
		if names[cl.Name] {
			return fmt.Errorf("config: duplicate cluster name %q", cl.Name)
		}
		names[cl.Name] = true
		known[cl.Name] = true
	}

	var listeners []listener
	if c.ControlListen != "" {
		l, err := parseListener("control_listen", c.ControlListen)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		listeners = append(listeners, l)
	}

	for _, cl := range c.Clusters {
		if err := cl.validate(known); err != nil {
			return fmt.Errorf("config: cluster %q: %w", cl.Name, err)
		}
		l, err := parseListener(fmt.Sprintf("cluster %q rest_listen", cl.Name), cl.RESTListen)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		listeners = append(listeners, l)
		for i, n := range cl.Nodes {
			nl, err := parseListener(fmt.Sprintf("cluster %q node %d listen", cl.Name, i), n.Listen)
			if err != nil {
				return fmt.Errorf("config: %w", err)
			}
			listeners = append(listeners, nl)
		}
	}

	for i := range listeners {
		for j := i + 1; j < len(listeners); j++ {
			if listeners[i].conflicts(listeners[j]) {
				return fmt.Errorf("config: listener collision: %s and %s both use port %d",
					listeners[i].what, listeners[j].what, listeners[i].port)
			}
		}
	}
	return nil
}

func (cl Cluster) validate(known map[string]bool) error {
	if cl.Password == "" {
		return fmt.Errorf("password is required")
	}
	if !cl.Scenario.Valid() {
		return fmt.Errorf("unknown scenario %q (valid: %v)", cl.Scenario, Scenarios)
	}
	if cl.RateLimitCount < 0 {
		return fmt.Errorf("rate_limit_count must not be negative")
	}
	if cl.DownloadDelay < 0 {
		return fmt.Errorf("download_delay must not be negative")
	}
	if cl.PartnerSystem != "" {
		if cl.PartnerSystem == cl.Name {
			return fmt.Errorf("partner_system must not be the cluster itself")
		}
		if !known[cl.PartnerSystem] {
			return fmt.Errorf("partner_system %q is not a known cluster (add it to clusters or external_systems)", cl.PartnerSystem)
		}
	}
	if len(cl.Nodes) == 0 {
		return fmt.Errorf("at least one node is required")
	}
	seen := map[string]bool{}
	for i, n := range cl.Nodes {
		if _, err := parseListener(fmt.Sprintf("node %d address", i), n.Address); err != nil {
			return err
		}
		if host, _, _ := net.SplitHostPort(n.Address); host == "" {
			return fmt.Errorf("node %d address %q needs a host", i, n.Address)
		}
		if seen[n.Address] {
			return fmt.Errorf("duplicate node address %q", n.Address)
		}
		seen[n.Address] = true
		if !n.Behavior.Valid() {
			return fmt.Errorf("node %d: unknown behavior %q (valid: %v)", i, n.Behavior, Behaviors)
		}
		if n.DropAfter < 0 {
			return fmt.Errorf("node %d: drop_after must not be negative", i)
		}
	}
	return nil
}
