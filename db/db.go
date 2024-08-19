package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

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
			volumes INTEGER,
			chapters INTEGER,
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
	query := "SELECT * FROM entryInfo WHERE itemId == ?;"
	rows, err := Db.Query(query, id)
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
	query := "SELECT * FROM userViewingInfo WHERE itemId == ?;"
	rows, err := Db.Query(query, id)
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
	query := "SELECT * FROM metadata WHERE itemId == ?;"
	rows, err := Db.Query(query, id)
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

	entryQuery := `INSERT INTO entryInfo (
			itemId, title, format, location, purchasePrice, collection, parentId
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := Db.Exec(entryQuery, id,
		entryInfo.Title,
		entryInfo.Format,
		entryInfo.Location,
		entryInfo.PurchasePrice,
		entryInfo.Collection,
		entryInfo.Parent)
	if err != nil {
		return err
	}

	metadataQuery := `INSERT INTO metadata (
			itemId,
			rating,
			description,
			length,
			volumes,
			chapters,
			releaseYear,
			thumbnail,
			dataPoints
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = Db.Exec(metadataQuery, metadataEntry.ItemId,
		metadataEntry.Rating,
		metadataEntry.Description,
		metadataEntry.Length,
		metadataEntry.Volumes,
		metadataEntry.Chapters,
		metadataEntry.ReleaseYear,
		metadataEntry.Thumbnail,
		metadataEntry.Datapoints)
	if err != nil {
		return err
	}

	if userViewingEntry.StartDate == "" {
		userViewingEntry.StartDate = "[]"
	}
	if userViewingEntry.EndDate == "" {
		userViewingEntry.EndDate = "[]"
	}

	userViewingQuery := `INSERT INTO userViewingInfo (
			itemId,
			status,
			viewCount,
			startDate,
			endDate,
			userRating
		) VALUES (?, ?, ?, ?, ?, ?)`

	_, err = Db.Exec(userViewingQuery,
		userViewingEntry.ItemId,
		userViewingEntry.Status,
		userViewingEntry.ViewCount,
		userViewingEntry.StartDate,
		userViewingEntry.EndDate,
		userViewingEntry.UserRating,
	)
	if err != nil {
		return err
	}

	return nil
}

func ScanFolderWithParent(path string, collection string, parent int64) []error{
	stat, err := os.Stat(path)
	if err != nil {
		return []error{err}
	}
	if !stat.IsDir() {
		return []error{fmt.Errorf("%s is not a directory\n", path)}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return []error{err}
	}

	var errors []error
	for _, entry := range entries {
		var info InfoEntry
		var userEntry UserViewingEntry
		var metadata MetadataEntry
		name := entry.Name()

		fullPath := filepath.Join(path, entry.Name())
		info.Title = name
		info.Parent = parent
		info.Format = F_DIGITAL
		info.Location = fullPath
		info.Collection = collection

		err := AddEntry(&info, &metadata, &userEntry)
		if err != nil {
			errors = append(errors, err)
		}

		if entry.IsDir() {
			newErrors := ScanFolderWithParent(fullPath, collection, info.ItemId)
			errors = append(errors, newErrors...)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func ScanFolder(path string, collection string) []error {
	return ScanFolderWithParent(path, collection, 0)
}

func UpdateUserViewingEntry(entry *UserViewingEntry) error {
	Db.Exec(`
		UPDATE userViewingInfo
		SET
			status = ?,
			viewCount = ?,
			startDate = ?,
			endDate = ?,
			userRating = ?
		WHERE
			itemId = ?
	`, entry.Status, entry.ViewCount, entry.StartDate, entry.EndDate, entry.UserRating, entry.ItemId)

	return nil
}

func UpdateMetadataEntry(entry *MetadataEntry) error {
	Db.Exec(`
		UPDATE metadata
		SET
			rating = ?,
			description = ?,
			length = ?,
			volumes = ?,
			chapters = ?,
			releaseYear = ?,
			thumbnail = ?,
			dataPoints = ?,
		WHERE
			itemId = ?
	`, entry.Rating, entry.Description, entry.Length, entry.Volumes, entry.Chapters,
		entry.ReleaseYear, entry.Thumbnail, entry.Datapoints, entry.ItemId)

	return nil
}
