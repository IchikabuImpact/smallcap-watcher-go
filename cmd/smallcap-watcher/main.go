package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
		if err := service.RunBatch(database); err != nil {
			log.Fatalf("failed to run batch: %v", err)
		}
		log.Println("batch completed")
	}

	if *genFlag {
		if err := service.GenerateHTML(database); err != nil {
			log.Fatalf("failed to generate HTML: %v", err)
		}
		log.Println("HTML generated")
	}
}
