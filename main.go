package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Protick-Shikho/inventory/Infrastructure/connector"
	"github.com/Protick-Shikho/inventory/database"
	Handler "github.com/Protick-Shikho/inventory/handler"
	"github.com/Protick-Shikho/inventory/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {

	connectorInstance := &database.MySQLDatabase{} // Create a MySQL instance of the database connector

	var db connector.DatabaseConnection = connectorInstance

	err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	dbInstance := &database.MySQLDatabase{DB: db.GetDB()} // Create a MySQL instance of the main database

	var mainDB database.DataFetcher = dbInstance


	// Initialize use case
	alpha := 0.30
	holdingCostRate := 0.20
	stockQuantity := 10000.00
	orderingCost := 44.50

	// Initialize Forecast Service
	ForecastService := &service.ForecastService{
		DataFetcher:     mainDB,
		Alpha:           alpha,
		HoldingCostRate: holdingCostRate,
		OrderingCost:    orderingCost,
		StockQuantity:   stockQuantity,
	}

	forecastHandler := &Handler.ForecastHandler{
		ForecastService: ForecastService,
	}


	http.HandleFunc("/forecast", forecastHandler.GetForecast)
	http.HandleFunc("/EOQ", forecastHandler.GetEOQ)
	http.HandleFunc("/cost", forecastHandler.GetCost)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())


	// Start the HTTP server
	fmt.Println("Server is running on port 8080...")
	fmt.Println("http://localhost:8080/forecast?table=sales_data")
	fmt.Println("http://localhost:8080/EOQ?table=sales_data")
	fmt.Println("http://localhost:8080/cost?table=sales_data")
	fmt.Println("http://localhost:8080/metrics")
	

	log.Fatal(http.ListenAndServe(":8080", nil))
}
