// Command mock-flashsystem emulates one or more IBM FlashSystem clusters for
// testing the IP quorum tooling: REST API, mock quorum nodes and a test-only
// control API. See ipquorum-emulator/README.md.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olemyk/ipquorum-emulator/internal/cluster"
	"github.com/olemyk/ipquorum-emulator/internal/config"
	"github.com/olemyk/ipquorum-emulator/internal/control"
	"github.com/olemyk/ipquorum-emulator/internal/jar"
	"github.com/olemyk/ipquorum-emulator/internal/nodes"
	"github.com/olemyk/ipquorum-emulator/internal/rest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", envOr("EMULATOR_CONFIG", ""), "path to the YAML configuration (env: EMULATOR_CONFIG)")
	jarPath := flag.String("jar-template", envOr("FAKE_JAR_TEMPLATE", "fake-ip_quorum.jar"), "path to the fake ip_quorum.jar template (env: FAKE_JAR_TEMPLATE)")
	debug := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	if *configPath == "" {
		return errors.New("-config (or EMULATOR_CONFIG) is required")
	}

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	builder, err := jar.NewBuilderFromFile(*jarPath)
	if err != nil {
		return err
	}

	reg := cluster.NewRegistry(cfg)
	nodeSets := map[string]*nodes.Set{}
	var restServers []*rest.Server

	for _, cl := range reg.All() {
		set := nodes.NewSet(cl.Config(), log)
		if err := set.Start(); err != nil {
			return err
		}
		defer set.Close()
		nodeSets[cl.Name()] = set

		srv, err := rest.NewServer(cl, reg, builder, log)
		if err != nil {
			return err
		}
		if err := srv.Start(); err != nil {
			return err
		}
		restServers = append(restServers, srv)
		log.Info("cluster ready", "cluster", cl.Name(), "rest", srv.Addr(),
			"nodes", len(cl.Config().Nodes), "scenario", cl.Scenario(), "partner", cl.Config().PartnerSystem)
	}

	var ctl *http.Server
	if cfg.ControlListen != "" {
		ln, err := net.Listen("tcp", cfg.ControlListen)
		if err != nil {
			return fmt.Errorf("control API: %w", err)
		}
		ctl = &http.Server{
			Handler:           control.NewHandler(reg, nodeSets, log),
			ReadHeaderTimeout: 10 * time.Second,
		}
		go func() {
			if err := ctl.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error("control API stopped", "err", err)
			}
		}()
		log.Info("control API ready (test only, no authentication)", "listen", ln.Addr().String())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Info("shutting down")

	for _, s := range restServers {
		_ = s.Close()
	}
	if ctl != nil {
		_ = ctl.Close()
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
