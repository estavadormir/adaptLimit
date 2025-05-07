package predict

import (
	"sync"
	"time"
)

type DataPoint struct {
	Timestamp time.Time
	Value     float64
}

type Forecaster struct {
	history []DataPoint

	maxHistory int

	maWindow int

	alpha float64

	mu sync.RWMutex
}

func NewForecaster(options ...Option) *Forecaster {
	f := &Forecaster{
		maxHistory: 1000,
		maWindow:   10,
		alpha:      0.3,
	}

	for _, option := range options {
		option(f)
	}

	return f
}

type Option func(*Forecaster)

func WithMaxHistory(max int) Option {
	return func(f *Forecaster) {
		if max > 0 {
			f.maxHistory = max
		}
	}
}

func WithMAWindow(window int) Option {
	return func(f *Forecaster) {
		if window > 0 {
			f.maWindow = window
		}
	}
}

func WithAlpha(alpha float64) Option {
	return func(f *Forecaster) {
		if alpha >= 0 && alpha <= 1 {
			f.alpha = alpha
		}
	}
}

func (f *Forecaster) AddDataPoint(value float64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.history = append(f.history, DataPoint{
		Timestamp: time.Now(),
		Value:     value,
	})

	if len(f.history) > f.maxHistory {
		f.history = f.history[len(f.history)-f.maxHistory:]
	}
}

func (f *Forecaster) PredictNext() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.history) < 2 {
		if len(f.history) == 1 {
			return f.history[0].Value
		}
		return 0
	}

	return f.predictExponentialSmoothing()
}

func (f *Forecaster) PredictMovingAverage() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	histLen := len(f.history)
	if histLen == 0 {
		return 0
	}

	window := min(f.maWindow, histLen)
	sum := 0.0
	for i := histLen - window; i < histLen; i++ {
		sum += f.history[i].Value
	}

	return sum / float64(window)
}

func (f *Forecaster) predictExponentialSmoothing() float64 {
	histLen := len(f.history)
	if histLen == 0 {
		return 0
	}

	startIdx := max(0, histLen-f.maWindow)
	forecast := f.history[startIdx].Value

	for i := startIdx + 1; i < histLen; i++ {
		forecast = f.alpha*f.history[i].Value + (1-f.alpha)*forecast
	}

	return forecast
}

func (f *Forecaster) DetectPattern() (period time.Duration, confidence float64) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.history) < 24 {
		return 0, 0
	}

	patterns := []time.Duration{
		time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	bestConfidence := 0.0
	bestPeriod := time.Duration(0)

	for _, testPeriod := range patterns {
		confidence := f.evaluatePattern(testPeriod)
		if confidence > bestConfidence {
			bestConfidence = confidence
			bestPeriod = testPeriod
		}
	}

	return bestPeriod, bestConfidence
}

func (f *Forecaster) evaluatePattern(period time.Duration) float64 {
	bucketCount := 24 // Divide the period into 24 buckets
	buckets := make([][]float64, bucketCount)

	now := time.Now()

	for _, point := range f.history {
		age := now.Sub(point.Timestamp)
		position := int((age.Nanoseconds() % period.Nanoseconds()) * int64(bucketCount) / period.Nanoseconds())
		position = min(position, bucketCount-1)

		buckets[position] = append(buckets[position], point.Value)
	}

	totalVariance := f.calculateVariance(f.extractValues())
	if totalVariance == 0 {
		return 0 // No variation in the data
	}

	var withinBucketVariance float64
	var totalWeight float64

	for _, bucket := range buckets {
		if len(bucket) > 1 {
			variance := f.calculateVariance(bucket)
			weight := float64(len(bucket))
			withinBucketVariance += variance * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	withinBucketVariance /= totalWeight

	explainedVariance := 1.0 - (withinBucketVariance / totalVariance)

	dataConfidence := minFloat64(1.0, float64(len(f.history))/100.0)

	return explainedVariance * dataConfidence
}

func (f *Forecaster) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	varSum := 0.0
	for _, v := range values {
		diff := v - mean
		varSum += diff * diff
	}

	return varSum / float64(len(values))
}

func (f *Forecaster) extractValues() []float64 {
	values := make([]float64, len(f.history))
	for i, point := range f.history {
		values[i] = point.Value
	}
	return values
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
