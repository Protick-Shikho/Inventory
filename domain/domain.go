package domain

import (
	"fmt"
	"math"
)

// ApplyExponentialSmoothing applies the Exponential Smoothing algorithm to forecast demand.
func ApplyExponentialSmoothing(demand []float64, alpha float64) ([]float64, float64, error) {
	if len(demand) == 0 {
		return nil, 0, fmt.Errorf("demand data is empty")
	}

	forecastedValues := []float64{}
	initialForecast := demand[0]

	// Initialize the forecast with the first demand value
	forecastedValues = append(forecastedValues, 0)
	forecastedValues = append(forecastedValues, initialForecast)

	for i := 2; i < len(demand); i++ {
		nextValue := forecastedValues[i-1] + alpha*(demand[i-1]-forecastedValues[i-1])
		forecastedValues = append(forecastedValues, nextValue)
	}

	upcomingForecast := forecastedValues[len(forecastedValues)-1] + alpha*(demand[len(demand)-1]-forecastedValues[len(forecastedValues)-1])

	return forecastedValues, upcomingForecast, nil
}

func HoldingCost(holdingCostRate float64) float64 {
	return holdingCostRate * 0.06
}

func TotalCost(StockQuantity, HoldingCostRate, OrderingCost, Demand, eoq float64) (float64, float64) {

	shortage := (Demand * 50) - StockQuantity

	totalHoldingCost := HoldingCostRate * shortage

	tc := OrderingCost + totalHoldingCost

	totalHoldingCost = eoq * HoldingCostRate
	totalOrderingCost := (Demand / eoq) * OrderingCost

	tcEOQ := totalHoldingCost + totalOrderingCost

	return tc, tcEOQ
}

func EOQ(demand, holdingCost, orderingCost float64) float64 {

	if holdingCost <= 0 || demand <= 0 || orderingCost <= 0 {
		return 0
	}

	EOQ := math.Sqrt((2.00 * demand * 50 * orderingCost) / holdingCost)

	return EOQ
}
