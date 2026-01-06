package service

import (
	"database/sql"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ListItem struct {
	Ticker        string
	CompanyName   string
	CurrentPrice  sql.NullFloat64
	PreviousClose string
	DividendYield string
	PER           string
	PBR           sql.NullFloat64
	MarketCap     string
	Volume        sql.NullInt64
	PriceMovement string
	Signal        string
}

type DetailItem struct {
	Date          string
	CurrentPrice  sql.NullFloat64
	PreviousClose string
	DividendYield string
	PER           string
	PBR           sql.NullFloat64
	MarketCap     string
	Volume        sql.NullInt64
	PriceMovement string
	Signal        string
}

type ListView struct {
	Items []ListItem
}

type DetailView struct {
	Ticker        string
	CompanyName   string
	CurrentPrice  sql.NullFloat64
	PreviousClose string
	Signal        string
	MarketCap     string
	Items         []DetailItem
}

func GenerateHTML(db *sql.DB) error {
	if err := os.MkdirAll("output/detail", 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll("output/static", 0o755); err != nil {
		return err
	}

	if err := copyFile("static/style.css", "output/static/style.css"); err != nil {
		return err
	}

	items, err := loadListItems(db)
	if err != nil {
		return err
	}

	funcs := template.FuncMap{
		"formatFloat": formatFloat,
		"formatInt":   formatInt,
		"formatPBR":   formatPBR,
		"formatPER":   formatPER,
		"formatYield": formatYield,
		"formatChange": func(value string) string {
			if value == "" {
				return ""
			}
			clean := strings.TrimSpace(strings.ReplaceAll(value, "%", ""))
			clean = strings.TrimPrefix(clean, "+")
			val, err := strconv.ParseFloat(clean, 64)
			if err != nil {
				return value
			}
			if val > 0 {
				return fmt.Sprintf("+%.2f%%", val)
			}
			if val < 0 {
				return fmt.Sprintf("%.2f%%", val)
			}
			return "+0.00%"
		},
		"changeClass":  changeClass,
		"signalClass":  signalClass,
		"numericValue": numericValue,
	}

	listTmpl, err := template.New("list.html").Funcs(funcs).ParseFiles("templates/list.html")
	if err != nil {
		return err
	}

	listFile, err := os.Create("output/index.html")
	if err != nil {
		return err
	}
	defer listFile.Close()

	if err := listTmpl.Execute(listFile, ListView{Items: items}); err != nil {
		return err
	}

	for _, item := range items {
		detailItems, err := loadDetailItems(db, item.Ticker)
		if err != nil {
			return err
		}

		detailTmpl, err := template.New("detail.html").Funcs(funcs).ParseFiles("templates/detail.html")
		if err != nil {
			return err
		}

		filePath := filepath.Join("output", "detail", item.Ticker+".html")
		detailFile, err := os.Create(filePath)
		if err != nil {
			return err
		}

		view := DetailView{
			Ticker:        item.Ticker,
			CompanyName:   item.CompanyName,
			CurrentPrice:  item.CurrentPrice,
			PreviousClose: item.PreviousClose,
			Signal:        item.Signal,
			MarketCap:     item.MarketCap,
			Items:         detailItems,
		}
		if err := detailTmpl.Execute(detailFile, view); err != nil {
			detailFile.Close()
			return err
		}
		detailFile.Close()
	}

	return nil
}

func loadListItems(db *sql.DB) ([]ListItem, error) {
	rows, err := db.Query(`SELECT ticker, companyName, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val FROM watch_list ORDER BY ticker`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var item ListItem
		if err := rows.Scan(
			&item.Ticker,
			&item.CompanyName,
			&item.CurrentPrice,
			&item.PreviousClose,
			&item.DividendYield,
			&item.PER,
			&item.PBR,
			&item.MarketCap,
			&item.Volume,
			&item.PriceMovement,
			&item.Signal,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func loadDetailItems(db *sql.DB, ticker string) ([]DetailItem, error) {
	rows, err := db.Query(`SELECT yymmdd, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val FROM watch_detail WHERE ticker = ? ORDER BY yymmdd DESC`, ticker)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DetailItem
	for rows.Next() {
		var item DetailItem
		if err := rows.Scan(
			&item.Date,
			&item.CurrentPrice,
			&item.PreviousClose,
			&item.DividendYield,
			&item.PER,
			&item.PBR,
			&item.MarketCap,
			&item.Volume,
			&item.PriceMovement,
			&item.Signal,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(in)
	return err
}

func formatFloat(value sql.NullFloat64) string {
	if !value.Valid {
		return ""
	}
	if value.Float64 == float64(int64(value.Float64)) {
		return fmt.Sprintf("%.0f", value.Float64)
	}
	return fmt.Sprintf("%.2f", value.Float64)
}

func formatInt(value sql.NullInt64) string {
	if !value.Valid {
		return ""
	}
	return formatIntWithCommas(value.Int64)
}

func formatIntWithCommas(value int64) string {
	in := strconv.FormatInt(value, 10)
	if len(in) <= 3 {
		return in
	}

	var b strings.Builder
	for i, r := range in {
		if i != 0 && (len(in)-i)%3 == 0 {
			b.WriteRune(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func formatPBR(value sql.NullFloat64) string {
	formatted := formatFloat(value)
	if formatted == "" {
		return ""
	}
	return formatted + "x"
}

func formatPER(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	if strings.Contains(clean, "倍") {
		return clean
	}
	if strings.Contains(clean, "－") || strings.Contains(clean, "-") {
		return "－倍"
	}
	return clean + "倍"
}

func formatYield(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	if strings.Contains(clean, "%") || strings.Contains(clean, "％") {
		return clean
	}
	if strings.Contains(clean, "－") || strings.Contains(clean, "-") {
		return "－％"
	}
	return clean + "％"
}

func changeClass(value string) string {
	clean := strings.TrimSpace(strings.ReplaceAll(value, "%", ""))
	if clean == "" {
		return "neutral"
	}
	val, err := strconv.ParseFloat(strings.TrimPrefix(clean, "+"), 64)
	if err != nil {
		return "neutral"
	}
	if val > 0 {
		return "positive"
	}
	if val < 0 {
		return "negative"
	}
	return "neutral"
}

func signalClass(value string) string {
	switch strings.ToLower(value) {
	case "buy":
		return "buy"
	case "sell":
		return "sell"
	default:
		return "neutral"
	}
}

func numericValue(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	clean = strings.NewReplacer(",", "", "%", "", "％", "", "倍", "", "x", "", "－", "-").Replace(clean)
	clean = strings.TrimPrefix(clean, "+")
	if _, err := strconv.ParseFloat(clean, 64); err != nil {
		return ""
	}
	return clean
}
