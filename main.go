package main

import (
	"database/sql"
	"flag"

	"github.com/mattn/go-sqlite3"
)

var db *sql.DB

const (
	F_VHS = iota //0
	F_CD = iota //1
	F_DVD = iota //2
	F_BLURAY = iota //3
	F_4KBLURAY = iota //4
	F_MANGA = iota //5
	F_BOOK = iota //6
	F_DIGITAL = iota //7
)

func initDb(dbPath string) *sql.DB {
	conn, err := sql.Open("sqlite3", dbPath) 
	sqlite3.Version()
	if err != nil {
		panic(err)
	}
	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS entryInfo (rowid INTEGER, title TEXT, format INTEGER, location TEXT, purchasePrice NUMERIC, collection TEXT)")

	if err != nil{
		panic("Failed to create table")
	}
	return conn
}

//lets the user add an item in their library
func addEntry() {}

//scans a folder and adds all items to the library as best it can
//it will not add it to any collections
func scanFolder() {}

func main() {
	dbPathPtr := flag.String("db-path", "./all.db", "Path to the database file")
	flag.Parse()
	db = initDb(*dbPathPtr)
}
