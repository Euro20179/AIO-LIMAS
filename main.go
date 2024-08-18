package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/mattn/go-sqlite3"
)

var db *sql.DB

const (
	F_VHS       = iota // 0
	F_CD        = iota // 1
	F_DVD       = iota // 2
	F_BLURAY    = iota // 3
	F_4KBLURAY  = iota // 4
	F_MANGA     = iota // 5
	F_BOOK      = iota // 6
	F_DIGITAL   = iota // 7
	F_VIDEOGAME = iota // 8
	F_BOARDGAME = iota // 9
)

func initDb(dbPath string) *sql.DB {
	conn, err := sql.Open("sqlite3", dbPath)
	sqlite3.Version()
	if err != nil {
		panic(err)
	}
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS entryInfo (
			 itemId INTEGER,
			 title TEXT,
			 format INTEGER,
			 location TEXT,
			 purchasePrice NUMERIC,
			 collection TEXT
		)`)
	if err != nil {
		panic("Failed to create general info table\n" + err.Error())
	}
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS metadata (
			itemId INTEGER,
			rating NUMERIC,
			description TEXT,
			length NUEMERIC,
			releaseYear INTEGER
		)
`)
	if err != nil{
		panic("Failed to create metadata table\n" + err.Error())
	}

	//startDate and endDate are expected to number[] stringified into json
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS userViewingInfo (
			itemId INTEGER,
			status TEXT,
			viewCount INTEGER,
			startDate TEXT,
			endDate TEXT,
			userRating NUMERIC
		)
	`)

	if err != nil{
		panic("Failed to create user status/mal/letterboxd table\n" + err.Error())
	}
	return conn
}

// lets the user add an item in their library
func addEntry(w http.ResponseWriter, req *http.Request) {
	title := req.URL.Query().Get("title")
	if title == "" {
		w.WriteHeader(400)
		w.Write([]byte("No title provided\n"))
		return
	}
	id := rand.Uint64()

	query := fmt.Sprintf("INSERT INTO entryInfo (itemId, title, format, location, purchasePrice, collection) VALUES (%d, '%s', 'digital', 'test', 0, 'test')", id, title)
	_, err := db.Exec(query)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into entryInfo table\n" + err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}

//should be an all in one query thingy, that should be able to query based on any matching column in any table
func queryEntries(w http.ResponseWriter, req *http.Request) {

}

// scans a folder and adds all items to the library as best it can
// it will not add it to any collections
func scanFolder() {}

func main() {
	dbPathPtr := flag.String("db-path", "./all.db", "Path to the database file")
	flag.Parse()
	println(*dbPathPtr)
	db = initDb(*dbPathPtr)

	v1 := "/api/v1"
	http.HandleFunc(v1+"/add-entry", addEntry)
	http.ListenAndServe(":8080", nil)
}
