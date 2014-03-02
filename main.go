package main

import (
	"log"
	"net/http"
	"os"
)

//Location of sqlite DB
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

	//Migrate DB, close on shutdown
	mainDB = migrate(dbPath)
	defer mainDB.Close()

	//Scan on startup
	scanAllPaths()
	log.Println("Ready")

	http.HandleFunc("/", displayAllHandler)
	http.HandleFunc("/json", jsonDumpHandler)
	http.HandleFunc("/update", movieUpdateHandler)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.ListenAndServe(":8080", nil)
}
