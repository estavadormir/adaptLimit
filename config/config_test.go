package config_test

import (
	"testing"
	"time"

	"github.com/estavadormir/adaptlimit/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.InitialLimit != 10 {
		t.Errorf("Expected InitialLimit to be 10, got %d", cfg.InitialLimit)
	}

	if cfg.Interval != time.Second {
		t.Errorf("Expected Interval to be 1s, got %v", cfg.Interval)
	}

	// Test other default values
}

func TestConfigChaining(t *testing.T) {
	cfg := config.DefaultConfig().
		WithInitialLimit(20).
		WithInterval(time.Millisecond * 500)

	if cfg.InitialLimit != 20 {
		t.Errorf("Expected InitialLimit to be 20, got %d", cfg.InitialLimit)
	}

	if cfg.Interval != time.Millisecond*500 {
		t.Errorf("Expected Interval to be 500ms, got %v", cfg.Interval)
	}
}
