package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockServer_SetOpt(t *testing.T) {
	server, client := NewMockClient(t, map[string]any{})
	defer server.Close()

	// Test EasyResultsOpt
	server.SetOpt(EasyResultsOpt, "getVersion", map[string]any{
		"solana-core": "2.0.21",
		"feature-set": 2891131721,
	})

	ctx := context.Background()
	version, err := client.GetVersion(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "2.0.21", version)

	// Test BlockTimeOpt
	server.SetOpt(BlockTimeOpt, int64(100), int64(1234567890))
	blockTime, err := client.GetBlockTime(ctx, 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(1234567890), blockTime)
}

func TestMockServer_GetEpochInfo(t *testing.T) {
	server, client := NewMockClient(t, map[string]any{
		"getEpochInfo": map[string]int64{
			"absoluteSlot":     100,
			"blockHeight":      90,
			"epoch":            2,
			"slotIndex":        100,
			"slotsInEpoch":     432000,
			"transactionCount": 1000,
		},
	})
	defer server.Close()

	ctx := context.Background()
	epochInfo, err := client.GetEpochInfo(ctx, CommitmentConfirmed)
	assert.NoError(t, err)
	assert.Equal(t, &EpochInfo{
		AbsoluteSlot:     100,
		BlockHeight:      90,
		Epoch:            2,
		SlotIndex:        100,
		SlotsInEpoch:     432000,
		TransactionCount: 1000,
	}, epochInfo)
}

func TestMockServer_GetHealth(t *testing.T) {
    // Test healthy node
    server, client := NewMockClient(t, map[string]any{
        "getHealth": "ok",
    })
    defer server.Close()

    ctx := context.Background()
    health, err := client.GetHealth(ctx)
    assert.NoError(t, err)
    assert.Equal(t, "ok", health)

    // Test unhealthy node
    server.SetOpt(EasyResultsOpt, "getHealth", &RPCError{
        Code:    NodeUnhealthyCode,
        Message: "Node is behind by 1000 slots",
        Method:  "getHealth",
        Data: map[string]any{
            "numSlotsBehind": int64(1000),
        },
    })

    // Clear the cache to force a new request
    client.cacheMutex.Lock()
    client.healthCache = nil
    client.cacheMutex.Unlock()

    // Check the unhealthy response
    health, err = client.GetHealth(ctx)
    assert.Error(t, err)
    if assert.NotNil(t, err) {
        var rpcErr *RPCError
        if assert.True(t, errors.As(err, &rpcErr)) {
            assert.Equal(t, NodeUnhealthyCode, rpcErr.Code)
            assert.Equal(t, "getHealth", rpcErr.Method)
        }
    }
}

func TestMockServer_GetFirstAvailableBlock(t *testing.T) {
	server, client := NewMockClient(t, map[string]any{
		"getFirstAvailableBlock": int64(1000),
	})
	defer server.Close()

	ctx := context.Background()
	block, err := client.GetFirstAvailableBlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), block)
}

func TestMockServer_GetMinimumLedgerSlot(t *testing.T) {
	server, client := NewMockClient(t, map[string]any{
		"minimumLedgerSlot": int64(500),
	})
	defer server.Close()

	ctx := context.Background()
	slot, err := client.GetMinimumLedgerSlot(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), slot)
}

func TestMockServer_ErrorHandling(t *testing.T) {
	server, client := NewMockClient(t, map[string]any{})
	defer server.Close()

	errorResp := &RPCError{
		Code:    -32601,
		Message: "Method not found",
		Method:  "getVersion",
	}
	server.SetOpt(EasyResultsOpt, "getVersion", errorResp)

	ctx := context.Background()
	version, err := client.GetVersion(ctx)
	assert.Error(t, err)
	assert.Equal(t, "", version)

	var rpcErr *RPCError
	assert.True(t, errors.As(err, &rpcErr))
	assert.Equal(t, int64(-32601), rpcErr.Code)
}
