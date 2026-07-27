package helper

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// InitDB opens a connection pool to PostgreSQL using DB_* environment
// variables (with sensible local defaults) and verifies it with a ping.
func InitDB() (*sql.DB, error) {
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "postgres"
	dbname := "postgres"
	sslmode := "disable"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
