package service

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func SeedWatchList(db *sql.DB, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lineNo++
		if lineNo == 1 && strings.HasPrefix(strings.ToLower(line), "ticker") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 1 {
			return fmt.Errorf("invalid TSV format at line %d", lineNo)
		}

		ticker := strings.TrimSpace(parts[0])
		companyName := ""
		if len(parts) > 1 {
			companyName = strings.TrimSpace(parts[1])
		}
		if ticker == "" {
			return fmt.Errorf("empty ticker at line %d", lineNo)
		}

		if _, err := db.Exec(
			`INSERT INTO watch_list (ticker, companyName) VALUES (?, ?) ON DUPLICATE KEY UPDATE companyName=VALUES(companyName)`,
			ticker,
			companyName,
		); err != nil {
			return err
		}
	}
	return scanner.Err()
}
