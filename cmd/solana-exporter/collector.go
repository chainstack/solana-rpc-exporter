package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/naviat/solana-exporter-rpc/pkg/rpc"
	"github.com/naviat/solana-exporter-rpc/pkg/slog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	// Labels
	StatusLabel      = "status"
	VersionLabel     = "version"
	ClusterLabel     = "cluster"
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
	cacheMutex     sync.RWMutex
	versionCache   *cachedVersion
	healthCache    *cachedHealth
	cacheValidity  time.Duration

	/// descriptors:
	NodeVersion             *GaugeDesc
	NodeHealth             *GaugeDesc
	NodeSlotHeight         *GaugeDesc
	NodeBlockHeight        *GaugeDesc
	NodeTransactionCount   *GaugeDesc
	NodeNumSlotsBehind     *GaugeDesc
	NodeBlockProcessingTime *GaugeDesc
	NodeMinimumLedgerSlot  *GaugeDesc
	NodeFirstAvailableBlock *GaugeDesc
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
			ClusterLabel, VersionLabel,
		),
		NodeHealth: NewGaugeDesc(
			"solana_node_health",
			"Health status of the RPC node",
			ClusterLabel,
		),
		NodeSlotHeight: NewGaugeDesc(
			"solana_node_slot_height",
			"Current slot height of the RPC node",
			ClusterLabel,
		),
		NodeBlockHeight: NewGaugeDesc(
			"solana_node_block_height",
			"Current block height of the RPC node",
			ClusterLabel,
		),
		NodeTransactionCount: NewGaugeDesc(
			"solana_node_transaction_count",
			"Total number of transactions processed by the RPC node",
			ClusterLabel,
		),
		NodeNumSlotsBehind: NewGaugeDesc(
			"solana_node_num_slots_behind",
			"Number of slots the RPC node is behind",
			ClusterLabel,
		),
		NodeBlockProcessingTime: NewGaugeDesc(
			"solana_node_block_processing_time_seconds",
			"Time taken to process the latest block in seconds",
			ClusterLabel,
		),
		NodeMinimumLedgerSlot: NewGaugeDesc(
			"solana_node_minimum_ledger_slot",
			"Lowest slot that the RPC node has information about in its ledger",
			ClusterLabel,
		),
		NodeFirstAvailableBlock: NewGaugeDesc(
			"solana_node_first_available_block",
			"First available block in the RPC node's ledger",
			ClusterLabel,
		),
	}
	return collector
}

func (c *SolanaCollector) Describe(ch chan<- *prometheus.Desc) {
    ch <- c.NodeVersion.Desc
    ch <- c.NodeHealth.Desc
    ch <- c.NodeSlotHeight.Desc
    ch <- c.NodeBlockHeight.Desc
    ch <- c.NodeTransactionCount.Desc
    ch <- c.NodeNumSlotsBehind.Desc
    ch <- c.NodeBlockProcessingTime.Desc
    ch <- c.NodeMinimumLedgerSlot.Desc
    ch <- c.NodeFirstAvailableBlock.Desc
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
        ch <- c.NodeVersion.MustNewConstMetric(1, c.config.ClusterName, version)
        c.logger.Debug("Version returned from cache")
        c.cacheMutex.RUnlock()
    } else {
        c.cacheMutex.RUnlock()
        callStart = time.Now()
        var err error
        version, err = c.rpcClient.GetVersion(ctx)
        c.logger.Debugf("GetVersion took: %v ms", time.Since(callStart).Milliseconds())
        if err == nil {
            ch <- c.NodeVersion.MustNewConstMetric(1, c.config.ClusterName, version)
            // Update cache
            c.cacheMutex.Lock()
            c.versionCache = &cachedVersion{
                value:     version,
                timestamp: time.Now(),
            }
            c.cacheMutex.Unlock()
        } else {
            c.logger.Errorw("Failed to collect version", "error", err)
            ch <- c.NodeVersion.MustNewConstMetric(0, c.config.ClusterName, "unknown")
            version = "unknown"
        }
    }

    // Check health cache
    c.cacheMutex.RLock()
    if c.healthCache != nil && time.Since(c.healthCache.timestamp) < c.cacheValidity {
        ch <- c.NodeHealth.MustNewConstMetric(float64(c.healthCache.isHealthy), c.config.ClusterName)
        ch <- c.NodeNumSlotsBehind.MustNewConstMetric(float64(c.healthCache.numSlotsBehind), c.config.ClusterName)
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
        ch <- c.NodeHealth.MustNewConstMetric(float64(isHealthy), c.config.ClusterName)
        ch <- c.NodeNumSlotsBehind.MustNewConstMetric(float64(numSlotsBehind), c.config.ClusterName)

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
        ch <- c.NodeMinimumLedgerSlot.MustNewConstMetric(float64(slot), c.config.ClusterName)
    } else {
        c.logger.Errorw("Failed to get minimum ledger slot", "error", err)
        ch <- c.NodeMinimumLedgerSlot.MustNewConstMetric(0, c.config.ClusterName)
    }

    // Collect first available block
    callStart = time.Now()
    block, err := c.rpcClient.GetFirstAvailableBlock(ctx)
    c.logger.Debugf("GetFirstAvailableBlock took: %v ms", time.Since(callStart).Milliseconds())
    if err == nil {
        ch <- c.NodeFirstAvailableBlock.MustNewConstMetric(float64(block), c.config.ClusterName)
    } else {
        c.logger.Errorw("Failed to get first available block", "error", err)
        ch <- c.NodeFirstAvailableBlock.MustNewConstMetric(0, c.config.ClusterName)
    }

    // Collect epoch info and measure block processing time
    blockStart := time.Now()
    epochInfo, err := c.rpcClient.GetEpochInfo(ctx, rpc.CommitmentConfirmed)
    blockProcessingTime := time.Since(blockStart).Seconds()
    c.logger.Debugf("GetEpochInfo took: %v ms", time.Since(blockStart).Milliseconds())
    
    if err == nil {
        ch <- c.NodeSlotHeight.MustNewConstMetric(float64(epochInfo.AbsoluteSlot), c.config.ClusterName)
        ch <- c.NodeBlockHeight.MustNewConstMetric(float64(epochInfo.BlockHeight), c.config.ClusterName)
        ch <- c.NodeTransactionCount.MustNewConstMetric(float64(epochInfo.TransactionCount), c.config.ClusterName)
        ch <- c.NodeBlockProcessingTime.MustNewConstMetric(blockProcessingTime, c.config.ClusterName)
        
        c.logger.Infow("Successfully collected metrics",
            "slot_height", epochInfo.AbsoluteSlot,
            "block_height", epochInfo.BlockHeight,
            "epoch", epochInfo.Epoch,
            "slots_behind", numSlotsBehind,
            "min_ledger_slot", slot,
            "first_available_block", block,
            "version", version,
            "block_processing_time", blockProcessingTime,
            "total_duration_ms", time.Since(totalStart).Milliseconds(),
        )
    } else {
        c.logger.Errorw("Failed to collect metrics", "error", err)
        ch <- c.NodeSlotHeight.MustNewConstMetric(0, c.config.ClusterName)
        ch <- c.NodeBlockHeight.MustNewConstMetric(0, c.config.ClusterName)
        ch <- c.NodeTransactionCount.MustNewConstMetric(0, c.config.ClusterName)
        ch <- c.NodeBlockProcessingTime.MustNewConstMetric(0, c.config.ClusterName)
    }

}
