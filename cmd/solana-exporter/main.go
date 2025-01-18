package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	slog.Init()
	logger := slog.Get()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Infof("Received signal %v, initiating shutdown...", sig)
		cancel()
	}()

	// Load configuration
	config, err := NewExporterConfigFromCLI(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	// Validate network name
	if !isValidNetwork(config.NetworkName) {
		logger.Fatalf("Invalid network name: %s. Must be one of: mainnet, testnet, devnet, localnet", config.NetworkName)
	}

	// Initialize RPC client
	client := rpc.NewRPCClient(config.RpcUrl, config.HttpTimeout)

	// Initialize collectors
	collector := NewSolanaCollector(client, config)
	slotWatcher := NewSlotWatcher(client, config)

	// Start slot watcher with infinite retry
	go func() {
		for {
			if err := slotWatcher.WatchSlots(ctx); err != nil && err != context.Canceled {
				logger.Warnf("Slot watcher error: %v, retrying in 5 seconds...", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}
		}
	}()

	// Register collector (continue even if registration fails)
	if err := prometheus.Register(collector); err != nil {
		logger.Warnf("Failed to register collector: %v, continuing anyway", err)
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	})

	server := &http.Server{
		Addr:    config.ListenAddress,
		Handler: mux,
	}

	// Start server
	go func() {
		logger.Infof("Starting metrics server on %s", config.ListenAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Failed to start metrics server: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("Shutting down...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Error during server shutdown: %v", err)
	}

	logger.Info("Exporter stopped")
}
