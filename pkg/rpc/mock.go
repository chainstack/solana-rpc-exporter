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
	EasyResultsOpt MockOpt = iota
	BlockTimeOpt
)

type MockServer struct {
	server   *http.Server
	listener net.Listener
	mu       sync.RWMutex

	easyResults map[string]any
	blockTimes  map[int64]int64
}

func NewMockServer(easyResults map[string]any) (*MockServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %v", err)
	}

	ms := &MockServer{
		listener:    listener,
		easyResults: easyResults,
		blockTimes:  make(map[int64]int64),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ms.handleRPCRequest)

	ms.server = &http.Server{Handler: mux}

	go func() {
		_ = ms.server.Serve(listener)
	}()

	return ms, nil
}

func (s *MockServer) URL() string {
	return fmt.Sprintf("http://%s", s.listener.Addr().String())
}

func (s *MockServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *MockServer) SetOpt(opt MockOpt, key any, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch opt {
	case EasyResultsOpt:
		if s.easyResults == nil {
			s.easyResults = make(map[string]any)
		}
		s.easyResults[key.(string)] = value
	case BlockTimeOpt:
		s.blockTimes[key.(int64)] = value.(int64)
	}
}

func (s *MockServer) getResult(method string, params ...any) (any, *RPCError) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch method {
	case "getBlockTime":
		if len(params) == 0 {
			return nil, &RPCError{
				Code:    -32602,
				Message: "Invalid params",
				Method:  method,
			}
		}
		slot := params[0].(float64)
		if blockTime, ok := s.blockTimes[int64(slot)]; ok {
			return blockTime, nil
		}
		return nil, &RPCError{
			Code:    -32004,
			Message: "Block not available",
			Method:  method,
		}

	default:
		// Fall back to easy results
		if result, ok := s.easyResults[method]; ok {
			// If result is an RPCError, return it directly
			if rpcErr, ok := result.(*RPCError); ok {
				rpcErr.Method = method
				return nil, rpcErr
			}
			return result, nil
		}
		return nil, &RPCError{
			Code:    -32601,
			Message: "Method not found",
			Method:  method,
		}
	}
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewMockClient(t *testing.T, easyResults map[string]any) (*MockServer, *Client) {
	server, err := NewMockServer(easyResults)
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
	}
	return server, client
}
