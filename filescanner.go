package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Utilities for scanning the filesystem and turning movie files into Movie objects

// Discriminates video files from eg subtitles, notes, thumbnails...
var AcceptedFiletypes = map[string]bool{
	"m4v":  true,
	"mp4":  true,
	"avi":  true,
	"flv":  true,
	"swf":  true,
	"mpg":  true,
	"mpeg": true,
	"mkv":  true}

// The md5 hash of the first 100mbytes of a file
func HashFile(path string) string {
	fd, err := os.Open(path)
	defer fd.Close()
	if err != nil {
		return ""
	}
	hasher := md5.New()
	//Unlikely to be a collision over the first 16mb
	//This doesn't depend on files being < this threshold
	io.CopyN(hasher, fd, 100000000)
	return hex.EncodeToString(hasher.Sum(nil))
}

// Return the file extension of the given path.  "Hello.mkv"->"mkv"
func GetFileType(path string) string {
	split := strings.Split(filepath.Base(path), ".")
	return split[len(split)-1]
}

// Return the filename denuded of extension. "Hello.world.mkv"->"Hello.world"
func GetFilenameNoExt(path string) string {
	split := strings.Split(filepath.Base(path), ".")
	return strings.Join(split[0:len(split)-1], ".")
}

// Base Movie type, with associated metadata
type Movie struct {
	Id          int64
	Path        string
	Byte_length int64
	Title       string
	Director    string
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
		Added_date:  time.Now(),
		Watched:     false,
		Hash:        ""}
}

// Mutatively populate the hash of the given movie
func (m *Movie) populateHash() {
	m.Hash = HashFile(m.Path)
}

// Returns a list of all absolute paths found under the base path that correspond to movies
func GetFilePaths(basepath string) ([]string, error) {
	stat, err := os.Lstat(basepath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("Not a directory")
	}
	abs, err := filepath.Abs(basepath)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	//path will be absolute
	visitor := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && AcceptedFiletypes[GetFileType(path)] {
			out = append(out, path)
		}
		return nil
	}
	filepath.Walk(abs, visitor)
	return out, nil
}

// Returns a list of Movies found under the base path, with their default initialization
func ScanForMovies(basepath string) []Movie {
	paths, _ := GetFilePaths(basepath)
	newlist := make([]Movie, 0)
	for _, path := range paths {
		newlist = append(newlist, *NewMovie(path))
	}
	return newlist
}

func PathExists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}
