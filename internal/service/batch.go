package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"smallcap-watcher-go/internal/api"
	"smallcap-watcher-go/internal/parse"
)

type WatchRow struct {
	Ticker string
}

func RunBatch(db *sql.DB, scraperBaseURL string) error {
	tickers, err := loadTickers(db)
	if err != nil {
		return err
	}

	client := api.NewClientWithBaseURL(scraperBaseURL)
	ctx := context.Background()

	for _, ticker := range tickers {
		payload, err := client.FetchStockData(ctx, ticker)
		if err != nil {
			log.Printf("failed to fetch %s: %v", ticker, err)
			continue
		}

		if err := updateStock(db, payload); err != nil {
			log.Printf("failed to update %s: %v", ticker, err)
		}
	}

	return nil
}

func loadTickers(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT ticker FROM watch_list")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickers []string
	for rows.Next() {
		var ticker string
		if err := rows.Scan(&ticker); err != nil {
			return nil, err
		}
		tickers = append(tickers, ticker)
	}
	return tickers, rows.Err()
}

func updateStock(db *sql.DB, payload api.StockResponse) error {
	currentPriceVal, currentOK := parse.ParseNumeric(payload.CurrentPrice)
	pbrVal, pbrOK := parse.ParseNumeric(payload.PBR)
	volumeVal, volumeOK := parse.ParseNumeric(payload.Volume)
	previousCloseVal, prevOK := parse.ParsePreviousClose(payload.PreviousClose)

	var pricemovement string
	var signal string
	if currentOK && prevOK && previousCloseVal != 0 {
		movement := ((currentPriceVal - previousCloseVal) / previousCloseVal) * 100
		pricemovement = fmt.Sprintf("%.2f%%", movement)
		signal = generateSignal(movement)
	} else {
		pricemovement = ""
		signal = "Neutral"
	}

	now := time.Now().Format("060102")

	_, err := db.Exec(
		`INSERT INTO watch_list (
			ticker, companyName, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			companyName=VALUES(companyName),
			currentPrice=VALUES(currentPrice),
			previousClose=VALUES(previousClose),
			dividendYield=VALUES(dividendYield),
			per=VALUES(per),
			pbr=VALUES(pbr),
			marketCap=VALUES(marketCap),
			volume=VALUES(volume),
			pricemovement=VALUES(pricemovement),
			signal_val=VALUES(signal_val)`,
		payload.Ticker,
		payload.CompanyName,
		nullFloat(currentPriceVal, currentOK),
		payload.PreviousClose,
		payload.DividendYield,
		payload.PER,
		nullFloat(pbrVal, pbrOK),
		payload.MarketCap,
		nullInt(volumeVal, volumeOK),
		pricemovement,
		signal,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`REPLACE INTO watch_detail (
			ticker, yymmdd, companyName, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.Ticker,
		now,
		payload.CompanyName,
		nullFloat(currentPriceVal, currentOK),
		payload.PreviousClose,
		payload.DividendYield,
		payload.PER,
		nullFloat(pbrVal, pbrOK),
		payload.MarketCap,
		nullInt(volumeVal, volumeOK),
		pricemovement,
		signal,
	)
	return err
}

func generateSignal(movement float64) string {
	if movement > 3.0 {
		return "Buy"
	}
	if movement < -3.0 {
		return "Sell"
	}
	return "Neutral"
}

func nullFloat(value float64, ok bool) sql.NullFloat64 {
	return sql.NullFloat64{Float64: value, Valid: ok}
}

func nullInt(value float64, ok bool) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(value), Valid: ok}
}
