package main

import (
	"context"
	"flag"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
	"time"
)

type ExporterConfig struct {
	HttpTimeout   time.Duration
	RpcUrl        string
	ListenAddress string
	SlotPace      time.Duration
	NetworkName   string
	Debug         bool
}

func NewExporterConfig(
	ctx context.Context,
	httpTimeout time.Duration,
	rpcUrl string,
	listenAddress string,
	slotPace time.Duration,
	networkName string,
	debug bool,
) (*ExporterConfig, error) {
	logger := slog.Get()
	logger.Infow(
		"Setting up exporter configuration",
		"httpTimeout", httpTimeout.Seconds(),
		"rpcUrl", rpcUrl,
		"listenAddress", listenAddress,
		"networkName", networkName,
		"slotPace", slotPace.Seconds(),
		"debug", debug,
	)

	config := ExporterConfig{
		HttpTimeout:   httpTimeout,
		RpcUrl:        rpcUrl,
		ListenAddress: listenAddress,
		SlotPace:      slotPace,
		NetworkName:   networkName,
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
		networkName   string
		debug         bool
	)

	flag.IntVar(
		&httpTimeout,
		"http-timeout",
		60,
		"HTTP timeout in seconds for RPC calls",
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
		"The address to listen on for HTTP requests",
	)
	flag.IntVar(
		&slotPace,
		"slot-pace",
		1,
		"Time between slot-watching metric collections in seconds",
	)
	flag.StringVar(
		&networkName,
		"network",
		"mainnet-beta",
		"Name of the Solana network (mainnet-beta, testnet, devnet, localnet)",
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
		networkName,
		debug,
	)
	if err != nil {
		return nil, err
	}
	return config, nil
}
