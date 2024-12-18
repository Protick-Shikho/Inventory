package database

type DataFetcher interface {
    
    FetchData(tableName, column string) ([]interface{}, error)
    Forecast(table string, forecastedValue []float64) error
}
