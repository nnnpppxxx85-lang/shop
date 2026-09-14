package db

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect(dsn string) (*sql.DB, error) {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(5 * time.Minute)
	if err := DB.Ping(); err != nil {
		return nil, err
	}
	return DB, nil
}
