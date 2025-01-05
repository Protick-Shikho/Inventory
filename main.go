package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Protick-Shikho/inventory/Infrastructure/connector"
	"github.com/Protick-Shikho/inventory/database"
	Handler "github.com/Protick-Shikho/inventory/handler"
	"github.com/Protick-Shikho/inventory/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTP request metrics
var httpRequestsTotal = prometheus.NewCounter(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(httpRequestsTotal)
}

// Middleware to track metrics for HTTP requests
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        httpRequestsTotal.Inc() // Increment total request count
        next.ServeHTTP(w, r)
    })
}


func main() {
	// Database connection setup
	connectorInstance := &database.MySQLDatabase{}

	var db connector.DatabaseConnection = connectorInstance

	if err := db.ConnectDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	dbInstance := &database.MySQLDatabase{DB: db.GetDB()}
	var mainDB database.DataFetcher = dbInstance

	// Initialize use case
	alpha := 0.30
	holdingCostRate := 0.20
	stockQuantity := 10000.00
	orderingCost := 44.50

	// Initialize Forecast Service
	forecastService := &service.ForecastService{
		DataFetcher:     mainDB,
		Alpha:           alpha,
		HoldingCostRate: holdingCostRate,
		OrderingCost:    orderingCost,
		StockQuantity:   stockQuantity,
	}

	forecastHandler := &Handler.ForecastHandler{
		ForecastService: forecastService,
	}

	// HTTP handlers
	mux := http.NewServeMux()
	mux.Handle("/forecast", metricsMiddleware(http.HandlerFunc(forecastHandler.GetForecast)))
	mux.Handle("/EOQ", metricsMiddleware(http.HandlerFunc(forecastHandler.GetEOQ)))
	mux.Handle("/cost", metricsMiddleware(http.HandlerFunc(forecastHandler.GetCost)))
	mux.Handle("/metrics", promhttp.Handler())

	// Configurable port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		fmt.Printf("Server is running on port %s...\n", port)
		fmt.Println("Endpoints:")
		fmt.Printf("http://localhost:%s/forecast?table=sales_data\n", port)
		fmt.Printf("http://localhost:%s/EOQ?table=sales_data\n", port)
		fmt.Printf("http://localhost:%s/cost?table=sales_data\n", port)
		fmt.Printf("http://localhost:%s/metrics\n", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
