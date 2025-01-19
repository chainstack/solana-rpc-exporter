package rpc

type (
	Response[T any] struct {
		Jsonrpc string   `json:"jsonrpc"`
		Result  T        `json:"result,omitempty"`
		Error   RPCError `json:"error,omitempty"`
		Id      int      `json:"id"`
	}

	ContextualResult[T any] struct {
		Context struct {
			Slot int64 `json:"slot"`
		} `json:"context"`
		Value T `json:"value"`
	}

	EpochInfo struct {
		AbsoluteSlot     int64 `json:"absoluteSlot"`
		BlockHeight      int64 `json:"blockHeight"`
		Epoch            int64 `json:"epoch"`
		SlotIndex        int64 `json:"slotIndex"`
		SlotsInEpoch     int64 `json:"slotsInEpoch"`
		TransactionCount int64 `json:"transactionCount"`
	}

	PerformanceStats struct {
		NumSlots         int64   `json:"numSlots"`
		NumTransactions  int64   `json:"numTransactions"`
		SamplePeriodSecs int64   `json:"samplePeriodSecs"`
		Slot             int64   `json:"slot"`
		TPS              float64 `json:"tps"`
	}

	SlotInfo struct {
		Parent int64  `json:"parent"`
		Slot   int64  `json:"slot"`
		Status string `json:"status"` // "confirmed", "processed", "finalized"
	}

	HealthStatus struct {
		Status      string `json:"status"`
		Message     string `json:"message,omitempty"`
		SlotsBehind int64  `json:"slotsBehind,omitempty"`
	}

	RPCVersionInfo struct {
		SolanaCore string `json:"solana-core"`
		FeatureSet uint32 `json:"feature-set"`
	}

	Block struct {
		BlockTime       int64         `json:"blockTime"`
		NumTransactions int           `json:"numTransactions"`
		Fee             int           `json:"fee"`
		Rewards         []BlockReward `json:"rewards,omitempty"`
	}

	BlockReward struct {
		Pubkey     string `json:"pubkey"`
		Lamports   int64  `json:"lamports"`
		RewardType string `json:"rewardType"`
	}
)

// Helper methods for HealthStatus
func (h *HealthStatus) IsHealthy() bool {
	return h.Status == "ok"
}
