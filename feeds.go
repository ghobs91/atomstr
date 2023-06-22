package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

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

func nostrUpdateFeed(feedItem *feedStruct) {
	//fmt.Println(feedItem)

	metadata := map[string]string{
		"name":    feedItem.Title + " (datahaus testing RSS Feed)",
		"about":   feedItem.Description + "\n\n" + feedItem.Link,
		"picture": feedItem.Image,
	}

	content, _ := json.Marshal(metadata)

	ev := nostr.Event{
		PubKey:    feedItem.Pub,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindSetMetadata,
		Tags:      nostr.Tags{},
		Content:   string(content),
	}
	ev.ID = string(ev.Serialize())
	ev.Sign(feedItem.Sec)
	log.Println("[INFO] Updating feed metadata")

	nostrPostItem(ev)
}

func processFeedUrl(feedItem *feedStruct) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // fetch feeds with 10s timeout
	defer cancel()
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURLWithContext(feedItem.Url, ctx)
	feedItem.Title = feed.Title
	feedItem.Description = feed.Description
	feedItem.Link = feed.Link
	if feed.Image != nil {
		feedItem.Image = feed.Image.URL
	} else {
		feedItem.Image = defaultFeedImage
	}
	//feedItem.Image = feed.Image
	nostrUpdateFeed(feedItem)
	for i := range feed.Items {
		processFeedPost(feedItem, feed.Items[i])
	}
}

func processFeedPost(feedItem *feedStruct, feedPost *gofeed.Item) {
	// if time right, then push
	p := bluemonday.StripTagsPolicy() // initialize html sanitizer

	if checkMaxAge(feedPost.Published, maxItemAgeHours) {
		feedText := feedPost.Title + "\n\n" + p.Sanitize(feedPost.Description)
		if feedPost.Link != "" {
			feedText = feedText + "\n\n" + feedPost.Link
		}
		postTime := convertTimeString(feedPost.Published)

		ev := nostr.Event{
			PubKey:    feedItem.Pub,
			CreatedAt: nostr.Timestamp(postTime.Unix()),
			Kind:      nostr.KindTextNote,
			Tags:      nil,
			Content:   feedText,
		}

		ev.Sign(feedItem.Sec)

		nostrPostItem(ev)
		/*
			nip19Pub, _ := nip19.EncodePublicKey(feedItem.Pub)
			fmt.Print(feedItem.Url + " ")
			fmt.Print(nip19Pub + " ")
			fmt.Println(postTime.Format(time.RFC3339) + "\n")
		*/
		//fmt.Println(feedText)
	}
}

func nostrPostItem(ev nostr.Event) {
	ctx := context.Background()
	//for _, url := range []string{"wss://nostr.data.haus", "wss://nostr-pub.wellorder.net"} {
	for _, url := range relaysToPublishTo {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = relay.Publish(ctx, ev)
		if err != nil {
			fmt.Println(err)
			continue
		}

		log.Printf("[INFO] Event published to %s\n", url)
	}
}

func dbWriteFeed(db *sql.DB, feedItem *feedStruct) bool {
	_, err := db.Exec(`insert into feeds (pub, sec, url) values(?, ?, ?)`, feedItem.Pub, feedItem.Sec, feedItem.Url)
	if err != nil {
		fmt.Println("[ERROR] Can't add feed!")
		log.Fatal(err)
	}
	nip19Pub, _ := nip19.EncodePublicKey(feedItem.Pub)
	log.Println("[INFO] Added feed " + feedItem.Url + "with public key " + nip19Pub)
	return true
}

func dbGetFeed(db *sql.DB, feedUrl string) *feedStruct {
	sqlStatement := `SELECT pub, sec, url FROM feeds WHERE url=$1;`
	row := db.QueryRow(sqlStatement, feedUrl)

	feedItem := feedStruct{}
	err := row.Scan(&feedItem.Pub, &feedItem.Sec, &feedItem.Url)

	if err != nil {
		log.Println("[INFO] Feed not found in DB")
	}
	return &feedItem
}

func checkValidFeedSource(feedUrl string) string {
	log.Println("[INFO] Trying to find feed at " + feedUrl)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedUrl, ctx)
	fmt.Println(feedUrl)
	fmt.Println(feed.Title)
	if err != nil {
		log.Println("[ERROR] Not a valid feed source")
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
		log.Println("[WARN] Feed already exists")
		log.Fatal()
	}

	feedItem := generateKeysForUrl(feedUrl)

	dbWriteFeed(db, feedItem)

	return feedItem
}

func listFeeds(db *sql.DB) {
	feeds := dbGetAllFeeds(db)

	for _, feedItem := range *feeds {
		nip19Pub, _ := nip19.EncodePublicKey(feedItem.Pub)
		fmt.Print(nip19Pub + " ")
		fmt.Println(feedItem.Url)
	}

}
