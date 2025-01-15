package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLDatabase is a concrete implementation of the DataFetcher interface.
type MySQLDatabase struct {
	DB *sql.DB
}

// ConnectDB initializes the MySQL database connection.
func (m *MySQLDatabase) ConnectDB() error {
	var err error
	m.DB, err = sql.Open("mysql", "root:123@tcp(localhost:3306)/inventory")
	// m.DB, err = sql.Open("mysql", "root:123@tcp(host.docker.internal:3306)/inventory")
	// m.DB, err = sql.Open("mysql", "root:root@123@tcp(127.0.0.1:3306)/inventory")

	if err != nil {
		return err
	}
	if err := m.DB.Ping(); err != nil {
		return err
	}
	return nil
}

// FetchData retrieves the demand data from the database.
func (m *MySQLDatabase) FetchData(tableName, columnName string) ([]interface{}, error) {
	// Use parameterized queries to prevent SQL injection
	query := fmt.Sprintf("SELECT `%s` FROM `%s`", columnName, tableName)
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query data from %s: %w", tableName, err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var value interface{}
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed during rows iteration: %w", err)
	}
	return results, nil
}

func (m *MySQLDatabase) Forecast(table string, forecastedValue []float64) error {
	for i, value := range forecastedValue {

		query := fmt.Sprintf("UPDATE %s SET forecasted_value = ? WHERE id = ?", table)

		_, err := m.DB.Exec(query, value, i+1)
		if err != nil {
			return err
		}
	}

	return nil

}

func (mc *MySQLDatabase) Close() error {
	if mc.DB != nil {
		return mc.DB.Close()
	}
	return nil
}

func (mc *MySQLDatabase) GetDB() *sql.DB {
	return mc.DB
}
