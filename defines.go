package main

import (
	"database/sql"
	"time"
)

var maxItemAgeHours time.Duration = 1
var fetchIntervalMinutes time.Duration = 15
var atomstrversion string = "0.1"
var relaysToPublishTo = [...]string{"wss://nostr.data.haus"}

const (
	dbPath           = "./atomstr.db"
	defaultFeedImage = "https://void.cat/d/NDrSDe4QMx9jh6bD9LJwcK"
)

type atomstr struct {
	Secret                        string   `envconfig:"SECRET" required:"true"`
	dbPath2                       string   `envconfig:"DB_DIR" default:"./atomstr.db2"`
	DefaultProfilePictureUrl      string   `envconfig:"DEFAULT_PROFILE_PICTURE_URL" default:"https://i.imgur.com/MaceU96.png"`
	RelaysToPublish               []string `envconfig:"RELAYS_TO_PUBLISH_TO" default:"wss://nostr.data.haus"`
	DefaultWaitTimeBetweenBatches int64    `envconfig:"DEFAULT_WAIT_TIME_BETWEEN_BATCHES" default:"60000"`
	EnableAutoNIP05Registration   bool     `envconfig:"ENABLE_AUTO_NIP05_REGISTRATION" default:"false"`
	MainDomainName                string   `envconfig:"MAIN_DOMAIN_NAME" default:""`

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
	Title       string
	Description string
	Link        string
	Image       string
}
