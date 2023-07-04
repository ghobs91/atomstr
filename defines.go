package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

var fetchInterval, _ = time.ParseDuration(getEnv("FETCH_INTERVAL", "15m"))
var metadataInterval, _ = time.ParseDuration(getEnv("METADATA_INTERVAL", "2h"))
var logLevel = getEnv("LOG_LEVEL", "INFO")
var webserverPort = getEnv("WEBSERVER_PORT", "8061")
var nip05Domain = getEnv("NIP05_DOMAIN", "atomstr.data.haus")
var maxWorkers, _ = strconv.Atoi(getEnv("MAX_WORKERS", "5"))
var r = getEnv("RELAYS_TO_PUBLISH_TO", "wss://nostr.data.haus")
var relaysToPublishTo = strings.Split(r, ",")
var defaultFeedImage = getEnv("DEFAULT_FEED_IMAGE", "https://void.cat/d/NDrSDe4QMx9jh6bD9LJwcK")
var dbPath = getEnv("DB_PATH", "./atomstr.db")
var atomstrversion string = "0.3"

type Atomstr struct {
	db *sql.DB
}

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
	Npub        string
	Title       string
	Description string
	Link        string
	Image       string
}

type webIndex struct {
	Relays []string
	Feeds  []feedStruct
}
type webAddFeed struct {
	Status string
	Feed   feedStruct
}
