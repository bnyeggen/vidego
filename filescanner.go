package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Utilities for scanning the filesystem and turning movie files into Movie objects
// TODO: Consider https://github.com/howeyc/fsnotify

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
	//Unlikely to be a collision over the first N bytes
	//This doesn't depend on files being < this threshold
	io.CopyN(hasher, fd, 32000000)
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

func scanAllPaths() {
	for _, path := range scannerPaths {
		for _, m := range ScanForMovies(path) {
			mByPath := GetMovieByPath(m.Path)
			if mByPath != nil && m.Byte_length == mByPath.Byte_length {
				//Existing name & same size - match
				log.Println("Skipping: " + m.Path)
				continue
			} else {
				//With a "real" db, this could be all goroutines w/ a waitgroup
				//But we're IO bound w/ our disk setup, and finding optimal IO
				//parallelism is a bit painful
				log.Println("Hashing & Storing: " + m.Path)
				m.populateHash()

				mByHash := GetMovieByHashAndSize(m.Hash, m.Byte_length)

				if mByHash == nil && mByPath == nil {
					//New file, with no pre-existing file to deal with
					e := StoreNew(&m)
					if e != nil {
						log.Println(e)
					}
				} else if mByHash == nil && mByPath != nil {
					//Conceptually a new file - the previous occupant can be blanked and collected below
					ClearPathByID(mByPath.Id)
					e := StoreNew(&m)
					if e != nil {
						log.Println(e)
					}
				} else if mByHash != nil && mByPath == nil {
					//Copied or moved from somewhere else to a new location - ignore the "somewhere else"
					e := StoreCopy(mByHash, m.Path)
					if e != nil {
						log.Println(e)
					}

				} else if mByHash != nil && mByPath != nil {
					//Copied or moved from somewhere else, overwriting existing location
					ClearPathByID(mByPath.Id)
					e := StoreCopy(mByHash, m.Path)
					if e != nil {
						log.Println(e)
					}
				}
			}
		}
	}

	//Finally, remove any records of files that no longer exist
	e := RemoveWithInvalidPaths()
	if e != nil {
		log.Println(e)
	}
}
