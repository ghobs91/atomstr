package main

import (
	"database/sql"
	"flag"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func processFeeds(db *sql.DB) {
	feeds := dbGetAllFeeds(db)
	if len(*feeds) == 0 {
		log.Println("[INFO] No feeds found")
		log.Fatal("no feeds")
	}
	//fmt.Println(feeds)
	log.Println("[INFO] Updating feeds")
	for _, feedItem := range *feeds {
		processFeedUrl(&feedItem)
	}
	log.Println("[INFO] Finished updating feeds")
}

func main() {
	db := dbInit()

	feedNew := flag.String("a", "", "Add a new URL to scrape")
	flag.Bool("l", false, "List all feeds with npubs")
	flag.Parse()
	flagset := make(map[string]bool) // map for flag.Visit. get bools to determine set flags
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if flagset["a"] {
		addSource(db, *feedNew)
	} else if flagset["l"] {
		listFeeds(db)
	} else {
		// first run
		nostrUpdateAllFeedsMetadata(db)
		processFeeds(db)

		metadataTicker := time.NewTicker(time.Hour * 1)
		updateTicker := time.NewTicker(time.Minute * fetchIntervalMinutes)

		for {
			select {
			case <-metadataTicker.C:
				nostrUpdateAllFeedsMetadata(db)
			case <-updateTicker.C:
				processFeeds(db)
			}
		}

	}

	log.Println("[INFO] Closing DB")
	db.Close()
	log.Println("[INFO] Shutting down")

}
