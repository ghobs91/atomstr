package main

import (
	"os"
	"time"
)

var homedir string = os.Getenv("HOME")
var editor string = os.Getenv("EDITOR")
var configLocation string = (homedir + "/" + ConfigDir + "/config.json")
var cacheLocation string = (homedir + "/" + CacheDir)
var feedsConfig string = "./feeds.ini"
var timezone, _ = time.Now().Zone()
var colorBlock string = "|"
var currentDot string = "•"
var Colors = [10]string{"\033[0;31m", "\033[0;32m", "\033[1;33m", "\033[1;34m", "\033[1;35m", "\033[1;36m", "\033[1;37m", "\033[1;38m", "\033[1;39m", "\033[1;40m"}
var showColor bool = true
var maxItemAgeHours time.Duration
var atomstrversion string = "0.1"

const (
	ConfigDir  = "config.yml"
	CacheDir   = ".cache/atomstr"
	dateFormat = "02.01.06"
)

type configStruct struct {
	Npub      string
	Nsec      string
	HomeRelay string
}

type eventStruct struct {
	Id       string
	Contacts []struct {
		p string
	}
	Relays string
}

type feedItemStruct struct {
	Title     string
	Published string
	Link      string
}
