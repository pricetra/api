package database

import (
	"database/sql"
	"os"
)

const PG_LOCAL_URL = "postgresql://postgres:postgres@localhost:5435/postgres?sslmode=disable"

func CreateDbConnection() (*sql.DB, error) {
	if db_url, exists := os.LookupEnv("PG_DATABASE_URL"); exists {
		return sql.Open("postgres", db_url)
	}
	return sql.Open("postgres", PG_LOCAL_URL)
}

func NewDbConnection() (*sql.DB, error) {
	return CreateDbConnection()
}
