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
	mvs := DumpDB()
	movies, _ := json.Marshal(mvs)
	w.Write(movies)
}

func scanHandler(w http.ResponseWriter, r *http.Request) {
	scanAllPaths()
}

// Receives updates to movies via PUT
func movieUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id := r.Form.Get("id")
	idAsInt, e := strconv.ParseUint(id, 10, 64)
	if e != nil {
		log.Println(e)
	}
	switch r.Method {
	case "PUT":
		field := strings.ToLower(r.Form.Get("field"))
		val := r.Form.Get("val")
		e := Modify(idAsInt, field, val)
		if e != nil {
			log.Println(e)
		}
		w.WriteHeader(http.StatusNoContent)
	case "GET":
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
	mvs := DumpDB()
	t.Execute(w, mvs)
}
