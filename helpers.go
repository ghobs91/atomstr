package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/nbd-wtf/go-nostr"
)

func getConf() *configStruct {
	configData, err := ioutil.ReadFile(configLocation)
	if err != nil {
		fmt.Print("Config not found. \n\nPlease copy config-sample.json to ~/.config/nstr/config.json and modify it accordingly.\n\n")
		log.Fatal(err)
	}
	conf := configStruct{}
	err = json.Unmarshal(configData, &conf)
	//fmt.Println(conf)
	if err != nil {
		log.Fatal(err)
	}

	return &conf
}
func getFeedUrls() ([]string, error) {
	file, err := os.Open(feedsConfig)
	if err != nil {
		fmt.Print("feeds file not found. \n\nPlease copy feeds-sample.json to ./feeds.json and modify it accordingly.\n\n")
		log.Fatal(err)
	}
	defer file.Close()
	var feedUrls []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		feedUrls = append(feedUrls, scanner.Text())
	}

	return feedUrls, scanner.Err()
}

func dbGetAllFeeds(db *sql.DB) *[]feedStruct {
	sqlStatement := `SELECT pub, sec, url FROM feeds`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal("[ERROR]: Returning feeds")
	}

	feedItems := []feedStruct{}

	for rows.Next() {
		feedItem := feedStruct{}
		if err := rows.Scan(&feedItem.Pub, &feedItem.Sec, &feedItem.Url); err != nil {
			log.Fatal("[ERROR]: Scanning for feeds")
		}
		feedItems = append(feedItems, feedItem)
	}

	return &feedItems
}

func fetchFeedData(feedUrl string) {
	//p := bluemonday.StripTagsPolicy() // initialize html sanitizer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // fetch feeds with 10s timeout
	defer cancel()
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURLWithContext(feedUrl, ctx)
	for i := range feed.Items {
		processFeedItem(feed.Items[i])
	}
}

func checkMaxAge(itemTime string, maxAgeHours time.Duration) bool {
	maxAge := time.Now().Add(-maxItemAgeHours * time.Hour)
	// find right date format
	postTime, err := time.Parse(time.RFC3339, itemTime)
	if err != nil {
		postTime, err = time.Parse(time.RFC1123Z, itemTime) // try other one
	}

	if postTime.After(maxAge) {
		return true
	}
	return false
}

func processFeedItem(feedItem *gofeed.Item) {
	// if time right, then push
	maxItemAgeHours = 24 // TODO: config

	if checkMaxAge(feedItem.Published, maxItemAgeHours) {
		//fmt.Println(myTime)
		fmt.Print(feedItem.Published + " ")
		fmt.Println(feedItem.Title)
		/*fmt.Println()
		desc := p.Sanitize(feed.Items[i].Description)
		fmt.Println(desc)
		fmt.Println(feed.Items[i].Link)
		fmt.Println()*/
	}
}

func dbInit() *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("[FATAL] open db: %v", err)
	}
	log.Printf("[INFO] database opened at %s", dbPath)
	//defer db.Close()

	_, err = db.Exec(sqlInit)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlInit)
	}

	return db
}

func dbWriteFeed(db *sql.DB, feedItem *feedStruct) bool {
	_, err := db.Exec(`insert into feeds (pub, sec, url) values(?, ?, ?)`, feedItem.Pub, feedItem.Sec, feedItem.Url)
	if err != nil {
		fmt.Println("ERROR: Can't add feed!")
		log.Fatal(err)
	}
	fmt.Println("INFO: Added feed " + feedItem.Url)
	return true
}

func dbGetFeed(db *sql.DB, feedUrl string) *feedStruct {
	sqlStatement := `SELECT pub, sec, url FROM feeds WHERE url=$1;`
	row := db.QueryRow(sqlStatement, feedUrl)

	feedItem := feedStruct{}
	err := row.Scan(&feedItem.Pub, &feedItem.Sec, &feedItem.Url)

	if err != nil {
		fmt.Println("[INFO]: Feed not found in DB")
	}
	return &feedItem
}

func generateKeysForUrl(feedUrl string) *feedStruct {
	feedElem := feedStruct{}
	feedElem.Url = feedUrl

	feedElem.Sec = nostr.GeneratePrivateKey() // generate new key
	feedElem.Pub, _ = nostr.GetPublicKey(feedElem.Sec)

	return &feedElem
}

func checkValidFeedSource(feedUrl string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedUrl, ctx)
	if err != nil {
		fmt.Println("ERROR: Not a valid feed source!")
		log.Fatal(err)
	}

	return feed.Title
}

func addSource(db *sql.DB, feedUrl string) *feedStruct {
	//var feedElem2 *feedStruct
	checkValidFeedSource(feedUrl)

	// check for existing feed
	feedTest := dbGetFeed(db, feedUrl)
	if feedTest.Url != "" {
		fmt.Println("WARN: Feed already exists")
		log.Fatal()
	}

	feedElem := generateKeysForUrl(feedUrl)

	dbWriteFeed(db, feedElem)

	return feedElem
}

/*
func saveFeed(feedUrl string) {
	sk := nostr.GeneratePrivateKey() // generate new key
	pub, _ := nostr.GetPublicKey(sk)
}
*/
