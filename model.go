package main

import (
	"os"
	"time"
)

// Base Movie type, with associated metadata
type Movie struct {
	Id          uint64
	Path        string
	Byte_length int64 //file length returned as signed
	Title       string
	Director    string
	Year        uint64
	Added_date  string //Much easier for serialization, and we only display it
	Watched     bool
	Hash        string //Not calculated by default. Stored as base64
}

// Creates a movie from the path with "blank" metadata
func NewMovie(path string) *Movie {
	info, _ := os.Lstat(path)
	return &Movie{
		Id:          0,
		Path:        path, //Assumes path is already absolute
		Byte_length: info.Size(),
		Title:       GetFilenameNoExt(path), //Treat filename excluding extension as title
		Director:    "",
		Year:        0,
		Added_date:  time.Now().Format("2006-01-02"),
		Watched:     false,
		Hash:        ""}
}

// Mutatively populate the hash of the given movie
func (m *Movie) populateHash() {
	m.Hash = HashFile(m.Path)
}
