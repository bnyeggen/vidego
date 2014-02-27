package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
)

//Gloabl state holding the open DB
var mainDB *sql.DB

// Initialize database if it hasn't been already, returning a pointer to it
func migrate(loc string) *sql.DB {
	myDB, err := sql.Open("sqlite3", loc)
	if err != nil {
		panic(err)
	}
	sql := `
	create table if not exists
	movies(id integer primary key,
	       path text not null unique,
	       byte_length integer not null,
	       title text not null, 
	       director text,
	       year integer,
	       added_date text not null,
	       watched integer not null,
	       hash text);
	create index if not exists movies_hash on movies(hash);
	`
	_, err = myDB.Exec(sql)
	if err != nil {
		panic(err)
	}
	return myDB
}

//Store without check
func Store(m *Movie) error {
	dateAsTxt, _ := m.Added_date.MarshalText()
	res, err := mainDB.Exec("insert into movies(path, byte_length, title, director, year, added_date, watched, hash) values(?,?,?,?,?,?,?,?)", m.Path, m.Byte_length, m.Title, m.Director, m.Year, dateAsTxt, m.Watched, m.Hash)
	if err != nil {
		return err
	}
	m.Id, err = res.LastInsertId()
	return err
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
	r, e := mainDB.Query("select path from movies")
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

//Specialized version of the general Modify fn
func UpdatePath(oldPath string, newPath string) error {
	_, err := mainDB.Exec("update movies set path=? where path=?", newPath, oldPath)
	return err
}

// Stores as a new record, or detects moves and updates if necessary
// Returns any errors
func CheckAndStore(m *Movie) error {
	//Find by path
	oldM := GetMovieByPath(m.Path)
	if oldM == nil {
		//No existing movie at that path, so either this movie was moved here, or it's new
		//Hash the file if it's empty
		if m.Hash == "" {
			m.populateHash()
		}
		//Check based on hash and length for movement
		oldM = GetMovieByHashAndSize(m.Hash, m.Byte_length)
		if oldM == nil {
			return Store(m)
		} else {
			return UpdatePath(oldM.Path, m.Path)
		}
	} else {
		//Movie exists at that path already - verify it's the same one
		if m.Byte_length == oldM.Byte_length {
			return nil
		} else {
			//Old one needs to be deleted, this one to be inserted
			e := Remove(oldM.Path)
			if e != nil {
				return e
			}
			//This will now succeed
			return Store(m)
		}
	}
}
