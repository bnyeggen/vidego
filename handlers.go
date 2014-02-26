package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//Response handlers that supply templates or invoke fns with appropraite args.
//See
//http://golang.org/doc/articles/wiki/

//Basically these are the controllers.

// Dumps present DB to JSON
func jsonDumpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mvs, e := DumpDB()
	if e != nil {
		log.Println(e)
	}
	movies, _ := json.Marshal(mvs)
	w.Write(movies)
}

// Receives updates to movies via PUT
// TODO: Return individual movies via GET
func movieUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.Form.Get("id")
	field := strings.ToLower(r.Form.Get("field"))
	val := r.Form.Get("val")

	switch r.Method {
	case "PUT":
		update := map[string]interface{}{
			"id":  id,
			field: val,
		}
		e := ModifyWithMap(update)
		if e != nil {
			log.Println(e)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Renders movies as a pretty web page
func displayAllHandler(w http.ResponseWriter, r *http.Request) {
	//Currently this doesn't actually require any templating
	t, _ := template.ParseFiles("resources/index.html")
	mvs, e := DumpDB()
	if e != nil {
		log.Println(e)
	}
	t.Execute(w, mvs)
}

func scanAllPaths() {
	for _, path := range scannerPaths {
		for _, mv := range ScanForMovies(path) {
			//These could be goroutines, but this should be IO bound
			//segments with little advantage to concurrency
			log.Println(mv.Path)
			CheckAndStore(&mv)
		}
	}
	//Then needs to purge all nonexistent movies
	paths, _ := DumpAllPaths()
	for _, path := range paths {
		if !PathExists(path) {
			Remove(path)
		}
	}
}

// Initiates a scan of the paths supplied on startup
func scanHandler(w http.ResponseWriter, r *http.Request) {
	scanAllPaths()
}
