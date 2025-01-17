package main

import (
	"github.com/naviat/solana-rpc-exporter/pkg/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	slog.Init()
	code := m.Run()
	os.Exit(code)
}
