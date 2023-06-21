package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
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

/*
func saveFeed(feedUrl string) {
	sk := nostr.GeneratePrivateKey() // generate new key
	pub, _ := nostr.GetPublicKey(sk)
}
*/
