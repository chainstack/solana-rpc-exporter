package main

import (
	"testing"

	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestGetEpochBounds(t *testing.T) {
	epoch := &rpc.EpochInfo{
		AbsoluteSlot: 25,
		SlotIndex:    5,
		SlotsInEpoch: 10,
	}
	first, last := GetEpochBounds(epoch)
	assert.Equal(t, int64(20), first)
	assert.Equal(t, int64(29), last)
}

func TestIsValidNetwork(t *testing.T) {
	assert.True(t, isValidNetwork("mainnet-beta"))
	assert.True(t, isValidNetwork("testnet"))
	assert.True(t, isValidNetwork("devnet"))
	assert.True(t, isValidNetwork("localnet"))
	assert.False(t, isValidNetwork("something-else"))
	assert.False(t, isValidNetwork(""))
}

func TestToString(t *testing.T) {
	assert.Equal(t, "123", toString(123))
	assert.Equal(t, "test", toString("test"))
	assert.Equal(t, "true", toString(true))
	assert.Equal(t, "<nil>", toString(nil))
}

func TestAssertf(t *testing.T) {
	// Test that nothing bad happens when condition is true.
	// (We can’t fully test a failing assertf without causing a fatal exit.)
	assert.NotPanics(t, func() {
		assertf(true, "should not fail")
	})
}
