package main

// ============================================================================
// STUDY GUIDE — PROCESS ENTRY POINT
//
// ROLE IN OM1:
//   This is where the Go executable begins. Read this file as lifecycle setup,
//   not as the agent's reasoning logic.
//
// MENTAL PATH:
//   CLI flags -> logger/metrics -> config.Load -> cancellation context
//   -> runtime.New -> Runtime.Run
//
// GO CONCEPTS TO NOTICE:
//   * package main + func main = executable entry point
//   * flag.String/Bool/Float64 return pointers; *configName dereferences one
//   * defer schedules cleanup for logger/metrics/context
//   * blank imports (_ ".../plugins/...") intentionally run plugin init()
//     functions so implementations register themselves before config loads them
//   * context cancellation propagates SIGINT/SIGTERM shutdown through OM1
//
// INTERVIEW VERSION:
//   "main loads configuration, creates process-wide services and cancellation,
//    constructs Runtime, then hands lifecycle ownership to Runtime.Run."
// ============================================================================

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/openmind/om1/internal/config"
	"github.com/openmind/om1/internal/logger"
	"github.com/openmind/om1/internal/metrics"
	"github.com/openmind/om1/internal/runtime"

	_ "github.com/openmind/om1/plugins/actions"
	_ "github.com/openmind/om1/plugins/backgrounds"
	_ "github.com/openmind/om1/plugins/inputs"
	_ "github.com/openmind/om1/plugins/llm"
)

// main performs process-level assembly. It does NOT run the Cortex reasoning loop itself.
// Think of it as the launch sequence that gets OM1 ready and then calls Runtime.Run.
func main() {
	var (
		configName = flag.String("config", "", "config name or path (required)")
		logLevel   = flag.String("log-level", cmp.Or(os.Getenv("LOG_LEVEL"), "info"), "log level: debug|info|warn|error (env: LOG_LEVEL)")
		hotReload  = flag.Bool("hot-reload", false, "reload config on file change")
		checkSecs  = flag.Float64("check-interval", 1.0, "hot-reload check interval (seconds)")
	)
	flag.Parse()

	if *configName == "" {
		fmt.Fprintln(os.Stderr, "error: --config is required")
		flag.Usage()
		os.Exit(1)
	}

	log := logger.BuildLogger(*logLevel)
	logger.Set(log)
	defer func() { _ = log.Sync() }()

	stopMetrics := metrics.StartServer(log)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = stopMetrics(shutdownCtx)
	}()

	// Convert the selected JSON5 configuration into typed Go configuration.
	// If this fails, none of the agent/orchestrator loops have started yet.
	cfg, err := config.Load(*configName)
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Construct the long-lived Runtime object. New builds it; Run below owns its lifecycle.
	rt := runtime.New(cfg, log, runtime.Options{
		HotReload:     *hotReload,
		CheckInterval: *checkSecs,
	})

	if err := rt.Run(ctx); err != nil && err != context.Canceled {
		log.Fatal("runtime exited with error", zap.Error(err))
	}
}
