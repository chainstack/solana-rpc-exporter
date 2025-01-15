package main

import (
	"fmt"
	"github.com/naviat/solana-exporter-rpc/pkg/rpc"
	"github.com/naviat/solana-exporter-rpc/pkg/slog"
)

// assertf is a utility function for runtime assertions
func assertf(condition bool, format string, args ...any) {
	logger := slog.Get()
	if !condition {
		logger.Fatalf(format, args...)
	}
}

// GetEpochBounds returns the first slot and last slot within an [inclusive] Epoch
func GetEpochBounds(info *rpc.EpochInfo) (int64, int64) {
	firstSlot := info.AbsoluteSlot - info.SlotIndex
	return firstSlot, firstSlot + info.SlotsInEpoch - 1
}

// toString is just a simple utility function for converting to strings
func toString(i any) string {
	return fmt.Sprintf("%v", i)
}

// isValidCluster checks if the provided cluster name is valid
func isValidCluster(cluster string) bool {
	validClusters := []string{"mainnet", "testnet", "devnet", "localnet"}
	for _, valid := range validClusters {
		if cluster == valid {
			return true
		}
	}
	return false
}
