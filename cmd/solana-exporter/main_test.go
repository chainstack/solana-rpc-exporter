package main

import (
	"github.com/naviat/solana-exporter-rpc/pkg/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	slog.Init()
	code := m.Run()
	os.Exit(code)
}
