package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"smallcap-watcher-go/internal/config"
	"smallcap-watcher-go/internal/db"
	"smallcap-watcher-go/internal/service"
)

func main() {
	initFlag := flag.Bool("init", false, "initialize database schema")
	batchFlag := flag.Bool("batch", false, "fetch data and update database")
	genFlag := flag.Bool("gen", false, "generate static HTML output")
	seedFlag := flag.Bool("seed", false, "seed watch list from src/tickers1.tsv")
	flag.Parse()

	if !*initFlag && !*batchFlag && !*genFlag && !*seedFlag {
		fmt.Fprintln(os.Stderr, "Usage: smallcap-watcher --init | --batch | --gen | --seed")
		os.Exit(1)
	}

	cfg := config.Load()
	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if *initFlag {
		if err := db.InitSchema(database); err != nil {
			log.Fatalf("failed to init schema: %v", err)
		}
		log.Println("schema initialized")
	}

	if *seedFlag {
		if err := service.SeedWatchList(database, "src/tickers1.tsv"); err != nil {
			log.Fatalf("failed to seed watch list: %v", err)
		}
		log.Println("watch list seeded")
	}

	if *batchFlag {
		if err := service.RunBatch(database, cfg.ScraperBaseURL, cfg.ScraperRequestInterval); err != nil {
			log.Fatalf("failed to run batch: %v", err)
		}
		log.Println("batch completed")
	}

	if *genFlag {
		if err := service.GenerateHTML(database, cfg.OutputDir); err != nil {
			log.Fatalf("failed to generate HTML: %v", err)
		}
		if err := verifyIndexFreshness(cfg.OutputDir, cfg.IndexMaxAge); err != nil {
			log.Fatalf("HTML generation guard failed: %v", err)
		}
		log.Println("HTML generated")
	}
}

func verifyIndexFreshness(outputDir string, maxAge time.Duration) error {
	indexPath := filepath.Join(outputDir, "index.html")
	indexInfo, err := os.Stat(indexPath)
	if err != nil {
		return fmt.Errorf("stat index %s: %w", indexPath, err)
	}
	indexAge := time.Since(indexInfo.ModTime())

	detailPath, detailInfo, err := newestDetailFileInfo(outputDir)
	if err != nil {
		return err
	}

	log.Printf("output healthcheck index_path=%s index_mtime=%s index_size=%d index_age=%s detail_path=%s detail_mtime=%s detail_size=%d",
		indexPath,
		indexInfo.ModTime().UTC().Format(time.RFC3339),
		indexInfo.Size(),
		indexAge.Round(time.Second),
		detailPath,
		detailInfo.ModTime().UTC().Format(time.RFC3339),
		detailInfo.Size(),
	)

	if indexInfo.Size() == 0 {
		return fmt.Errorf("index file is empty: %s", indexPath)
	}
	if maxAge > 0 && indexAge > maxAge {
		return fmt.Errorf("index file too old (age=%s max_age=%s path=%s)", indexAge.Round(time.Second), maxAge, indexPath)
	}
	if indexInfo.ModTime().Before(detailInfo.ModTime().Add(-1 * time.Minute)) {
		return fmt.Errorf("index is older than latest detail page (index=%s detail=%s)", indexInfo.ModTime().UTC().Format(time.RFC3339), detailInfo.ModTime().UTC().Format(time.RFC3339))
	}

	return nil
}

func newestDetailFileInfo(outputDir string) (string, os.FileInfo, error) {
	detailPattern := filepath.Join(outputDir, "detail", "*.html")
	matches, err := filepath.Glob(detailPattern)
	if err != nil {
		return "", nil, fmt.Errorf("glob detail files failed for %s: %w", detailPattern, err)
	}
	if len(matches) == 0 {
		return "", nil, fmt.Errorf("no detail files found under %s", filepath.Join(outputDir, "detail"))
	}

	var newestPath string
	var newestInfo os.FileInfo
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			return "", nil, fmt.Errorf("stat detail file %s: %w", path, err)
		}
		if newestInfo == nil || info.ModTime().After(newestInfo.ModTime()) {
			newestPath = path
			newestInfo = info
		}
	}

	return newestPath, newestInfo, nil
}
