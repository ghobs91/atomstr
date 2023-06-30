package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func (a *Atomstr) dbGetAllFeeds() *[]feedStruct {
	sqlStatement := `SELECT pub, sec, url FROM feeds`
	rows, err := a.db.Query(sqlStatement)
	if err != nil {
		log.Fatal("[ERROR] Returning feeds from DB failed")
	}

	feedItems := []feedStruct{}

	for rows.Next() {
		feedItem := feedStruct{}
		if err := rows.Scan(&feedItem.Pub, &feedItem.Sec, &feedItem.Url); err != nil {
			log.Fatal("[ERROR] Scanning for feeds failed")
		}
		feedItem.Npub, _ = nip19.EncodePublicKey(feedItem.Pub)
		feedItems = append(feedItems, feedItem)
	}

	return &feedItems
}

func nostrUpdateFeedMetadata(feedItem *feedStruct) {
	//fmt.Println(feedItem)

	metadata := map[string]string{
		"name":    feedItem.Title + " (RSS Feed)",
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
	log.Println("[DEBUG] Updating feed metadata for", feedItem.Title)

	nostrPostItem(ev)
}

func (a *Atomstr) nostrUpdateAllFeedsMetadata() {
	feeds := a.dbGetAllFeeds()

	log.Println("[INFO] Updating feeds metadata")
	for _, feedItem := range *feeds {
		data := checkValidFeedSource(feedItem.Url)
		if data.Title == "" {
			log.Println("[ERROR] error updating feed")
			continue
		}
		feedItem.Title = data.Title
		feedItem.Description = data.Description
		feedItem.Link = data.Link
		feedItem.Image = data.Image
		nostrUpdateFeedMetadata(&feedItem)
	}
	log.Println("[INFO] Finished updating feeds metadata")
}

func processFeedUrl(feedItem *feedStruct) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // fetch feeds with 10s timeout
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedItem.Url, ctx)
	if err != nil {
		log.Println("[ERROR] Can't update feed")
	} else {
		log.Println("[DEBUG] Updating feed ", feedItem.Url)
		feedItem.Title = feed.Title
		feedItem.Description = feed.Description
		feedItem.Link = feed.Link
		if feed.Image != nil {
			feedItem.Image = feed.Image.URL
		} else {
			feedItem.Image = defaultFeedImage
		}
		//feedItem.Image = feed.Image

		for i := range feed.Items {
			processFeedPost(feedItem, feed.Items[i])
		}
		log.Println("[DEBUG] Finished updating feed ", feedItem.Url)
	}
}

func processFeedPost(feedItem *feedStruct, feedPost *gofeed.Item) {
	// if time right, then push
	p := bluemonday.StripTagsPolicy() // initialize html sanitizer

	//fmt.Println(feedPost.Published)
	if checkMaxAge(feedPost.Published, fetchInterval) {
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

		log.Printf("[DEBUG] Event published to %s\n", url)
	}
}

func (a *Atomstr) dbWriteFeed(feedItem *feedStruct) bool {
	_, err := a.db.Exec(`insert into feeds (pub, sec, url) values(?, ?, ?)`, feedItem.Pub, feedItem.Sec, feedItem.Url)
	if err != nil {
		fmt.Println("[ERROR] Can't add feed!")
		log.Fatal(err)
	}
	nip19Pub, _ := nip19.EncodePublicKey(feedItem.Pub)
	log.Println("[INFO] Added feed " + feedItem.Url + " with public key " + nip19Pub)
	return true
}

func (a *Atomstr) dbGetFeed(feedUrl string) *feedStruct {
	sqlStatement := `SELECT pub, sec, url FROM feeds WHERE url=$1;`
	row := a.db.QueryRow(sqlStatement, feedUrl)

	feedItem := feedStruct{}
	err := row.Scan(&feedItem.Pub, &feedItem.Sec, &feedItem.Url)

	if err != nil {
		log.Println("[INFO] Feed not found in DB")
	}
	return &feedItem
}

func checkValidFeedSource(feedUrl string) *feedStruct {
	log.Println("[DEBUG] Trying to find feed at", feedUrl)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feedUrl, ctx)

	if err != nil {
		log.Println("[ERROR] Not a valid feed source")
	}
	// FIXME! That needs proper error handling.
	feedItem := feedStruct{}
	feedItem.Title = feed.Title
	feedItem.Description = feed.Description
	feedItem.Link = feed.Link
	if feed.Image != nil {
		feedItem.Image = feed.Image.URL
	} else {
		feedItem.Image = defaultFeedImage
	}

	return &feedItem
}

func (a *Atomstr) addSource(feedUrl string) *feedStruct {
	//var feedElem2 *feedStruct
	feedItem := checkValidFeedSource(feedUrl)
	if feedItem.Title == "" {
		log.Println("[ERROR] No valid feed found on", feedUrl)
		log.Fatal("nope")
	}

	// check for existing feed
	feedTest := a.dbGetFeed(feedUrl)
	if feedTest.Url != "" {
		log.Println("[WARN] Feed already exists")
		log.Fatal()
	}

	feedItem = generateKeysForUrl(feedUrl)

	a.dbWriteFeed(feedItem)

	return feedItem
}
func (a *Atomstr) deleteSource(feedUrl string) bool {
	// check for existing feed
	feedTest := a.dbGetFeed(feedUrl)
	if feedTest.Url != "" {
		sqlStatement := `DELETE FROM feeds WHERE url=$1;`
		_, err := a.db.Exec(sqlStatement, feedUrl)
		if err != nil {
			log.Println("[WARN] Can't remove Feed")
			log.Fatal(err)
		}
		log.Println("[INFO] Feed removed")
		return true
	} else {
		log.Println("[WARN] Feed not found")
		return false
	}
}

func (a *Atomstr) listFeeds() {
	feeds := a.dbGetAllFeeds()

	for _, feedItem := range *feeds {
		nip19Pub, _ := nip19.EncodePublicKey(feedItem.Pub)
		fmt.Print(nip19Pub + " ")
		fmt.Println(feedItem.Url)
	}

}
