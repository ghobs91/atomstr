package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func processFeeds(db *sql.DB) {
	feeds := dbGetAllFeeds(db)
	if len(*feeds) == 0 {
		log.Println("[WARN] No feeds found")
		log.Fatal("no feeds")
	}
	//fmt.Println(feeds)
	log.Println("[INFO] Updating feeds")
	for _, feedItem := range *feeds { // FIXME: error handling
		processFeedUrl(&feedItem)
	}
	log.Println("[INFO] Finished updating feeds")
}

func main() {
	db := dbInit()
	logger()

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
		//fmt.Println(fetchInterval)
		//os.Exit(1)
		// first run
		nostrUpdateAllFeedsMetadata(db)
		processFeeds(db)

		metadataTicker := time.NewTicker(metadataInterval)
		updateTicker := time.NewTicker(fetchInterval)

		cancelChan := make(chan os.Signal, 1)
		// catch SIGETRM or SIGINTERRUPT
		signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			for {
				select {
				case <-metadataTicker.C:
					nostrUpdateAllFeedsMetadata(db)
				case <-updateTicker.C:
					processFeeds(db)
				}
			}
		}()
		sig := <-cancelChan

		log.Println("[DEBUG] Caught signal %v", sig)
		metadataTicker.Stop()
		updateTicker.Stop()
		log.Println("[INFO] Closing DB")
		db.Close()
		log.Println("[INFO] Shutting down")

	}

}
