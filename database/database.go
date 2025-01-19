package database

type DataFetcher interface {
    
    FetchData(tableName, column string) ([]interface{}, string, error)
    Forecast(table string, forecastedValue []float64) error
}
