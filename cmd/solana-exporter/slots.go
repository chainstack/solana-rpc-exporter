package main

import (
	"context"
	"fmt"
	"time"

	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type SlotWatcher struct {
	client *rpc.Client
	logger *zap.SugaredLogger
	config *ExporterConfig

	// currentEpoch tracking
	currentEpoch  int64
	firstSlot     int64
	lastSlot      int64
	slotWatermark int64

	// prometheus metrics
	SlotHeightMetric         prometheus.Gauge
	BlockHeightMetric        prometheus.Gauge
	EpochFirstSlotMetric     prometheus.Gauge
	EpochLastSlotMetric      prometheus.Gauge
	SlotProcessingTimeMetric prometheus.Gauge
	SkippedSlotsMetric       prometheus.Counter
}

func NewSlotWatcher(client *rpc.Client, config *ExporterConfig) *SlotWatcher {
	logger := slog.Get()

	watcher := SlotWatcher{
		client: client,
		logger: logger,
		config: config,

		SlotHeightMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "solana_node_slot_height",
			Help: "Current slot height of the RPC node",
			ConstLabels: prometheus.Labels{
				"network": config.NetworkName,
			},
		}),

		BlockHeightMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "solana_node_block_height",
			Help: "Current block height of the RPC node",
			ConstLabels: prometheus.Labels{
				"network": config.NetworkName,
			},
		}),

		EpochFirstSlotMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "solana_node_epoch_first_slot",
			Help: "First slot of current epoch",
			ConstLabels: prometheus.Labels{
				"network": config.NetworkName,
			},
		}),

		EpochLastSlotMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "solana_node_epoch_last_slot",
			Help: "Last slot of current epoch",
			ConstLabels: prometheus.Labels{
				"network": config.NetworkName,
			},
		}),

		SlotProcessingTimeMetric: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "solana_node_slot_processing_time",
			Help: "Time taken to process each slot in seconds",
			ConstLabels: prometheus.Labels{
				"network": config.NetworkName,
			},
		}),

		SkippedSlotsMetric: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "solana_node_skipped_slots_total",
			Help: "Total number of skipped slots",
			ConstLabels: prometheus.Labels{
				"network": config.NetworkName,
			},
		}),
	}

	// Register metrics
	logger.Info("Registering slot watcher metrics...")
	for _, collector := range []prometheus.Collector{
		watcher.SlotHeightMetric,
		watcher.BlockHeightMetric,
		watcher.EpochFirstSlotMetric,
		watcher.EpochLastSlotMetric,
		watcher.SlotProcessingTimeMetric,
		watcher.SkippedSlotsMetric,
	} {
		if err := prometheus.Register(collector); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				logger.Fatal(fmt.Errorf("failed to register collector: %w", err))
			}
		}
	}

	return &watcher
}

func (w *SlotWatcher) WatchSlots(ctx context.Context) error {
	ticker := time.NewTicker(w.config.SlotPace)
	defer ticker.Stop()

	w.logger.Infof("Starting slot watcher, running every %vs", w.config.SlotPace.Seconds())

	for {
		select {
		case <-ctx.Done():
			w.logger.Infof("Stopping WatchSlots() at slot %v", w.slotWatermark)
			return ctx.Err()
		case <-ticker.C:
			start := time.Now()

			// Get epoch info with confirmed commitment
			epochInfo, err := w.client.GetEpochInfo(ctx, rpc.CommitmentConfirmed)
			if err != nil {
				w.logger.Errorf("Failed to get epoch info: %v", err)
				continue
			}

			// Update metrics
			w.SlotHeightMetric.Set(float64(epochInfo.AbsoluteSlot))
			w.BlockHeightMetric.Set(float64(epochInfo.BlockHeight))

			// Track epoch boundaries
			if w.currentEpoch == 0 || epochInfo.Epoch > w.currentEpoch {
				firstSlot, lastSlot := GetEpochBounds(epochInfo)
				w.currentEpoch = epochInfo.Epoch
				w.firstSlot = firstSlot
				w.lastSlot = lastSlot

				w.EpochFirstSlotMetric.Set(float64(firstSlot))
				w.EpochLastSlotMetric.Set(float64(lastSlot))
			}

			// Track skipped slots
			if epochInfo.AbsoluteSlot > w.slotWatermark+1 {
				skipped := epochInfo.AbsoluteSlot - (w.slotWatermark + 1)
				w.SkippedSlotsMetric.Add(float64(skipped))
			}

			w.slotWatermark = epochInfo.AbsoluteSlot

			// Record processing time
			w.SlotProcessingTimeMetric.Set(time.Since(start).Seconds())
		}
	}
}
