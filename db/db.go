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
			 en_title TEXT,
			 native_title TEXT,
			 format INTEGER,
			 location TEXT,
			 purchasePrice NUMERIC,
			 collection TEXT,
			 type TEXT,
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
			releaseYear INTEGER,
			thumbnail TEXT,
			mediaDependant TEXT,
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
			userRating NUMERIC,
			notes TEXT
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
	rows.Scan(&res.ItemId, &res.En_Title, &res.Native_Title, &res.Format, &res.Location, &res.PurchasePrice, &res.Collection, &res.Parent)
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
	rows.Scan(&res.ItemId, &res.Status, &res.ViewCount, &res.StartDate, &res.EndDate, &res.UserRating, &res.Notes)
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
	rows.Scan(&res.ItemId, &res.Rating, &res.Description, &res.ReleaseYear, &res.Thumbnail, &res.MediaDependant, &res.Datapoints)
	return res, nil
}

// **WILL ASSIGN THE ENTRYINFO.ID**
func AddEntry(entryInfo *InfoEntry, metadataEntry *MetadataEntry, userViewingEntry *UserViewingEntry) error {
	id := rand.Int64()

	entryInfo.ItemId = id
	metadataEntry.ItemId = id
	userViewingEntry.ItemId = id

	entryQuery := `INSERT INTO entryInfo (
			itemId,
			en_title,
			native_title,
			format,
			location,
			purchasePrice,
			collection,
			parentId,
			type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := Db.Exec(entryQuery, id,
		entryInfo.En_Title,
		entryInfo.Native_Title,
		entryInfo.Format,
		entryInfo.Location,
		entryInfo.PurchasePrice,
		entryInfo.Collection,
		entryInfo.Parent,
		entryInfo.Type)
	if err != nil {
		return err
	}

	metadataQuery := `INSERT INTO metadata (
			itemId,
			rating,
			description,
			mediaDependant,
			releaseYear,
			thumbnail,
			dataPoints
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = Db.Exec(metadataQuery, metadataEntry.ItemId,
		metadataEntry.Rating,
		metadataEntry.Description,
		metadataEntry.MediaDependant,
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
			userRating,
			notes
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = Db.Exec(userViewingQuery,
		userViewingEntry.ItemId,
		userViewingEntry.Status,
		userViewingEntry.ViewCount,
		userViewingEntry.StartDate,
		userViewingEntry.EndDate,
		userViewingEntry.UserRating,
		userViewingEntry.Notes,
	)
	if err != nil {
		return err
	}

	return nil
}

func ScanFolderWithParent(path string, collection string, parent int64) []error {
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
		info.En_Title = name
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
			userRating = ?,
			notes = ?
		WHERE
			itemId = ?
	`, entry.Status, entry.ViewCount, entry.StartDate, entry.EndDate, entry.UserRating, entry.Notes, entry.ItemId)

	return nil
}

func UpdateMetadataEntry(entry *MetadataEntry) error {
	Db.Exec(`
		UPDATE metadata
		SET
			rating = ?,
			description = ?,
			releaseYear = ?,
			thumbnail = ?,
			mediaDependant = ?,
			dataPoints = ?,
		WHERE
			itemId = ?
	`, entry.Rating, entry.Description,
		entry.ReleaseYear, entry.Thumbnail, entry.MediaDependant,
		entry.Datapoints, entry.ItemId)

	return nil
}

func UpdateInfoEntry(entry *InfoEntry) error {
	/*
		itemId INTEGER,
		en_title TEXT,
		native_title TEXT,
		format INTEGER,
		location TEXT,
		purchasePrice NUMERIC,
		collection TEXT,
		parentId INTEGER
	*/
	_, err := Db.Exec(`
		UPDATE entryInfo
		SET
			en_title = ?,
			native_title = ?,
			format = ?,
			locaiton = ?
			purchasePrice = ?,
			collection = ?,
			parentId = ?,
			type = ?
		WHERE
			itemId = ?
	`, entry.En_Title, entry.Native_Title, entry.Format,
		entry.Location, entry.PurchasePrice, entry.Collection,
		entry.Parent, entry.Type, entry.ItemId)
	return err
}
