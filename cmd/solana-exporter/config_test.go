package main

import (
	"context"
	"flag"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewExporterConfig(t *testing.T) {
	tests := []struct {
		name          string
		httpTimeout   time.Duration
		rpcUrl        string
		listenAddress string
		slotPace      time.Duration
		networkName   string
		debug         bool
		wantErr       bool
	}{
		{
			name:          "valid configuration",
			httpTimeout:   60 * time.Second,
			rpcUrl:        "http://localhost:8899",
			listenAddress: ":8080",
			slotPace:      time.Second,
			networkName:   "mainnet-beta",
			debug:         false,
			wantErr:       false,
		},
		{
			name:          "valid configuration with debug",
			httpTimeout:   30 * time.Second,
			rpcUrl:        "https://api.mainnet-beta.solana.com",
			listenAddress: ":9090",
			slotPace:      2 * time.Second,
			networkName:   "testnet",
			debug:         true,
			wantErr:       false,
		},
		{
			name:          "minimum valid configuration",
			httpTimeout:   1 * time.Second,
			rpcUrl:        "http://127.0.0.1:8899",
			listenAddress: "localhost:8080",
			slotPace:      500 * time.Millisecond,
			networkName:   "devnet",
			debug:         false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewExporterConfig(
				context.Background(),
				tt.httpTimeout,
				tt.rpcUrl,
				tt.listenAddress,
				tt.slotPace,
				tt.networkName,
				tt.debug,
			)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.httpTimeout, config.HttpTimeout)
			assert.Equal(t, tt.rpcUrl, config.RpcUrl)
			assert.Equal(t, tt.listenAddress, config.ListenAddress)
			assert.Equal(t, tt.slotPace, config.SlotPace)
			assert.Equal(t, tt.networkName, config.NetworkName)
			assert.Equal(t, tt.debug, config.Debug)
		})
	}
}

func parseFlagsForTest(args []string) (*ExporterConfig, error) {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	var (
		httpTimeout   int
		rpcUrl        string
		listenAddress string
		slotPace      int
		networkName   string
		debug         bool
	)

	flagSet.IntVar(&httpTimeout, "http-timeout", 60, "HTTP timeout in seconds")
	flagSet.StringVar(&rpcUrl, "rpc-url", "http://localhost:8899", "Solana RPC URL")
	flagSet.StringVar(&listenAddress, "listen-address", ":8080", "Listen address")
	flagSet.IntVar(&slotPace, "slot-pace", 1, "Slot pace in seconds")
	flagSet.StringVar(&networkName, "network", "mainnet-beta", "Network name")
	flagSet.BoolVar(&debug, "debug", false, "Enable debug mode")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	return NewExporterConfig(
		context.Background(),
		time.Duration(httpTimeout)*time.Second,
		rpcUrl,
		listenAddress,
		time.Duration(slotPace)*time.Second,
		networkName,
		debug,
	)
}

func TestNewExporterConfigFromCLI(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError bool
		validate      func(*testing.T, *ExporterConfig)
	}{
		{
			name: "default values",
			args: []string{},
			validate: func(t *testing.T, config *ExporterConfig) {
				assert.Equal(t, 60*time.Second, config.HttpTimeout)
				assert.Equal(t, "http://localhost:8899", config.RpcUrl)
				assert.Equal(t, ":8080", config.ListenAddress)
				assert.Equal(t, time.Second, config.SlotPace)
				assert.Equal(t, "mainnet-beta", config.NetworkName)
				assert.False(t, config.Debug)
			},
		},
		{
			name: "custom values",
			args: []string{
				"-http-timeout", "30",
				"-rpc-url", "https://api.testnet.solana.com",
				"-listen-address", ":9090",
				"-slot-pace", "2",
				"-network", "testnet",
				"-debug",
			},
			validate: func(t *testing.T, config *ExporterConfig) {
				assert.Equal(t, 30*time.Second, config.HttpTimeout)
				assert.Equal(t, "https://api.testnet.solana.com", config.RpcUrl)
				assert.Equal(t, ":9090", config.ListenAddress)
				assert.Equal(t, 2*time.Second, config.SlotPace)
				assert.Equal(t, "testnet", config.NetworkName)
				assert.True(t, config.Debug)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parseFlagsForTest(tt.args)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			tt.validate(t, config)
		})
	}
}
