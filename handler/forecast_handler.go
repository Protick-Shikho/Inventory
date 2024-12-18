package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/Protick-Shikho/inventory/entities"
	"github.com/Protick-Shikho/inventory/service"
	"github.com/Protick-Shikho/inventory/utils"
)

type ForecastHandler struct {
	ForecastService *service.ForecastService
}

func (fh *ForecastHandler) GetForecast(w http.ResponseWriter, r *http.Request) {

	table := utils.GetTable(r)

	forecast, err := fh.ForecastService.GenerateForecast(table)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating forecast: %v", err), http.StatusInternalServerError)
		return
	}

	// Response structure
	response := entities.Forecast{
		Month: forecast[len(forecast)-1].Month,
		Value: forecast[len(forecast)-1].Value,
	}

	// Set content type and return the forecast response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func (h *ForecastHandler) GetEOQ(w http.ResponseWriter, r *http.Request) {

	// Generate Holding Cost
	holdingCost := h.ForecastService.GetHoldingCost()

	table := utils.GetTable(r)

	apiURL := fmt.Sprintf("http://localhost:8080/forecast?table=%s", url.QueryEscape(table))

	// Make the GET request to the API
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch forecast: %v", err)
	}
	defer resp.Body.Close()

	// Check if the status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: Received non-OK HTTP status %s", resp.Status)
	}

	var forecast entities.Forecast
	err = json.NewDecoder(resp.Body).Decode(&forecast)
	if err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	//Generate EOQ
	EOQ := h.ForecastService.CalculateEOQ(forecast.Value)

	respones := entities.EOQ{
		EOQ:             EOQ,
		HoldingCostRate: holdingCost,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respones)
}

func (h *ForecastHandler) GetCost(w http.ResponseWriter, r *http.Request) {

	table := utils.GetTable(r)

	demand, err := h.ForecastService.GenerateForecast(table)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating forecast: %v", err), http.StatusInternalServerError)
		return
	}
	lastForecastValue := demand[len(demand)-1].Value
	EOQ := h.ForecastService.CalculateEOQ(lastForecastValue)

	Shortage, EOQTotalCost := h.ForecastService.GetCost(lastForecastValue, EOQ)

	response := entities.Cost{
		EOQ:            EOQTotalCost,
		ShortageAmount: Shortage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

