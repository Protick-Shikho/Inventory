package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

var ctx = context.Background()

// MySQLDatabase is a concrete implementation of the DataFetcher interface.
type MySQLDatabase struct {
	DB    *sql.DB
	Redis *redis.Client
}

func (m *MySQLDatabase) ConnectDB() error {
	var err error
	m.DB, err = sql.Open("mysql", "root:123@tcp(localhost:3306)/inventory")
	if err != nil {
		return err
	}
	if err := m.DB.Ping(); err != nil {
		return err
	}

	if m.Redis == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return nil
}

func (m *MySQLDatabase) FetchData(tableName, columnName string) ([]interface{}, string,  error) {
	cacheKey := fmt.Sprintf("%s:%s", tableName, columnName)
	cachedData, err := m.Redis.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		log.Printf("Cache miss for key: %s", cacheKey)

		query := fmt.Sprintf("SELECT `%s` FROM `%s`", columnName, tableName)
		rows, err := m.DB.Query(query)
		if err != nil {
			return nil, "", fmt.Errorf("failed to query data from %s: %w", tableName, err)
		}
		defer rows.Close()

		var results []interface{}
		for rows.Next() {
			var value interface{}
			if err := rows.Scan(&value); err != nil {
				return nil, "", fmt.Errorf("failed to scan row: %w", err)
			}
			results = append(results, value)
		}
		if err := rows.Err(); err != nil {
			return nil,"", fmt.Errorf("failed during rows iteration: %w", err)
		}

		cacheValue, err := json.Marshal(results)
		if err != nil {
			log.Printf("Failed to serialize cache value: %v", err)
		} else {
			err = m.Redis.Set(ctx, cacheKey, cacheValue, 5*time.Second).Err()
			if err != nil {
				log.Printf("Failed to set cache for key %s: %v", cacheKey, err)
			}
		}
		return results, "", nil
	} else if err != nil {
		return nil, "", fmt.Errorf("failed to get from Redis: %w", err)
	}

	log.Printf("Cache hit for key: %s", cacheKey)

	var cachedResults []interface{}
	err = json.Unmarshal([]byte(cachedData), &cachedResults)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return cachedResults, "abc", nil
}

// Forecast updates forecasted values in the specified table.
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

// Close closes the MySQL and Redis connections.
func (m *MySQLDatabase) Close() error {
	if m.DB != nil {
		if err := m.DB.Close(); err != nil {
			return err
		}
	}
	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			return err
		}
	}
	return nil
}

// GetDB returns the database connection.
func (m *MySQLDatabase) GetDB() *sql.DB {
	return m.DB
}
