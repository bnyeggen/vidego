package main

import (
	"os"
	"time"
)

// Base Movie type, with associated metadata
type Movie struct {
	Id          int64
	Path        string
	Byte_length int64
	Title       string
	Director    string
	Year        uint16
	Added_date  time.Time //jsonizes as "0001-01-01T00:00:00Z"
	Watched     bool
	Hash        string //Not calculated by default. Stored as base64
}

// Creates a movie from the path with "blank" metadata
func NewMovie(path string) *Movie {
	info, _ := os.Lstat(path)
	return &Movie{
		Id:          -1,
		Path:        path, //Assumes path is already absolute
		Byte_length: info.Size(),
		Title:       GetFilenameNoExt(path), //Treat filename excluding extension as title
		Director:    "",
		Year:        0,
		Added_date:  time.Now(),
		Watched:     false,
		Hash:        ""}
}

// Mutatively populate the hash of the given movie
func (m *Movie) populateHash() {
	m.Hash = HashFile(m.Path)
}
