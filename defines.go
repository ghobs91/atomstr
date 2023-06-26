package main

import (
	"time"
)

var maxItemAge, _ = time.ParseDuration(getEnv("MAX_ITEM_AGE", "1h"))
var fetchInterval, _ = time.ParseDuration(getEnv("FETCH_INTERVAL", "15m"))
var metadataInterval, _ = time.ParseDuration(getEnv("METADATA_INTERVAL", "2h"))
var logLevel = getEnv("LOG_LEVEL", "INFO")
var atomstrversion string = "0.1"
var relaysToPublishTo = []string{"wss://nostr.data.haus"}

const (
	dbPath           = "./atomstr.db"
	defaultFeedImage = "https://void.cat/d/NDrSDe4QMx9jh6bD9LJwcK"
)

var sqlInit = `
CREATE TABLE IF NOT EXISTS feeds (
	pub VARCHAR(64) PRIMARY KEY,
	sec VARCHAR(64) NOT NULL,
	url TEXT NOT NULL
);
`

type feedStruct struct {
	Url         string
	Sec         string
	Pub         string
	Title       string
	Description string
	Link        string
	Image       string
}
