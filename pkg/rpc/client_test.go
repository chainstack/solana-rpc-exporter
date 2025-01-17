package rpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newMethodTester(t *testing.T, method string, result any) (*MockServer, *Client) {
	t.Helper()
	return NewMockClient(t, map[string]any{method: result}, nil, nil, nil)
}

func TestClient_GetBalance(t *testing.T) {
	_, client := newMethodTester(t,
		"getBalance",
		map[string]any{"context": map[string]int{"slot": 1}, "value": 5 * LamportsInSol},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	balance, err := client.GetBalance(ctx, CommitmentFinalized, "")
	assert.NoError(t, err)
	assert.Equal(t, float64(5), balance)
}

func TestClient_GetBlock(t *testing.T) {
	_, client := newMethodTester(t,
		"getBlock",
		map[string]any{
			"blockTime": 1234567890,
			"numTransactions": 1,
			"fee": 10,
			"rewards": []map[string]any{
				{"pubkey": "aaa", "lamports": 10, "rewardType": "fee"},
			},
			"transactions": []map[string]any{
				{"transaction": map[string]any{"message": map[string]any{"accountKeys": []any{"aaa", "bbb", "ccc"}}}},
			},
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	block, err := client.GetBlock(ctx, CommitmentFinalized, 0, "full")
	assert.NoError(t, err)
	assert.Equal(t,
		&Block{
			BlockTime: 1234567890,
			NumTransactions: 1,
			Fee: 10,
			Rewards: []BlockReward{
				{Pubkey: "aaa", Lamports: 10, RewardType: "fee"},
			},
			Transactions: []TransactionInfo{
				{Transaction: map[string]any{"message": map[string]any{"accountKeys": []any{"aaa", "bbb", "ccc"}}}},
			},
		},
		block,
	)
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

func TestClient_GetInflationReward(t *testing.T) {
	_, client := newMethodTester(t,
		"getInflationReward",
		[]map[string]int{
			{
				"amount": 2_500,
				"epoch": 2,
			},
		},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addresses := []string{"test-address"}
	inflationReward, err := client.GetInflationReward(ctx, CommitmentFinalized, addresses, 2)
	assert.NoError(t, err)
	assert.Equal(t,
		[]InflationReward{{Amount: 2_500, Epoch: 2}},
		inflationReward,
	)
}

func TestClient_GetMinimumLedgerSlot(t *testing.T) {
	_, client := newMethodTester(t, "minimumLedgerSlot", int64(250))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slot, err := client.GetMinimumLedgerSlot(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(250), slot)
}

func TestClient_GetSlotLeader(t *testing.T) {
	expectedLeader := "Leader123"
	_, client := newMethodTester(t, "getSlotLeader", expectedLeader)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	leader, err := client.GetSlotLeader(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedLeader, leader)
}

func TestClient_GetVersion(t *testing.T) {
	expectedResult := map[string]any{"feature-set": 2891131721, "solana-core": "1.16.7"}
	_, client := newMethodTester(t, "getVersion", expectedResult)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	version, err := client.GetVersion(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult["solana-core"], version)

	// Test cached response
	version, err = client.GetVersion(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult["solana-core"], version)
}

func TestClient_GetPerfCounters(t *testing.T) {
	expectedCounters := map[string]int64{
		"blocksProcessed": 100,
		"transactionsProcessed": 1000,
	}
	_, client := newMethodTester(t, "getPerfCounters", expectedCounters)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counters, err := client.GetPerfCounters(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedCounters, counters)
}

func TestClient_GetSlotInfo(t *testing.T) {
	_, client := newMethodTester(t,
		"getSlotInfo",
		map[string]any{
			"parent": int64(99),
			"slot": int64(100),
			"status": "finalized",
		},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slotInfo, err := client.GetSlotInfo(ctx, CommitmentFinalized, 100)
	assert.NoError(t, err)
	assert.Equal(t,
		&SlotInfo{
			Parent: 99,
			Slot: 100,
			Status: "finalized",
		},
		slotInfo,
	)
}

func TestClient_TestConnection(t *testing.T) {
    // Test successful connection
    _, client := newMethodTester(t, "getVersion", map[string]any{"solana-core": "1.16.7"})
    err := client.TestConnection(context.Background())
    assert.NoError(t, err)

    // Test failed connection - need to return RPCError for failure case
    _, client = newMethodTester(t, "getVersion", nil)
    server := client.RpcUrl // Save the URL before replacing it
    client.RpcUrl = "http://invalid-url:1234" // Use invalid URL to force error
    err = client.TestConnection(context.Background())
    assert.Error(t, err)
    client.RpcUrl = server // Restore the URL
}
