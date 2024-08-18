package globals;

import (
	"database/sql"

	"github.com/mattn/go-sqlite3"
)

var Db *sql.DB

func InitDb(dbPath string) {
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
	Db = conn
}
