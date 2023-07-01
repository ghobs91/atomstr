package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func (a *Atomstr) processFeeds() {
	feeds := a.dbGetAllFeeds()
	if len(*feeds) == 0 {
		log.Println("[WARN] No feeds found")
		//log.Fatal("no feeds")
	}
	//fmt.Println(feeds)
	log.Println("[INFO] Updating feeds")
	/*
		for _, feedItem := range *feeds { // FIXME: error handling
			processFeedUrl(&feedItem)
		}*/

	// create a channel for work "tasks"
	ch := make(chan *feedStruct)

	wg := sync.WaitGroup{}

	// start the workers
	for t := 0; t < maxWorkers; t++ {
		wg.Add(1)
		go processFeedUrl(ch, &wg)
	}

	// push the lines to the queue channel for processing
	for _, feedItem := range *feeds {
		ch <- &feedItem
	}

	// this will cause the workers to stop and exit their receive loop
	close(ch)

	// make sure they all exit
	wg.Wait()

	log.Println("[INFO] Finished updating feeds")
}

func main() {
	//db := dbInit()
	a := &Atomstr{db: dbInit()}

	logger()

	feedNew := flag.String("a", "", "Add a new URL to scrape")
	feedDelete := flag.String("d", "", "Remove a feed from db")
	flag.Bool("l", false, "List all feeds with npubs")
	flag.Parse()
	flagset := make(map[string]bool) // map for flag.Visit. get bools to determine set flags
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if flagset["a"] {
		a.addSource(*feedNew)
	} else if flagset["l"] {
		a.listFeeds()
	} else if flagset["d"] {
		a.deleteSource(*feedDelete)
	} else {
		//fmt.Println(fetchInterval)
		//os.Exit(1)
		go a.webserver()
		// first run
		a.nostrUpdateAllFeedsMetadata()
		a.processFeeds()

		metadataTicker := time.NewTicker(metadataInterval)
		updateTicker := time.NewTicker(fetchInterval)

		cancelChan := make(chan os.Signal, 1)
		// catch SIGETRM or SIGINTERRUPT
		signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			for {
				select {
				case <-metadataTicker.C:
					a.nostrUpdateAllFeedsMetadata()
				case <-updateTicker.C:
					a.processFeeds()
				}
			}
		}()
		sig := <-cancelChan

		log.Println("[DEBUG] Caught signal %v", sig)
		metadataTicker.Stop()
		updateTicker.Stop()
		log.Println("[INFO] Closing DB")
		a.db.Close()
		log.Println("[INFO] Shutting down")

	}

}
