package main

import (
	"context"
	"flag"
	"time"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
)

type ExporterConfig struct {
	HttpTimeout   time.Duration
	RpcUrl        string
	ListenAddress string
	SlotPace      time.Duration
	ClusterName   string
	Debug         bool
}

func NewExporterConfig(
	ctx context.Context,
	httpTimeout time.Duration,
	rpcUrl string,
	listenAddress string,
	slotPace time.Duration,
	clusterName string,
	debug bool,
) (*ExporterConfig, error) {
	logger := slog.Get()
	logger.Infow(
		"Setting up export config with ",
		"httpTimeout", httpTimeout.Seconds(),
		"rpcUrl", rpcUrl,
		"listenAddress", listenAddress,
		"clusterName", clusterName,
		"debug", debug,
	)

	config := ExporterConfig{
		HttpTimeout:   httpTimeout,
		RpcUrl:        rpcUrl,
		ListenAddress: listenAddress,
		SlotPace:      slotPace,
		ClusterName:   clusterName,
		Debug:         debug,
	}
	return &config, nil
}

func NewExporterConfigFromCLI(ctx context.Context) (*ExporterConfig, error) {
	var (
		httpTimeout   int
		rpcUrl        string
		listenAddress string
		slotPace      int
		clusterName   string
		debug         bool
	)
	
	flag.IntVar(
		&httpTimeout,
		"http-timeout",
		60,
		"HTTP timeout to use, in seconds.",
	)
	flag.StringVar(
		&rpcUrl,
		"rpc-url",
		"http://localhost:8899",
		"Solana RPC URL (including protocol and path), "+
			"e.g., 'http://localhost:8899' or 'https://api.mainnet-beta.solana.com'",
	)
	flag.StringVar(
		&listenAddress,
		"listen-address",
		":8080",
		"Listen address",
	)
	flag.IntVar(
		&slotPace,
		"slot-pace",
		1,
		"Time between slot-watching metric collections, defaults to 1s.",
	)
	flag.StringVar(
		&clusterName,
		"cluster-name",
		"mainnet",
		"Name of the Solana cluster (mainnet, testnet, devnet)",
	)
	flag.BoolVar(
		&debug,
		"debug",
		false,
		"Enable debug logging",
	)
	flag.Parse()

	config, err := NewExporterConfig(
		ctx,
		time.Duration(httpTimeout)*time.Second,
		rpcUrl,
		listenAddress,
		time.Duration(slotPace)*time.Second,
		clusterName,
		debug,
	)
	if err != nil {
		return nil, err
	}
	return config, nil
}
