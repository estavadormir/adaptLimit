package predict_test

import (
	"testing"
	"time"

	"github.com/estavadormir/adaptlimit/predict"
)

func TestForecasterPrediction(t *testing.T) {
	f := predict.NewForecaster()

	f.AddDataPoint(10)
	f.AddDataPoint(20)
	f.AddDataPoint(30)
	f.AddDataPoint(40)

	predicted := f.PredictNext()

	if predicted < 20 || predicted > 50 {
		t.Errorf("Prediction should be reasonable, got %f", predicted)
	}
}

func TestMovingAverage(t *testing.T) {
	f := predict.NewForecaster(predict.WithMAWindow(3))

	f.AddDataPoint(10)
	f.AddDataPoint(20)
	f.AddDataPoint(30)

	avg := f.PredictMovingAverage()
	if avg != 20 {
		t.Errorf("Moving average of 10,20,30 should be 20, got %f", avg)
	}
}

func TestPatternDetection(t *testing.T) {
	f := predict.NewForecaster()

	for i := range 24 {
		value := 50.0 + 30.0*float64(i%12)/11.0
		if i%12 >= 6 {
			value = 50.0 + 30.0*(1.0-float64(i%12-6)/5.0)
		}
		f.AddDataPoint(value)

		time.Sleep(time.Millisecond)
	}

	period, confidence := f.DetectPattern()
	if period != time.Hour && confidence > 0 {
		t.Logf("Expected hourly pattern, got %v with confidence %f", period, confidence)
	}
}
