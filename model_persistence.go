package main

import (
	"encoding/json"
	"errors"
	"github.com/fiatmoney/clownshoes"
	"strconv"
	"sync/atomic"
)

//Global state holding the open DB
var mainDB *clownshoes.DocumentBundle

//Atomically incremented for new insertions
var maxID uint64

// Initialize database if it hasn't been already.  Should only be called once.
func migrate(loc string) {
	mainDB = clownshoes.NewDB(loc)
	if !mainDB.HasIndexNamed("hash") {
		mainDB.AddIndex("hash", func(b []byte) string {
			var m Movie
			json.Unmarshal(b, &m)
			return m.Hash
		})
	}
	if !mainDB.HasIndexNamed("id") {
		mainDB.AddIndex("id", func(b []byte) string {
			var m Movie
			json.Unmarshal(b, &m)
			return strconv.FormatUint(m.Id, 10)
		})
	}
	if !mainDB.HasIndexNamed("path") {
		mainDB.AddIndex("path", func(b []byte) string {
			var m Movie
			json.Unmarshal(b, &m)
			return m.Path
		})
	}

	//Extract max ID
	docs := mainDB.GetDocuments(func(b []byte) bool { return true })
	maxID = 0
	for _, doc := range docs {
		var mv Movie
		json.Unmarshal(doc.Payload, &mv)
		if mv.Id > maxID {
			maxID = mv.Id
		}
	}
}

//Store without check, auto-generating ID
func StoreNew(m *Movie) error {
	id := atomic.AddUint64(&maxID, 1)
	m.Id = id
	asBytes, _ := json.Marshal(m)
	asDoc := clownshoes.NewDocument(asBytes)
	mainDB.PutDocument(asDoc)
	return nil
}

//Stores as a "new" movie in the given location, with existing metadata and a new ID
func StoreCopy(m *Movie, newPath string) error {
	m.Path = newPath
	return StoreNew(m)
}

func Modify(id uint64, field string, val string) error {
	switch field {
	case "title":
		mainDB.ReplaceDocumentsWhere("id", strconv.FormatUint(id, 10), func(b []byte) ([]byte, bool) {
			var mv Movie
			json.Unmarshal(b, &mv)
			mv.Title = val
			out, _ := json.Marshal(mv)
			return out, true
		})
		return nil
	case "watched":
		watched, _ := strconv.ParseBool(val)
		mainDB.ReplaceDocumentsWhere("id", strconv.FormatUint(id, 10), func(b []byte) ([]byte, bool) {
			var mv Movie
			json.Unmarshal(b, &mv)
			mv.Watched = watched
			out, _ := json.Marshal(mv)
			return out, true
		})
		return nil
	case "director":
		mainDB.ReplaceDocumentsWhere("id", strconv.FormatUint(id, 10), func(b []byte) ([]byte, bool) {
			var mv Movie
			json.Unmarshal(b, &mv)
			mv.Director = val
			out, _ := json.Marshal(mv)
			return out, true
		})
		return nil
	case "year":
		year, _ := strconv.ParseUint(val, 10, 64)
		mainDB.ReplaceDocumentsWhere("id", strconv.FormatUint(id, 10), func(b []byte) ([]byte, bool) {
			var mv Movie
			json.Unmarshal(b, &mv)
			mv.Year = year
			out, _ := json.Marshal(mv)
			return out, true
		})
		return nil
	}
	return errors.New("No valid update to map")
}

// Remove all movie records with the given path.
func Remove(path string) error {
	mainDB.RemoveDocumentsWhere("path", path, func(b []byte) bool { return true })
	return nil
}

func DumpDB() []Movie {
	docs := mainDB.GetDocuments(func(b []byte) bool {
		return true
	})
	mvs := make([]Movie, 0, len(docs))
	for _, doc := range docs {
		var mv Movie
		json.Unmarshal(doc.Payload, &mv)
		mvs = append(mvs, mv)
	}
	return mvs
}

//Remove any movie with a missing path
func RemoveWithInvalidPaths() error {
	mainDB.RemoveDocumentsWhere("path", "", func([]byte) bool { return true })
	mainDB.RemoveDocuments(func(b []byte) bool {
		var mv Movie
		json.Unmarshal(b, &mv)
		return !PathExists(mv.Path)
	})
	return nil
}

func GetMovieByPath(path string) *Movie {
	docs := mainDB.GetDocumentsWhere("path", path)
	if len(docs) > 0 {
		var m Movie
		json.Unmarshal(docs[0].Payload, &m)
		return &m
	}
	return nil
}

func GetMovieByHashAndSize(hash string, byte_length int64) *Movie {
	docs := mainDB.GetDocumentsWhere("hash", hash)
	for _, doc := range docs {
		var mv Movie
		json.Unmarshal(doc.Payload, &mv)
		if mv.Byte_length == byte_length {
			return &mv
		}
	}
	return nil
}

func GetMovieByID(id uint64) *Movie {
	docs := mainDB.GetDocumentsWhere("id", strconv.FormatUint(id, 10))
	if len(docs) > 0 {
		var mv Movie
		json.Unmarshal(docs[0].Payload, &mv)
		return &mv
	}
	return nil
}

//Set the given item's path to ""
func ClearPathByID(id uint64) error {
	mainDB.ReplaceDocumentsWhere("id", strconv.FormatUint(id, 10), func(b []byte) ([]byte, bool) {
		var mv Movie
		json.Unmarshal(b, &mv)
		mv.Path = ""
		out, _ := json.Marshal(mv)
		return out, true
	})
	return nil
}
