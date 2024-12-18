package connector

import "database/sql"

type DatabaseConnection interface {
	ConnectDB() error
	Close() error
	GetDB() *sql.DB
}
