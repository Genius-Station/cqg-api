package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq" 
)

func InitDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		getEnv("MYSQL_USER"),
		getEnv("MYSQL_PASSWORD"),
		getEnv("MYSQL_HOST"),
		getEnv("MYSQL_PORT"),
		getEnv("MYSQL_DATABASE"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}


	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func InitDBPg() (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("PG_USER"),
		getEnv("PG_PASSWORD"),
		getEnv("PG_HOST"),
		getEnv("PG_PORT"),
		getEnv("PG_DATABASE"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key string, fallback ...string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return "" 
}
