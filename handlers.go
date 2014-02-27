package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
func movieUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.Form.Get("id")

	switch r.Method {
	case "PUT":
		field := strings.ToLower(r.Form.Get("field"))
		val := r.Form.Get("val")
		update := map[string]interface{}{
			"id":  id,
			field: val,
		}
		e := ModifyWithMap(update)
		if e != nil {
			log.Println(e)
		}
		w.WriteHeader(http.StatusNoContent)
	case "GET":
		idAsInt, e := strconv.ParseInt(id, 10, 64)
		if e != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mv := GetMovieByID(idAsInt)
		if mv == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mvAsJSON, _ := json.Marshal(mv)
		w.Header().Set("Content-Type", "application/json")
		w.Write(mvAsJSON)
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
			e := CheckAndStore(&mv)
			if e != nil {
				log.Println(e)
			}
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
