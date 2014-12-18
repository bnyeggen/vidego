package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

//Location of DB
var dbPath string

//All paths to be scanned
var scannerPaths []string

// This should take as command line parameters
// 1) the metadata DB path
// 2) one or more root directories for scanning
func main() {
	if len(os.Args) < 3 {
		panic("Must have 2+ arguments: metadata DB path & root directories for scanning")
	}
	//Parse arguments
	dbPath = os.Args[1]
	scannerPaths = os.Args[2:]

	//Migrate DB, sync on shutdown
	migrate(dbPath)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Println("Shutting down...")
			mainDB.Compact()
			mainDB.Sync()
			log.Println("Goodbye")
			os.Exit(0)
		}
	}()

	//Scan on startup
	scanAllPaths()
	log.Println("Ready")

	http.HandleFunc("/scan", scanHandler)

	http.HandleFunc("/json", jsonDumpHandler)
	http.HandleFunc("/update", movieUpdateHandler)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	//This surfaces at, eg, http://localhost:8080/movies/vol1/crapvid/Bernie.m4v
	//But this only works if the path has a trailing /
	for _, path := range scannerPaths {
		var nPath string
		if strings.HasSuffix(path, "/") {
			nPath = path
		} else {
			nPath = path + "/"
		}
		http.Handle("/movies"+nPath, http.StripPrefix("/movies"+nPath, http.FileServer(http.Dir(nPath))))
	}
	http.HandleFunc("/", displayAllHandler)
	http.ListenAndServe(":8080", nil)
}
