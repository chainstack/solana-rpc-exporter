package rpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMockServer_getBalance(t *testing.T) {
	_, client := NewMockClient(
		t, 
		nil, 
		map[string]int{"aaa": 2 * LamportsInSol},
		nil,
		map[int]MockSlotInfo{},
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	balance, err := client.GetBalance(ctx, CommitmentFinalized, "aaa")
	assert.NoError(t, err)
	assert.Equal(t, float64(2), balance)
}

func TestMockServer_getBlock(t *testing.T) {
	_, client := NewMockClient(t,
		nil,
		nil,
		nil,
		map[int]MockSlotInfo{
			1: {Block: &MockBlockInfo{Fee: 10, NumTransactions: 1, BlockTime: 1234567890}},
			2: {Block: &MockBlockInfo{Fee: 5, NumTransactions: 2, BlockTime: 1234567891}},
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	block, err := client.GetBlock(ctx, CommitmentFinalized, 1, "full")
	assert.NoError(t, err)
	assert.Equal(t,
		Block{
			BlockTime: 1234567890,
			NumTransactions: 1,
			Fee: 10,
		},
		*block,
	)

	block, err = client.GetBlock(ctx, CommitmentFinalized, 2, "none")
	assert.NoError(t, err)
	assert.Equal(t,
		Block{
			BlockTime: 1234567891,
			NumTransactions: 2,
			Fee: 5,
		},
		*block,
	)
}

func TestMockServer_getInflationReward(t *testing.T) {
	_, client := NewMockClient(t,
		nil,
		nil,
		map[string]int{"AAA": 2_500, "BBB": 2_501, "CCC": 2_502},
		nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rewards, err := client.GetInflationReward(ctx, CommitmentFinalized, []string{"AAA", "BBB"}, 2)
	assert.NoError(t, err)
	assert.Equal(t,
		[]InflationReward{{Amount: 2_500, Epoch: 2}, {Amount: 2_501, Epoch: 2}},
		rewards,
	)
}

func TestMockServer_getSlotInfo(t *testing.T) {
	_, client := NewMockClient(t,
		nil,
		nil,
		nil,
		map[int]MockSlotInfo{
			1: {Parent: 0, Status: "finalized"},
			2: {Parent: 1, Status: "confirmed"},
			3: {Parent: 2, Status: "processed"},
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slotInfo, err := client.GetSlotInfo(ctx, CommitmentFinalized, 2)
	assert.NoError(t, err)
	assert.Equal(t,
		SlotInfo{
			Parent: 1,
			Slot: 2,
			Status: "confirmed",
		},
		*slotInfo,
	)
}
