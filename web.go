package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/nbd-wtf/go-nostr/nip05"
)

func (a *Atomstr) webMain(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.tmpl"))
	feeds := a.dbGetAllFeeds()
	tmpl.Execute(w, *feeds)
}

func (a *Atomstr) webAdd(w http.ResponseWriter, r *http.Request) {
	//tmpl := template.Must(template.ParseFiles("templates/add.tmpl"))
	feedItem := a.addSource(r.FormValue("url"))

	if feedItem.Pub != "" {
		fmt.Fprintln(w, "<h1>Success!</h1><p>Added feed "+r.FormValue("url")+" with public key "+feedItem.Pub+"</p>")
	} else {
		fmt.Fprintln(w, "<h1>No feed found or feed already exists.</h1>")
	}

	fmt.Fprintln(w, "<p><a href=..>Go back</a></p>")

	//tmpl.Execute(w, *feeds)
}

func (a *Atomstr) webNip05(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	name, _ = url.QueryUnescape(name)
	w.Header().Set("Content-Type", "application/json")

	var response []byte
	if name != "" && name != "_" {
		feedItem := a.dbGetFeed(name)

		nip05WellKnownResponse := nip05.WellKnownResponse{
			Names: map[string]string{
				name: feedItem.Pub,
			},
			Relays: nil,
		}
		response, _ = json.Marshal(nip05WellKnownResponse)
		_, _ = w.Write(response)
	}
}

func (a *Atomstr) webserver() {
	http.HandleFunc("/", a.webMain)
	http.HandleFunc("/add", a.webAdd)
	http.HandleFunc("/.well-known/nostr.json", a.webNip05)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("[INFO] Starting webserver at port", webserverPort)
	log.Fatal(http.ListenAndServe(":"+webserverPort, nil))
}
