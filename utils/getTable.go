package utils

import (
	"fmt"
	"net/http"
)

func GetTable(r *http.Request) string {

	table := r.URL.Query().Get("table")
	if table == "" {
		fmt.Println("table query parameter is required")
		return ""
	}
	return table
}
