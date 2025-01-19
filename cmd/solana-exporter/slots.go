package main

import (
	"context"
	"time"

	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
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
}

func NewSlotWatcher(client *rpc.Client, config *ExporterConfig) *SlotWatcher {
	logger := slog.Get()

	watcher := SlotWatcher{
		client: client,
		logger: logger,
		config: config,
	}

	return &watcher
}

func (w *SlotWatcher) WatchSlots(ctx context.Context) error {
	ticker := time.NewTicker(w.config.SlotPace)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			epochInfo, err := w.client.GetEpochInfo(ctx, rpc.CommitmentConfirmed)
			if err != nil {
				continue
			}

			if w.currentEpoch == 0 || epochInfo.Epoch > w.currentEpoch {
				firstSlot, lastSlot := GetEpochBounds(epochInfo)
				w.currentEpoch = epochInfo.Epoch
				w.firstSlot = firstSlot
				w.lastSlot = lastSlot
			}

			w.slotWatermark = epochInfo.AbsoluteSlot
		}
	}
}
