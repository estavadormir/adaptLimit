package config_test

import (
	"testing"
	"time"

	"github.com/estavadormir/adaptlimit/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	if cfg.InitialLimit != 100 {
		t.Errorf("Expected InitialLimit to be 100, got %d", cfg.InitialLimit)
	}

	if cfg.Interval != time.Second {
		t.Errorf("Expected Interval to be 1s, got %v", cfg.Interval)
	}

	if cfg.MinLimit != 10 {
		t.Errorf("Expected MinLimit to be 10, got %d", cfg.MinLimit)
	}

	if cfg.MaxLimit != 1000 {
		t.Errorf("Expected MaxLimit to be 1000, got %d", cfg.MaxLimit)
	}
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

	cfg = cfg.WithMinLimit(5)
	if cfg.MinLimit != 5 {
		t.Errorf("Expected MinLimit to be 5, got %d", cfg.MinLimit)
	}

	cfg = cfg.WithMaxLimit(200)
	if cfg.MaxLimit != 200 {
		t.Errorf("Expected MaxLimit to be 200, got %d", cfg.MaxLimit)
	}
}
