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

	// Keep trying to connect to RPC node
	go func() {
		for {
			if err := client.TestConnection(ctx); err != nil {
				logger.Warnf("Connection to RPC node failed: %v. Will retry in 5 seconds...", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}
			break
		}
		logger.Info("Successfully connected to RPC node")
	}()

	// Initialize collectors
	collector := NewSolanaCollector(client, config)
	slotWatcher := NewSlotWatcher(client, config)

	// Start slot watcher
	go func() {
		retryDelay := time.Second * 5
		for {
			if err := slotWatcher.WatchSlots(ctx); err != nil && err != context.Canceled {
				logger.Warnf("Slot watcher error: %v, retrying in %v...", err, retryDelay)
				select {
				case <-ctx.Done():
					return
				case <-time.After(retryDelay):
					// Exponential backoff with max 1 minute
					retryDelay = time.Duration(min(retryDelay.Seconds()*2, 60)) * time.Second
					continue
				}
			}
			retryDelay = time.Second * 5 // Reset delay on success
		}
	}()

	// Register collector with retry
	for {
		if err := prometheus.Register(collector); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
				// If already registered, that's fine
				break
			}
			logger.Warnf("Failed to register collector: %v, retrying in 5 seconds...", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
		break
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check RPC node health but don't fail exporter health check
		_, err := client.GetHealth(r.Context())
		if err != nil {
			logger.Warnf("RPC node health check failed: %v", err)
			w.WriteHeader(http.StatusOK) // Still return OK as exporter is running
			w.Write([]byte("exporter running, rpc node unhealthy"))
			return
		}
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

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Error during server shutdown: %v", err)
	}

	logger.Info("Exporter stopped")
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
