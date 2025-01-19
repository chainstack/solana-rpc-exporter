package rpc

import (
	"encoding/json"
	"fmt"
)

const (
	// RPC Error Codes
	NodeUnhealthyCode = -32005
	NodeBehindCode    = -32009
	TimeoutCode       = -32000
)

type (
	NodeUnhealthyErrorData struct {
		NumSlotsBehind int64 `json:"numSlotsBehind"`
	}

	NodeBehindErrorData struct {
		SlotsBehind int64 `json:"slotsBehind"`
	}

	RPCError struct {
		Message string         `json:"message"`
		Code    int64         `json:"code"`
		Data    map[string]any `json:"data"`
		Method  string         // Added for context
	}
)

func (e *RPCError) Error() string {
	return fmt.Sprintf("%s RPC error (code: %d): %s (data: %v)", e.Method, e.Code, e.Message, e.Data)
}

func UnpackRpcErrorData[T any](rpcErr *RPCError, formatted T) error {
	bytesData, err := json.Marshal(rpcErr.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal %s RPC error data: %w", rpcErr.Method, err)
	}

	if err = json.Unmarshal(bytesData, formatted); err != nil {
		return fmt.Errorf("failed to unmarshal %s RPC error data: %w", rpcErr.Method, err)
	}

	return nil
}

// Helper functions for error checking
func IsNodeUnhealthy(err error) bool {
	if rpcErr, ok := err.(*RPCError); ok {
		return rpcErr.Code == NodeUnhealthyCode
	}
	return false
}

func IsNodeBehind(err error) bool {
	if rpcErr, ok := err.(*RPCError); ok {
		return rpcErr.Code == NodeBehindCode
	}
	return false
}

func IsTimeout(err error) bool {
	if rpcErr, ok := err.(*RPCError); ok {
		return rpcErr.Code == TimeoutCode
	}
	return false
}
