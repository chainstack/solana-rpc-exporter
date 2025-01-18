package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	// Labels
	StatusLabel  = "status"
	VersionLabel = "version"
	NetworkLabel = "network"
)

const (
	DefaultCacheValidity = 60 * time.Second
)

type cachedVersion struct {
	value     string
	timestamp time.Time
}

type cachedHealth struct {
	isHealthy      int
	numSlotsBehind int64
	timestamp      time.Time
}

type SolanaCollector struct {
	rpcClient *rpc.Client
	logger    *zap.SugaredLogger
	config    *ExporterConfig

	// Cache fields
	cacheMutex    sync.RWMutex
	versionCache  *cachedVersion
	healthCache   *cachedHealth
	cacheValidity time.Duration

	/// descriptors:
	NodeVersion             *GaugeDesc
	NodeHealth              *GaugeDesc
	NodeTransactionCount    *GaugeDesc
	NodeNumSlotsBehind      *GaugeDesc
	NodeMinimumLedgerSlot   *GaugeDesc
	NodeFirstAvailableBlock *GaugeDesc
	NodeEpoch               *GaugeDesc
}

func NewSolanaCollector(client *rpc.Client, config *ExporterConfig) *SolanaCollector {
	collector := &SolanaCollector{
		rpcClient:     client,
		logger:        slog.Get(),
		config:        config,
		cacheValidity: DefaultCacheValidity,

		NodeVersion: NewGaugeDesc(
			"solana_node_version_info",
			"Version information of the RPC node",
			NetworkLabel, VersionLabel,
		),
		NodeHealth: NewGaugeDesc(
			"solana_node_health",
			"Health status of the RPC node",
			NetworkLabel,
		),
		NodeTransactionCount: NewGaugeDesc(
			"solana_node_transaction_count",
			"Total number of transactions processed by the RPC node",
			NetworkLabel,
		),
		NodeNumSlotsBehind: NewGaugeDesc(
			"solana_node_num_slots_behind",
			"Number of slots the RPC node is behind",
			NetworkLabel,
		),
		NodeMinimumLedgerSlot: NewGaugeDesc(
			"solana_node_minimum_ledger_slot",
			"Lowest slot that the RPC node has information about in its ledger",
			NetworkLabel,
		),
		NodeFirstAvailableBlock: NewGaugeDesc(
			"solana_node_first_available_block",
			"First available block in the RPC node's ledger",
			NetworkLabel,
		),
		NodeEpoch: NewGaugeDesc(
			"solana_network_epoch",
			"Current epoch number",
			NetworkLabel,
		),
	}
	return collector
}

func (c *SolanaCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.NodeVersion.Desc
	ch <- c.NodeHealth.Desc
	ch <- c.NodeTransactionCount.Desc
	ch <- c.NodeNumSlotsBehind.Desc
	ch <- c.NodeMinimumLedgerSlot.Desc
	ch <- c.NodeFirstAvailableBlock.Desc
	ch <- c.NodeEpoch.Desc
}

func (c *SolanaCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	totalStart := time.Now()
	var callStart time.Time
	var version string
	var numSlotsBehind int64

	// Check version cache
	c.cacheMutex.RLock()
	if c.versionCache != nil && time.Since(c.versionCache.timestamp) < c.cacheValidity {
		version = c.versionCache.value
		ch <- c.NodeVersion.MustNewConstMetric(1, c.config.NetworkName, version)
		c.logger.Debug("Version returned from cache")
		c.cacheMutex.RUnlock()
	} else {
		c.cacheMutex.RUnlock()
		callStart = time.Now()
		var err error
		version, err = c.rpcClient.GetVersion(ctx)
		c.logger.Debugf("GetVersion took: %v ms", time.Since(callStart).Milliseconds())
		if err == nil {
			ch <- c.NodeVersion.MustNewConstMetric(1, c.config.NetworkName, version)
			// Update cache
			c.cacheMutex.Lock()
			c.versionCache = &cachedVersion{
				value:     version,
				timestamp: time.Now(),
			}
			c.cacheMutex.Unlock()
		} else {
			c.logger.Errorw("Failed to collect version", "error", err)
			ch <- c.NodeVersion.MustNewConstMetric(0, c.config.NetworkName, "unknown")
			version = "unknown"
		}
	}

	// Check health cache
	c.cacheMutex.RLock()
	if c.healthCache != nil && time.Since(c.healthCache.timestamp) < c.cacheValidity {
		ch <- c.NodeHealth.MustNewConstMetric(float64(c.healthCache.isHealthy), c.config.NetworkName)
		ch <- c.NodeNumSlotsBehind.MustNewConstMetric(float64(c.healthCache.numSlotsBehind), c.config.NetworkName)
		numSlotsBehind = c.healthCache.numSlotsBehind
		c.logger.Debug("Health returned from cache")
		c.cacheMutex.RUnlock()
	} else {
		c.cacheMutex.RUnlock()
		callStart = time.Now()
		_, err := c.rpcClient.GetHealth(ctx)
		c.logger.Debugf("GetHealth took: %v ms", time.Since(callStart).Milliseconds())
		isHealthy := 1
		if err != nil {
			isHealthy = 0
			var rpcError *rpc.RPCError
			if errors.As(err, &rpcError) {
				var errorData rpc.NodeUnhealthyErrorData
				if rpcError.Data != nil && rpc.UnpackRpcErrorData(rpcError, &errorData) == nil {
					numSlotsBehind = errorData.NumSlotsBehind
				}
			}
		}
		ch <- c.NodeHealth.MustNewConstMetric(float64(isHealthy), c.config.NetworkName)
		ch <- c.NodeNumSlotsBehind.MustNewConstMetric(float64(numSlotsBehind), c.config.NetworkName)

		// Update cache
		c.cacheMutex.Lock()
		c.healthCache = &cachedHealth{
			isHealthy:      isHealthy,
			numSlotsBehind: numSlotsBehind,
			timestamp:      time.Now(),
		}
		c.cacheMutex.Unlock()
	}

	// Collect minimum ledger slot
	callStart = time.Now()
	slot, err := c.rpcClient.GetMinimumLedgerSlot(ctx)
	c.logger.Debugf("GetMinimumLedgerSlot took: %v ms", time.Since(callStart).Milliseconds())
	if err == nil {
		ch <- c.NodeMinimumLedgerSlot.MustNewConstMetric(float64(slot), c.config.NetworkName)
	} else {
		c.logger.Errorw("Failed to get minimum ledger slot", "error", err)
		ch <- c.NodeMinimumLedgerSlot.MustNewConstMetric(0, c.config.NetworkName)
	}

	// Collect first available block
	callStart = time.Now()
	block, err := c.rpcClient.GetFirstAvailableBlock(ctx)
	c.logger.Debugf("GetFirstAvailableBlock took: %v ms", time.Since(callStart).Milliseconds())
	if err == nil {
		ch <- c.NodeFirstAvailableBlock.MustNewConstMetric(float64(block), c.config.NetworkName)
	} else {
		c.logger.Errorw("Failed to get first available block", "error", err)
		ch <- c.NodeFirstAvailableBlock.MustNewConstMetric(0, c.config.NetworkName)
	}

	// Collect transaction count from epoch info
	callStart = time.Now()
	epochInfo, err := c.rpcClient.GetEpochInfo(ctx, rpc.CommitmentConfirmed)
	c.logger.Debugf("GetEpochInfo took: %v ms", time.Since(callStart).Milliseconds())
	// Collect network epoch
	ch <- c.NodeEpoch.MustNewConstMetric(float64(epochInfo.Epoch), c.config.NetworkName)

	if err == nil {
		ch <- c.NodeTransactionCount.MustNewConstMetric(float64(epochInfo.TransactionCount), c.config.NetworkName)

		c.logger.Infow("Successfully collected metrics",
			"slot_height", epochInfo.AbsoluteSlot,
			"block_height", epochInfo.BlockHeight,
			"epoch", epochInfo.Epoch,
			"slots_behind", numSlotsBehind,
			"min_ledger_slot", slot,
			"first_available_block", block,
			"version", version,
			"block_processing_time", time.Since(callStart).Seconds(),
			"total_duration_ms", time.Since(totalStart).Milliseconds(),
		)
	} else {
		c.logger.Errorw("Failed to collect metrics", "error", err)
		ch <- c.NodeTransactionCount.MustNewConstMetric(0, c.config.NetworkName)
	}
}
