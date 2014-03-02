package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"time"
)

//Gloabl state holding the open DB
var mainDB *sql.DB

// Initialize database if it hasn't been already, returning a pointer to it
func migrate(loc string) *sql.DB {
	myDB, err := sql.Open("sqlite3", loc)
	if err != nil {
		panic(err)
	}
	//Path is nullable so we can retain metadata by hash while we're handling moves
	sql := `
	create table if not exists
	movies(id integer primary key,
	       path text,
	       byte_length integer not null,
	       title text not null, 
	       director text,
	       year integer,
	       added_date text not null,
	       watched integer not null,
	       hash text);
	create index if not exists movies_hash on movies(hash);
	create index if not exists movies_path on movies(path);
	`
	_, err = myDB.Exec(sql)
	if err != nil {
		panic(err)
	}
	return myDB
}

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

//Store without check, auto-generating ID
func StoreNew(m *Movie) error {
	dateAsTxt, _ := m.Added_date.MarshalText()
	res, err := mainDB.Exec("insert into movies(path, byte_length, title, director, year, added_date, watched, hash) values(?,?,?,?,?,?,?,?)", m.Path, m.Byte_length, m.Title, m.Director, m.Year, dateAsTxt, m.Watched, m.Hash)
	if err != nil {
		return err
	}
	m.Id, err = res.LastInsertId()
	return err
}

//Stores as a "new" movie in the given location, with existing metadata and a new ID
func StoreCopy(m *Movie, newPath string) error {
	m.Id = 0
	m.Path = newPath
	return StoreNew(m)
}

//Constrained to be only Id and one value
func ModifyWithMap(m map[string]interface{}) error {
	id, ok := m["id"]
	if !ok {
		return errors.New("No ID in ModifyWithMap")
	}
	if newTitle, ok := m["title"]; ok {
		_, err := mainDB.Exec("update movies set title=? where id=?", newTitle, id)
		return err
	} else if newDirector, ok := m["director"]; ok {
		_, err := mainDB.Exec("update movies set director=? where id=?", newDirector, id)
		return err
	} else if watched, ok := m["watched"]; ok {
		_, err := mainDB.Exec("update movies set watched=? where id=?", watched, id)
		return err
	} else if year, ok := m["year"]; ok {
		_, err := mainDB.Exec("update movies set year=? where id=?", year, id)
		return err
	}
	return errors.New("No valid update to map")
}

// Remove the movie record with the given path.
func Remove(path string) error {
	_, err := mainDB.Exec("delete from movies where path = ?", path)
	return err
}

func DumpDB() ([]Movie, error) {
	//TODO: Reuse the getMovieByWhereClause machinery as much as possible
	out := make([]Movie, 0)
	r, e := mainDB.Query("select id, path, byte_length, title, director, year, added_date, watched, hash from movies")
	if e != nil {
		return out, e
	}
	for r.Next() {
		var m Movie
		var added_date_txt string
		e = r.Scan(&m.Id, &m.Path, &m.Byte_length, &m.Title, &m.Director, &m.Year, &added_date_txt, &m.Watched, &m.Hash)
		m.Added_date.UnmarshalText([]byte(added_date_txt))
		if e != nil {
			return out, e
		}
		out = append(out, m)
	}
	return out, nil
}

func DumpAllPaths() ([]string, error) {
	out := make([]string, 0)
	r, e := mainDB.Query("select path from movies where path != null")
	if e != nil {
		return out, e
	}
	for r.Next() {
		var s string
		e = r.Scan(&s)
		if e != nil {
			return out, e
		}
		out = append(out, s)
	}
	return out, nil
}

func DeleteNullPaths() error {
	_, e := mainDB.Exec("delete from movies where path = null")
	return e
}

//This just ensures we have the same fields & order when we query
//Obviously, don't SQL inject yourself in the clause w/ user input
func getMovieByWhereClause(clause string, args ...interface{}) *Movie {
	var m Movie
	q := "select id, path, byte_length, title, director, year, added_date, watched, hash from movies " + clause
	r := mainDB.QueryRow(q, args...)
	var added_date_txt string
	e := r.Scan(&m.Id, &m.Path, &m.Byte_length, &m.Title, &m.Director, &m.Year, &added_date_txt, &m.Watched, &m.Hash)
	m.Added_date.UnmarshalText([]byte(added_date_txt))
	if e == sql.ErrNoRows {
		return nil
	}
	return &m
}

func GetMovieByPath(path string) *Movie {
	return getMovieByWhereClause("where path=?", path)
}

func GetMovieByHashAndSize(hash string, byte_length int64) *Movie {
	return getMovieByWhereClause("where hash=? and byte_length=?", hash, byte_length)
}

func GetMovieByID(id int64) *Movie {
	return getMovieByWhereClause("where id=?", id)
}

func UpdatePath(oldPath string, newPath string) error {
	_, err := mainDB.Exec("update movies set path=? where path=?", newPath, oldPath)
	return err
}

func ClearPathByID(id int64) error {
	_, err := mainDB.Exec(`update movies set path="" where id=?`, id)
	return err
}
