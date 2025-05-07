package predict_test

import (
	"testing"

	"github.com/estavadormir/adaptlimit/predict"
)

func TestForecasterPrediction(t *testing.T) {
	f := predict.NewForecaster()

	f.AddDataPoint(10)
	f.AddDataPoint(20)
	f.AddDataPoint(30)
	f.AddDataPoint(40)

	predicted := f.PredictNext()

	if predicted < 25 || predicted > 50 {
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
