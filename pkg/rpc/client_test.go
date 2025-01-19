package rpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"errors"
	"net/http"
	"time"
)

func newMethodTester(t *testing.T, method string, result any) (*MockServer, *Client) {
	t.Helper()
	return NewMockClient(t, map[string]any{method: result})
}

func TestClient_GetBlockTime(t *testing.T) {
	server, client := NewMockClient(t, map[string]any{})
	defer server.Close()

	server.SetOpt(BlockTimeOpt, int64(100), int64(1234567890))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blockTime, err := client.GetBlockTime(ctx, 100)
	assert.NoError(t, err)
	assert.Equal(t, int64(1234567890), blockTime)
}

func TestClient_GetEpochInfo(t *testing.T) {
	_, client := newMethodTester(t,
		"getEpochInfo",
		map[string]int64{
			"absoluteSlot":     166_598,
			"blockHeight":      166_500,
			"epoch":            27,
			"slotIndex":        2_790,
			"slotsInEpoch":     8_192,
			"transactionCount": 22_661_093,
		},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	epochInfo, err := client.GetEpochInfo(ctx, CommitmentFinalized)
	assert.NoError(t, err)
	assert.Equal(t,
		&EpochInfo{
			AbsoluteSlot:     166_598,
			BlockHeight:      166_500,
			Epoch:            27,
			SlotIndex:        2_790,
			SlotsInEpoch:     8_192,
			TransactionCount: 22_661_093,
		},
		epochInfo,
	)
}

func TestClient_GetFirstAvailableBlock(t *testing.T) {
	_, client := newMethodTester(t, "getFirstAvailableBlock", int64(250_000))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	block, err := client.GetFirstAvailableBlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(250_000), block)
}

func TestClient_GetHealth(t *testing.T) {
	_, client := newMethodTester(t, "getHealth", "ok")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	health, err := client.GetHealth(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "ok", health)

	// Test cached response
	health, err = client.GetHealth(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "ok", health)
}

func TestClient_GetMinimumLedgerSlot(t *testing.T) {
	_, client := newMethodTester(t, "minimumLedgerSlot", int64(250))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slot, err := client.GetMinimumLedgerSlot(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(250), slot)
}

func TestClient_GetVersion_Error(t *testing.T) {
    server, client := NewMockClient(t, map[string]any{})
    defer server.Close()
    
    // Set an explicit error response
    errorResp := &RPCError{
        Code:    -32601,
        Message: "Method not found",
        Method:  "getVersion",
    }
    server.SetOpt(EasyResultsOpt, "getVersion", errorResp)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    version, err := client.GetVersion(ctx)
    assert.Error(t, err)
    assert.Equal(t, "", version)
    
    var rpcErr *RPCError
    assert.True(t, errors.As(err, &rpcErr))
    assert.Equal(t, int64(-32601), rpcErr.Code)
}

func TestClient_TestConnection(t *testing.T) {
    // Test successful connection
    server, client := NewMockClient(t, map[string]any{
        "getVersion": map[string]any{"solana-core": "2.0.21"},
    })
    defer server.Close()
    err := client.TestConnection(context.Background())
    assert.NoError(t, err)

    // Test failed connection with RPC error
    server2, client2 := NewMockClient(t, map[string]any{})
    defer server2.Close()
    
    errorResp := &RPCError{
        Code:    -32601,
        Message: "Method not found",
        Method:  "getVersion",
    }
    server2.SetOpt(EasyResultsOpt, "getVersion", errorResp)
    
    err = client2.TestConnection(context.Background())
    assert.Error(t, err)
    var rpcErr *RPCError
    assert.True(t, errors.As(err, &rpcErr))

    // Test connection timeout
    badClient := &Client{
        HttpClient: http.Client{
            Timeout: 100 * time.Millisecond,
        },
        RpcUrl: "http://invalid-url:1234",
    }
    err = badClient.TestConnection(context.Background())
    assert.Error(t, err)
}
