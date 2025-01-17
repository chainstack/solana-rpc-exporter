package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

type MockOpt int

const (
	BalanceOpt MockOpt = iota
	InflationRewardsOpt
	EasyResultsOpt
	SlotInfoOpt
)

type (
	MockServer struct {
		server   *http.Server
		listener net.Listener
		mu       sync.RWMutex

		balances         map[string]int
		inflationRewards map[string]int
		easyResults      map[string]any
		slotInfos        map[int]MockSlotInfo
	}

	MockBlockInfo struct {
		Fee             int
		NumTransactions int
		BlockTime       int64
	}

	MockSlotInfo struct {
		Parent int64
		Status string // "confirmed", "processed", "finalized"
		Block  *MockBlockInfo
	}
)

// NewMockServer creates a new mock server instance
func NewMockServer(
	easyResults map[string]any,
	balances map[string]int,
	inflationRewards map[string]int,
	slotInfos map[int]MockSlotInfo,
) (*MockServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %v", err)
	}

	ms := &MockServer{
		listener:         listener,
		easyResults:      easyResults,
		balances:         balances,
		inflationRewards: inflationRewards,
		slotInfos:        slotInfos,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ms.handleRPCRequest)

	ms.server = &http.Server{Handler: mux}

	go func() {
		_ = ms.server.Serve(listener)
	}()

	return ms, nil
}

// URL returns the URL of the mock server
func (s *MockServer) URL() string {
	return fmt.Sprintf("http://%s", s.listener.Addr().String())
}

// Close shuts down the mock server
func (s *MockServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *MockServer) MustClose() {
	if err := s.Close(); err != nil {
		panic(err)
	}
}

func (s *MockServer) SetOpt(opt MockOpt, key any, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch opt {
	case BalanceOpt:
		if s.balances == nil {
			s.balances = make(map[string]int)
		}
		s.balances[key.(string)] = value.(int)
	case InflationRewardsOpt:
		if s.inflationRewards == nil {
			s.inflationRewards = make(map[string]int)
		}
		s.inflationRewards[key.(string)] = value.(int)
	case EasyResultsOpt:
		if s.easyResults == nil {
			s.easyResults = make(map[string]any)
		}
		s.easyResults[key.(string)] = value
	case SlotInfoOpt:
		if s.slotInfos == nil {
			s.slotInfos = make(map[int]MockSlotInfo)
		}
		s.slotInfos[key.(int)] = value.(MockSlotInfo)
	}
}

func (s *MockServer) getResult(method string, params ...any) (any, *RPCError) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if method == "getBalance" && s.balances != nil {
		address := params[0].(string)
		result := map[string]any{
			"context": map[string]int{"slot": 1},
			"value":   s.balances[address],
		}
		return result, nil
	}

	if method == "getInflationReward" && s.inflationRewards != nil {
		addresses := params[0].([]any)
		config := params[1].(map[string]any)
		epoch := int(config["epoch"].(float64))
		rewards := make([]map[string]int, len(addresses))
		for i, item := range addresses {
			address := item.(string)
			rewards[i] = map[string]int{"amount": s.inflationRewards[address], "epoch": epoch}
		}
		return rewards, nil
	}

	if method == "getBlock" && s.slotInfos != nil {
		slot := int(params[0].(float64))
		slotInfo, ok := s.slotInfos[slot]
		if !ok {
			return nil, &RPCError{Code: BlockCleanedUpCode, Message: "Block cleaned up."}
		}
		if slotInfo.Block == nil {
			return nil, &RPCError{Code: SlotSkippedCode, Message: "Slot skipped."}
		}

		return map[string]any{
			"blockTime":       slotInfo.Block.BlockTime,
			"numTransactions": slotInfo.Block.NumTransactions,
			"fee":             slotInfo.Block.Fee,
		}, nil
	}

	if method == "getSlotInfo" && s.slotInfos != nil {
		slot := int(params[0].(float64))
		slotInfo, ok := s.slotInfos[slot]
		if !ok {
			return nil, &RPCError{Code: BlockCleanedUpCode, Message: "Slot cleaned up."}
		}

		return map[string]any{
			"parent": slotInfo.Parent,
			"slot":   slot,
			"status": slotInfo.Status,
		}, nil
	}

	// default is use easy results
	result, ok := s.easyResults[method]
	if !ok {
		return nil, &RPCError{Code: -32601, Message: "Method not found"}
	}
	return result, nil
}

func (s *MockServer) handleRPCRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var request Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := Response[any]{Jsonrpc: "2.0", Id: request.Id}
	result, rpcErr := s.getResult(request.Method, request.Params...)
	if rpcErr != nil {
		response.Error = *rpcErr
	} else {
		response.Result = result
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// NewMockClient creates a new test client with a running mock server
func NewMockClient(
	t *testing.T,
	easyResults map[string]any,
	balances map[string]int,
	inflationRewards map[string]int,
	slotInfos map[int]MockSlotInfo,
) (*MockServer, *Client) {
	server, err := NewMockServer(easyResults, balances, inflationRewards, slotInfos)
	if err != nil {
		t.Fatalf("failed to create mock server: %v", err)
	}

	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Errorf("failed to close mock server: %v", err)
		}
	})

	client := &Client{
		HttpClient: http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
			},
			Timeout: time.Second,
		},
		RpcUrl:        server.URL(),
		HttpTimeout:   time.Second,
		cacheValidity: 60 * time.Second,
		logger:        nil, // don't set logger for tests
	}
	return server, client
}
