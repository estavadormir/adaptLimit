package config

import (
	"time"
)

type Config struct {
	//the init rate limit per interval
	InitialLimit int

	//the minimum rate limit allowed
	MinLimit int

	//the maximum rate limit allowed
	MaxLimit int

	//the time period for the rate limit
	Interval time.Duration

	//how often the limits are adjusted
	AdjustInterval time.Duration

	//how often system metrics are collected
	MetricsInterval time.Duration

	//the threshold for high system load (0.0-1.0)
	HighLoadThreshold float64

	//the threshold for low system load (0.0-1.0)
	LowLoadThreshold float64

	//the threshold for high error rate (0.0-1.0)
	HighErrorThreshold float64

	//the threshold for low error rate (0.0-1.0)
	LowErrorThreshold float64

	//the target response time for requests
	TargetResponseTime time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		InitialLimit:       100,
		MinLimit:           10,
		MaxLimit:           1000,
		Interval:           time.Second,
		AdjustInterval:     time.Second * 30,
		MetricsInterval:    time.Second * 5,
		HighLoadThreshold:  0.75,
		LowLoadThreshold:   0.25,
		HighErrorThreshold: 0.05,
		LowErrorThreshold:  0.01,
		TargetResponseTime: time.Millisecond * 200,
	}
}

func (c *Config) WithInitialLimit(limit int) *Config {
	c.InitialLimit = limit
	return c
}

func (c *Config) WithMinLimit(limit int) *Config {
	c.MinLimit = limit
	return c
}

func (c *Config) WithMaxLimit(limit int) *Config {
	c.MaxLimit = limit
	return c
}

func (c *Config) WithInterval(interval time.Duration) *Config {
	c.Interval = interval
	return c
}

func (c *Config) WithAdjustInterval(interval time.Duration) *Config {
	c.AdjustInterval = interval
	return c
}

func (c *Config) WithMetricsInterval(interval time.Duration) *Config {
	c.MetricsInterval = interval
	return c
}

func (c *Config) WithLoadThresholds(low, high float64) *Config {
	c.LowLoadThreshold = low
	c.HighLoadThreshold = high
	return c
}

func (c *Config) WithErrorThresholds(low, high float64) *Config {
	c.LowErrorThreshold = low
	c.HighErrorThreshold = high
	return c
}

func (c *Config) WithTargetResponseTime(duration time.Duration) *Config {
	c.TargetResponseTime = duration
	return c
}
