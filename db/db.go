package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"

	"github.com/mattn/go-sqlite3"
)

var Db *sql.DB

func InitDb(dbPath string) {
	conn, err := sql.Open("sqlite3", dbPath)
	sqlite3.Version()
	if err != nil {
		panic(err)
	}
	// parent is for somethign like a season of a show
	_, err = conn.Exec(`CREATE TABLE IF NOT EXISTS entryInfo (
			 itemId INTEGER,
			 title TEXT,
			 format INTEGER,
			 location TEXT,
			 purchasePrice NUMERIC,
			 collection TEXT,
			 parentId INTEGER
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
			releaseYear INTEGER,
			thumbnail TEXT,
			dataPoints TEXT
		)
`)
	if err != nil {
		panic("Failed to create metadata table\n" + err.Error())
	}

	// startDate and endDate are expected to number[] stringified into json
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
	if err != nil {
		panic("Failed to create user status/mal/letterboxd table\n" + err.Error())
	}
	Db = conn
}

func GetInfoEntryById(id int64) (InfoEntry, error) {
	var res InfoEntry
	query := fmt.Sprintf("SELECT * FROM entryInfo WHERE itemId == %d;", id)
	rows, err := Db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&res.ItemId, &res.Title, &res.Format, &res.Location, &res.PurchasePrice, &res.Collection, &res.Parent)
	return res, nil
}

func GetUserViewEntryById(id int64) (UserViewingEntry, error) {
	var res UserViewingEntry
	query := fmt.Sprintf("SELECT * FROM userViewingInfo WHERE itemId == %d;", id)
	rows, err := Db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&res.ItemId, &res.Status, &res.ViewCount, &res.StartDate, &res.EndDate, &res.UserRating)
	return res, nil
}

func GetMetadataEntryById(id int64) (MetadataEntry, error) {
	var res MetadataEntry
	query := fmt.Sprintf("SELECT * FROM metadata WHERE itemId == %d;", id)
	rows, err := Db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&res.ItemId, &res.Rating, &res.Description, &res.Length, &res.ReleaseYear)
	return res, nil
}

// **WILL ASSIGN THE ENTRYINFO.ID**
func AddEntry(entryInfo *InfoEntry, metadataEntry *MetadataEntry, userViewingEntry *UserViewingEntry) error {
	id := rand.Int64()

	entryInfo.ItemId = id
	metadataEntry.ItemId = id
	userViewingEntry.ItemId = id

	entryQuery := fmt.Sprintf(
		`INSERT INTO entryInfo (
			itemId, title, format, location, purchasePrice, collection, parentId
		) VALUES (%d, '%s', %d, '%s', %f, '%s', %d)`,
		id,
		entryInfo.Title,
		entryInfo.Format,
		entryInfo.Location,
		entryInfo.PurchasePrice,
		entryInfo.Collection,
		entryInfo.Parent,
	)
	_, err := Db.Exec(entryQuery)
	if err != nil {
		return err
	}

	metadataQuery := fmt.Sprintf(`INSERT INTO metadata (
			itemId,
			rating,
			description,
			length,
			releaseYear,
			thumbnail,
			dataPoints
		) VALUES (%d, %f, '%s', %d, %d, '%s', '%s')`,
		metadataEntry.ItemId,
		metadataEntry.Rating,
		metadataEntry.Description,
		metadataEntry.Length,
		metadataEntry.ReleaseYear,
		metadataEntry.Thumbnail,
		metadataEntry.Datapoints,
	)
	_, err = Db.Exec(metadataQuery)
	if err != nil {
		return err
	}

	if userViewingEntry.StartDate == "" {
		userViewingEntry.StartDate = "[]"
	} 
	if userViewingEntry.EndDate == "" {
		userViewingEntry.EndDate = "[]"
	}

	userViewingQuery := fmt.Sprintf(`INSERT INTO userViewingInfo (
			itemId,
			status,
			viewCount,
			startDate,
			endDate,
			userRating
		) VALUES (%d, '%s', %d, '%s', '%s', %f)`,
		userViewingEntry.ItemId,
		userViewingEntry.Status,
		userViewingEntry.ViewCount,
		userViewingEntry.StartDate,
		userViewingEntry.EndDate,
		userViewingEntry.UserRating,
	)
	_, err = Db.Exec(userViewingQuery)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserViewingEntry(entry *UserViewingEntry) error{
	Db.Exec(fmt.Sprintf(`
		UPDATE userViewingInfo
		SET
			status = '%s',
			viewCount = %d,
			startDate = '%s',
			endDate = '%s',
			userRating = %f
		WHERE
			itemId = %d
	`, entry.Status, entry.ViewCount, entry.StartDate, entry.EndDate, entry.UserRating, entry.ItemId))

	return nil
}
