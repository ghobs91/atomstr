package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

func checkEnv(key, defValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defValue
	}
	return val
}

func convertTimeString(itemTime string) *time.Time {
	// find right date format
	postTime, err := time.Parse(time.RFC3339, itemTime)
	if err != nil {
		postTime, err = time.Parse(time.RFC1123Z, itemTime) // try other one
	}
	return &postTime
}

func checkMaxAge(itemTime string, maxAgeHours time.Duration) bool {
	maxAge := time.Now().Add(-maxItemAgeHours * time.Hour)

	postTime := convertTimeString(itemTime)

	if postTime.After(maxAge) {
		return true
	}
	return false
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

func generateKeysForUrl(feedUrl string) *feedStruct {
	feedElem := feedStruct{}
	feedElem.Url = feedUrl

	feedElem.Sec = nostr.GeneratePrivateKey() // generate new key
	feedElem.Pub, _ = nostr.GetPublicKey(feedElem.Sec)

	return &feedElem
}
