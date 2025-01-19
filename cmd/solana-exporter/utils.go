package main

import (
	"fmt"
	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
)

// GetEpochBounds returns the first slot and last slot within an [inclusive] Epoch
func GetEpochBounds(info *rpc.EpochInfo) (int64, int64) {
	firstSlot := info.AbsoluteSlot - info.SlotIndex
	return firstSlot, firstSlot + info.SlotsInEpoch - 1
}

// isValidNetwork checks if the provided network name is valid
func isValidNetwork(network string) bool {
	validNetworks := []string{"mainnet-beta", "testnet", "devnet", "localnet"}
	for _, valid := range validNetworks {
		if network == valid {
			return true
		}
	}
	return false
}

// assertf is a utility function for runtime assertions
func assertf(condition bool, format string, args ...any) {
	logger := slog.Get()
	if !condition {
		logger.Fatalf(format, args...)
	}
}

// toString is a simple utility function for converting to strings
func toString(i any) string {
	return fmt.Sprintf("%v", i)
}
