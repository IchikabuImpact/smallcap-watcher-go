package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"smallcap-watcher-go/internal/config"
)

func Open(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBName)
	return sql.Open("mysql", dsn)
}

func InitSchema(db *sql.DB) error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS watch_list (
			ticker VARCHAR(10) PRIMARY KEY,
			companyName VARCHAR(255),
			currentPrice DECIMAL(10,2),
			previousClose VARCHAR(20),
			dividendYield VARCHAR(20),
			per VARCHAR(20),
			pbr DECIMAL(5,2),
			marketCap VARCHAR(50),
			volume INT,
			pricemovement VARCHAR(50),
			signal_val VARCHAR(50),
			memo TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS watch_detail (
			ticker VARCHAR(10) NOT NULL,
			yymmdd CHAR(6) NOT NULL,
			companyName VARCHAR(255),
			currentPrice DECIMAL(10,2),
			previousClose VARCHAR(20),
			dividendYield VARCHAR(20),
			per VARCHAR(20),
			pbr DECIMAL(5,2),
			marketCap VARCHAR(50),
			volume INT,
			pricemovement VARCHAR(50),
			signal_val VARCHAR(50),
			memo TEXT,
			PRIMARY KEY (ticker, yymmdd)
		)`,
	}

	for _, stmt := range schema {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
