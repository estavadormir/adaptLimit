package metrics

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Collector struct {
	cpuLoad    float64
	memoryLoad float64
	interval   time.Duration
	mu         sync.RWMutex
	stopCh     chan struct{}
}

func NewCollector(interval time.Duration) *Collector {
	c := &Collector{
		interval: interval,
		stopCh:   make(chan struct{}),
	}

	go c.collect()

	return c
}

func (c *Collector) CPULoad() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cpuLoad
}

func (c *Collector) MemoryLoad() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.memoryLoad
}

func (c *Collector) Stop() {
	close(c.stopCh)
}

func (c *Collector) collect() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	c.updateMetrics()

	for {
		select {
		case <-ticker.C:
			c.updateMetrics()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Collector) updateMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	numGoroutines := runtime.NumGoroutine()
	maxGoroutines := 1000
	c.cpuLoad = float64(numGoroutines) / float64(maxGoroutines)
	if c.cpuLoad > 1.0 {
		c.cpuLoad = 1.0
	}

	c.memoryLoad = float64(m.Alloc) / float64(m.Sys)

	if runtime.GOOS == "linux" {
		c.updateLinuxMetrics()
	}
}

func (c *Collector) updateLinuxMetrics() {
	loadBytes, err := os.ReadFile("/proc/loadavg")
	if err == nil && len(loadBytes) > 0 {
		var load float64
		_, err = fmt.Sscanf(string(loadBytes), "%f", &load)
		if err == nil {
			cpuCount := runtime.NumCPU()
			normalizedLoad := load / float64(cpuCount)
			c.cpuLoad = min(normalizedLoad, 1.0)
		}
	}

	memBytes, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		var total, free, available int64
		lines := strings.Split(string(memBytes), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(line, "MemTotal: %d", &total)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(line, "MemAvailable: %d", &available)
			} else if strings.HasPrefix(line, "MemFree:") && available == 0 {
				fmt.Sscanf(line, "MemFree: %d", &free)
			}
		}

		if total > 0 {
			used := total - max(available, free)
			c.memoryLoad = float64(used) / float64(total)
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
