package main

import (
	"database/sql"
	"flag"

	_ "github.com/mattn/go-sqlite3"
)

type Relay struct {
	Secret                        string   `envconfig:"SECRET" required:"true"`
	DatabaseDirectory             string   `envconfig:"DB_DIR" default:"./atomstr.db2"`
	DefaultProfilePictureUrl      string   `envconfig:"DEFAULT_PROFILE_PICTURE_URL" default:"https://i.imgur.com/MaceU96.png"`
	RelaysToPublish               []string `envconfig:"RELAYS_TO_PUBLISH_TO" default:"wss://nostr.data.haus"`
	DefaultWaitTimeBetweenBatches int64    `envconfig:"DEFAULT_WAIT_TIME_BETWEEN_BATCHES" default:"60000"`
	EnableAutoNIP05Registration   bool     `envconfig:"ENABLE_AUTO_NIP05_REGISTRATION" default:"false"`
	MainDomainName                string   `envconfig:"MAIN_DOMAIN_NAME" default:""`

	db *sql.DB
}

/*
	func getFeeds(db *sql.DB) {
		feeds, _ := getFeedUrls()
		fmt.Println(feeds)
		for _, url := range feeds {
			fetchFeedData(url)
		}
			sort.Slice(elements, func(i, j int) bool {
				return elements[i].Start.Before(elements[j].Start) // time.Time sort by start time for events
			})

			if len(elements) == 0 {
				log.Fatal("no events") // get out if nothing found
			}

			for _, e := range elements {
				e.fancyOutput() // pretty print
			}
	}
*/
func getFeeds(db *sql.DB) {
	feeds := dbGetAllFeeds(db)
	//fmt.Println(feeds)
	for _, feedItem := range *feeds {
		fetchFeedData(feedItem.Url)
	}
}

func main() {
	db := dbInit()

	newRSS := flag.String("a", "", "Add a new URL to scrape")
	flag.Parse()
	flagset := make(map[string]bool) // map for flag.Visit. get bools to determine set flags
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if flagset["a"] {
		addSource(db, *newRSS)
	} else {
		getFeeds(db)
		//addSource(db, "https://rss.dw.com/atom/rss-de-all")
	}

	db.Close()

}
