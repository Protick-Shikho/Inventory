package service

import (
	"fmt"
	"strconv"

	"github.com/Protick-Shikho/inventory/database"
	"github.com/Protick-Shikho/inventory/domain"
	"github.com/Protick-Shikho/inventory/entities"
)

// ForecastService handles the logic for forecasting.
type ForecastService struct {
	DataFetcher database.DataFetcher
	Alpha       float64
	HoldingCostRate float64
	OrderingCost float64
	StockQuantity float64
}

// GenerateForecast fetches demand data, applies the Exponential Smoothing algorithm, and persists the results.
func (fs *ForecastService) GenerateForecast(table string) ([]entities.Forecast, error) {
	demandRaw, err := fs.DataFetcher.FetchData(table, "demand")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	// Convert demand data to float64
	demand := make([]float64, len(demandRaw))
	for i, val := range demandRaw {
		switch v := val.(type) {
		case float64:
			demand[i] = v
		case int64:
			demand[i] = float64(v)
		case uint8:
			strValue := string(v)
			floatValue, err := strconv.ParseFloat(strValue, 64) // Parse string to float64
			if err != nil {
				return nil, fmt.Errorf("failed to parse string value to float64: %v (value: %s)", err, strValue)
			}
			demand[i] = floatValue
		default:
			return nil, fmt.Errorf("failed to convert demand data to float64: %v (type %T)", val, val)
		}
	}

	// Apply the Exponential Smoothing algorithm (from the domain layer)
	forecastedValues, upcomingForecast, err := domain.ApplyExponentialSmoothing(demand, fs.Alpha)
	if err != nil {
		return nil, fmt.Errorf("failed to generate forecast: %w", err)
	}

	// Persist the forecasted values into the database
	err = fs.DataFetcher.Forecast(table, forecastedValues)
	if err != nil {
		return nil, fmt.Errorf("failed to insert data: %w", err)
	}

	// Convert forecasted values into Forecast entities
	forecasts := make([]entities.Forecast, len(forecastedValues))
	var temp int
	if len(forecasts)%12 == 0 {
		temp = 1
	} else {
		temp = len(forecasts) + 1
	}

	forecasts[len(forecasts)-1] = entities.Forecast{
		Month: temp,
		Value: upcomingForecast,
	}

	return forecasts, nil
}



func (s *ForecastService) GetCost(demand, eoq float64) (float64, float64) {
	Shortage, EOQ := domain.TotalCost(s.StockQuantity, s.HoldingCostRate, s.OrderingCost, demand, eoq)
	return Shortage, EOQ
}

func (e *ForecastService) CalculateEOQ(demand float64) float64 {
	
	return domain.EOQ(demand, e.HoldingCostRate, e.OrderingCost)
}

func (e *ForecastService) GetHoldingCost() float64 {
	return domain.HoldingCost(e.HoldingCostRate)
}

