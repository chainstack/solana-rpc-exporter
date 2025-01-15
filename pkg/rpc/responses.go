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

	ClusterNodes struct {
		Pubkey       string `json:"pubkey"`
		Gossip      string `json:"gossip"`
		TPU         string `json:"tpu"`
		RPC         string `json:"rpc"`
		Version     string `json:"version"`
		FeatureSet  uint32 `json:"featureSet"`
		ShredVer    uint16 `json:"shredVersion"`
	}

	PerformanceStats struct {
		NumSlots              int64   `json:"numSlots"`
		NumTransactions       int64   `json:"numTransactions"`
		SamplePeriodSecs     int64   `json:"samplePeriodSecs"`
		Slot                 int64   `json:"slot"`
		TPS                  float64 `json:"tps"`
		TPSMax               float64 `json:"tpsMax"`
		TPSMin               float64 `json:"tpsMin"`
	}

	SlotInfo struct {
		Parent         int64  `json:"parent"`
		Root          int64  `json:"root"`
		Slot          int64  `json:"slot"`
		Status        string `json:"status"` // "confirmed", "processed", "finalized"
	}

	HealthStatus struct {
		Status      string `json:"status"`
		Message     string `json:"message,omitempty"`
		SlotsBehind int64  `json:"slotsBehind,omitempty"`
	}

	RPCVersionInfo struct {
		SolanaCore  string `json:"solana-core"`
		FeatureSet  uint32 `json:"feature-set"`
	}

	BlockInfo struct {
		BlockTime         int64                    `json:"blockTime"`
		BlockHeight      int64                    `json:"blockHeight"`
		NumTransactions  int64                    `json:"numTransactions"`
		Slot             int64                    `json:"slot"`
		ParentSlot       int64                    `json:"parentSlot"`
		Blockhash        string                   `json:"blockhash"`
		PreviousBlockhash string                  `json:"previousBlockhash"`
		Rewards          []BlockReward            `json:"rewards,omitempty"`
		Transactions     []TransactionInfo        `json:"transactions,omitempty"`
	}

	BlockReward struct {
		Pubkey       string `json:"pubkey"`
		Lamports    int64  `json:"lamports"`
		PostBalance int64  `json:"postBalance"`
		RewardType   string `json:"rewardType"`
		Commission   *int   `json:"commission,omitempty"`
	}

	TransactionInfo struct {
		Signature string           `json:"signature"`
		Slot      int64           `json:"slot"`
		BlockTime *int64          `json:"blockTime,omitempty"`
		Err       interface{}     `json:"err,omitempty"`
		Meta      *TransactionMeta `json:"meta,omitempty"`
	}

	TransactionMeta struct {
		Fee               int64    `json:"fee"`
		PreBalances      []int64  `json:"preBalances"`
		PostBalances     []int64  `json:"postBalances"`
		LogMessages      []string `json:"logMessages,omitempty"`
		Status           interface{} `json:"status"`
		ComputeUnitsConsumed *int64 `json:"computeUnitsConsumed,omitempty"`
	}

	SupplyInfo struct {
		Total                   int64    `json:"total"`
		Circulating            int64    `json:"circulating"`
		NonCirculating         int64    `json:"nonCirculating"`
		NonCirculatingAccounts []string `json:"nonCirculatingAccounts"`
	}

	TokenAmount struct {
		Amount   string   `json:"amount"`
		Decimals int      `json:"decimals"`
		UiAmount *float64 `json:"uiAmount"`
	}
)

// Helper methods for HealthStatus
func (h *HealthStatus) IsHealthy() bool {
	return h.Status == "ok"
}

// Helper methods for TokenAmount
func (t *TokenAmount) GetUiAmount() float64 {
	if t.UiAmount != nil {
		return *t.UiAmount
	}
	return 0
}
