package main

import (
	"database/sql"
	"fmt"
)

type Relay struct {
	Secret                        string   `envconfig:"SECRET" required:"true"`
	DatabaseDirectory             string   `envconfig:"DB_DIR" default:"db/rsslay.sqlite"`
	DefaultProfilePictureUrl      string   `envconfig:"DEFAULT_PROFILE_PICTURE_URL" default:"https://i.imgur.com/MaceU96.png"`
	RelaysToPublish               []string `envconfig:"RELAYS_TO_PUBLISH_TO" default:"wss://nostr.data.haus"`
	DefaultWaitTimeBetweenBatches int64    `envconfig:"DEFAULT_WAIT_TIME_BETWEEN_BATCHES" default:"60000"`
	EnableAutoNIP05Registration   bool     `envconfig:"ENABLE_AUTO_NIP05_REGISTRATION" default:"false"`
	MainDomainName                string   `envconfig:"MAIN_DOMAIN_NAME" default:""`

	db *sql.DB
}

/*
func InitDatabase(r *Relay) *sql.DB {
	finalConnection := dsn
	if *dsn == "" {
		log.Print("[INFO] dsn required is not present... defaulting to DB_DIR")
		finalConnection = &r.DatabaseDirectory
	}

	// Create empty dir if not exists
	dbPath := path.Dir(*finalConnection)
	err := os.MkdirAll(dbPath, 0660)
	if err != nil {
		log.Printf("[INFO] unable to initialize DB_DIR at: %s. Error: %v", dbPath, err)
	}

	// Connect to SQLite database.
	sqlDb, err := sql.Open("sqlite3", *finalConnection)
	if err != nil {
		log.Fatalf("[FATAL] open db: %v", err)
	}

	log.Printf("[INFO] database opened at %s", *finalConnection)

	return sqlDb

}*/

func getFeeds() {
	feeds, _ := getFeedUrls()
	fmt.Println(feeds)
	for _, url := range feeds {
		fetchFeedData(url)
	}
	/*
		sort.Slice(elements, func(i, j int) bool {
			return elements[i].Start.Before(elements[j].Start) // time.Time sort by start time for events
		})

		if len(elements) == 0 {
			log.Fatal("no events") // get out if nothing found
		}

		for _, e := range elements {
			e.fancyOutput() // pretty print
		}
	*/
}

func main() {

	getFeeds()

}
