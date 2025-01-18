package main

import (
	"context"
	"fmt"
	"github.com/naviat/solana-rpc-exporter/pkg/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

type RPCSimulator struct {
	Server           *rpc.MockServer
	Slot             int64
	BlockHeight      int64
	TransactionCount int64
	IsHealthy        bool
	SlotsBehind      int64
	Version          string

	// Constants for the simulator
	SlotTime      time.Duration
	MinLedgerSlot int64
}

func NewRPCSimulator(t *testing.T, slot int64) (*RPCSimulator, *rpc.Client) {
	mockServer, client := rpc.NewMockClient(t,
		map[string]any{
			"getVersion": map[string]any{
				"solana-core": "1.16.7",
				"feature-set": 2891131721,
			},
			"getHealth": "ok",
		},
		nil, nil, nil,
	)

	simulator := &RPCSimulator{
		Server:           mockServer,
		Slot:             0,
		BlockHeight:      0,
		TransactionCount: 0,
		IsHealthy:        true,
		Version:          "1.16.7",
		SlotTime:         400 * time.Millisecond,
		MinLedgerSlot:    0,
	}

	// Initialize state
	simulator.PopulateState(0)
	if slot > 0 {
		for simulator.Slot < slot {
			simulator.Slot++
			simulator.PopulateState(simulator.Slot)
		}
	}

	return simulator, client
}

func (s *RPCSimulator) PopulateState(slot int64) {
	// Update block height (some slots might be skipped)
	if slot%4 != 3 { // Every 4th slot is skipped
		s.BlockHeight++
		// Add some random transactions
		s.TransactionCount += rand.Int63n(10) + 1
	}

	// Update minimum ledger slot (keep last 2000 slots)
	if slot > 2000 {
		s.MinLedgerSlot = slot - 2000
	}

	// Randomly become unhealthy (1% chance)
	if rand.Float64() < 0.01 {
		s.IsHealthy = false
		s.SlotsBehind = rand.Int63n(1000) + 1
	} else {
		s.IsHealthy = true
		s.SlotsBehind = 0
	}

	// Update server state
	s.updateServerState()
}

func (s *RPCSimulator) updateServerState() {
	// Update epoch info
	s.Server.SetOpt(rpc.EasyResultsOpt, "getEpochInfo", map[string]int64{
		"absoluteSlot":     s.Slot,
		"blockHeight":      s.BlockHeight,
		"epoch":            s.Slot / 432000, // Assuming epoch length
		"slotIndex":        s.Slot % 432000,
		"slotsInEpoch":     432000,
		"transactionCount": s.TransactionCount,
	})

	// Update health status
	if s.IsHealthy {
		s.Server.SetOpt(rpc.EasyResultsOpt, "getHealth", "ok")
	} else {
		s.Server.SetOpt(rpc.EasyResultsOpt, "getHealth", &rpc.RPCError{
			Code:    rpc.NodeUnhealthyCode,
			Message: fmt.Sprintf("Node is behind by %d slots", s.SlotsBehind),
			Data: map[string]any{
				"numSlotsBehind": s.SlotsBehind,
			},
		})
	}

	// Update minimum ledger slot
	s.Server.SetOpt(rpc.EasyResultsOpt, "minimumLedgerSlot", s.MinLedgerSlot)
	s.Server.SetOpt(rpc.EasyResultsOpt, "getFirstAvailableBlock", s.MinLedgerSlot)
}

func (s *RPCSimulator) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.Slot++
			s.PopulateState(s.Slot)
			// Add 5% random noise to slot time
			noiseRange := float64(s.SlotTime) * 0.05
			noise := time.Duration((rand.Float64()*2 - 1) * noiseRange)
			time.Sleep(s.SlotTime + noise)
		}
	}
}

func TestSolanaCollector_WithSimulator(t *testing.T) {
	simulator, client := NewRPCSimulator(t, 100) // Start at slot 100
	config := &ExporterConfig{
		NetworkName: "test-network",
	}
	collector := NewSolanaCollector(client, config)

	// Run simulator in background
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go simulator.Run(ctx)

	// Test multiple collection cycles
	for i := 0; i < 3; i++ {
		registry := prometheus.NewPedanticRegistry()
		registry.MustRegister(collector)

		// Collect and verify metrics
		metrics, err := registry.Gather()
		assert.NoError(t, err)

		// Verify basic metric presence and values
		found := make(map[string]bool)
		for _, metric := range metrics {
			found[metric.GetName()] = true

			switch metric.GetName() {
			case "solana_node_health":
				value := metric.GetMetric()[0].GetGauge().GetValue()
				assert.True(t, value == 0 || value == 1)
			case "solana_node_transaction_count":
				assert.True(t, metric.GetMetric()[0].GetGauge().GetValue() > 0)
			}
		}

		// Verify all expected metrics exist
		expectedMetrics := []string{
			"solana_node_version_info",
			"solana_node_health",
			"solana_node_transaction_count",
			"solana_node_num_slots_behind",
			"solana_node_minimum_ledger_slot",
			"solana_node_first_available_block",
		}
		for _, name := range expectedMetrics {
			assert.True(t, found[name], "Missing metric: %s", name)
		}

		time.Sleep(500 * time.Millisecond)
	}
}
