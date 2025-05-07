package metrics_test

import (
	"testing"
	"time"

	"github.com/estavadormir/adaptlimit/metrics"
)

func TestMetricsCollector(t *testing.T) {
	collector := metrics.NewCollector(time.Millisecond * 100)
	defer collector.Stop()

	time.Sleep(time.Millisecond * 200)

	cpuLoad := collector.CPULoad()
	memLoad := collector.MemoryLoad()

	if cpuLoad < 0 || cpuLoad > 1 {
		t.Errorf("CPULoad should be between 0 and 1, got %f", cpuLoad)
	}

	if memLoad < 0 || memLoad > 1 {
		t.Errorf("MemoryLoad should be between 0 and 1, got %f", memLoad)
	}
}
