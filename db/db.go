package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/huandu/go-sqlbuilder"
	"github.com/mattn/go-sqlite3"
)

var Db *sql.DB

func dbHasCol(db *sql.DB, colName string) bool {
	info, _ := db.Query("PRAGMA table_info('entryInfo')")
	defer info.Close()
	for info.Next() {
		var id, x, y int
		var name string
		var ty string
		var z string
		info.Scan(&id, &name, &ty, &x, &z, &y)
		if name == colName {
			return true
		}
	}
	return false
}

func ensureAnimeCol(db *sql.DB) {
	if !dbHasCol(db, "isAnime") {
		_, err := db.Exec("ALTER TABLE entryInfo ADD COLUMN isAnime INTEGER")
		if err != nil {
			panic("Could not add isAnime col\n" + err.Error())
		}
	}

	animeShows, _ := db.Query("SELECT * FROM entryInfo WHERE type == 'Anime'")
	var idsToUpdate []int64
	for animeShows.Next() {
		var out InfoEntry
		out.ReadEntry(animeShows)
		idsToUpdate = append(idsToUpdate, out.ItemId)
	}
	animeShows.Close()
	for _, id := range idsToUpdate {
		fmt.Printf("Updating: %d\n", id)
		_, err := db.Exec(`
			UPDATE entryInfo
			SET
				type = 'Show',
				isAnime = ?
			WHERE
				itemId = ?
			`, 1, id)
		if err != nil {
			panic(fmt.Sprintf("Could not update table entry %d to be isAnime", id) + "\n" + err.Error())
		}

	}
}

func ensureCopyOfCol(db *sql.DB) {
	if !dbHasCol(db, "copyOf") {
		_, err := db.Exec("ALTER TABLE entryInfo ADD COLUMN copyOf INTEGER")
		if err != nil {
			panic("Could not add isAnime col\n" + err.Error())
		}
		_, err = db.Exec("UPDATE entryInfo SET copyOf = 0")
		if err != nil {
			panic("Could not set copyIds to 0")
		}
	}
}

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
			 isAnime INTEGER
			copyOf INTEGER
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
	ensureAnimeCol(conn)
	ensureCopyOfCol(conn)
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

	hasEntry := rows.Next()
	if !hasEntry {
		return res, fmt.Errorf("Could not find id %d", id)
	}
	err = res.ReadEntry(rows)
	if err != nil {
		return res, err
	}
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
	err = res.ReadEntry(rows)
	if err != nil {
		return res, err
	}
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
	err = res.ReadEntry(rows)
	if err != nil {
		return res, err
	}
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
			type,
			isAnime,
			copyOf
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := Db.Exec(entryQuery, id,
		entryInfo.En_Title,
		entryInfo.Native_Title,
		entryInfo.Format,
		entryInfo.Location,
		entryInfo.PurchasePrice,
		entryInfo.Collection,
		entryInfo.Parent,
		entryInfo.Type,
		entryInfo.IsAnime,
		entryInfo.CopyOf)
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
	_, err := Db.Exec(`
		UPDATE metadata
		SET
			rating = ?,
			description = ?,
			releaseYear = ?,
			thumbnail = ?,
			mediaDependant = ?,
			dataPoints = ?
		WHERE
			itemId = ?
	`, entry.Rating, entry.Description,
		entry.ReleaseYear, entry.Thumbnail, entry.MediaDependant,
		entry.Datapoints, entry.ItemId)
	if err != nil {
		return err
	}

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
			location = ?,
			purchasePrice = ?,
			collection = ?,
			parentId = ?,
			type = ?,
			isAnime = ?,
			copyOf = ?
		WHERE
			itemId = ?
	`, entry.En_Title, entry.Native_Title, entry.Format,
		entry.Location, entry.PurchasePrice, entry.Collection,
		entry.Parent, entry.Type, entry.IsAnime, entry.CopyOf, entry.ItemId)
	return err
}

type EntryInfoSearch struct {
	TitleSearch       string
	NativeTitleSearch string
	Format            []Format
	LocationSearch    string
	PurchasePriceGt   float64
	PurchasePriceLt   float64
	InTags            []string
	HasParent         []int64
	Type              []MediaTypes
	IsAnime           int
	CopyIds           []int64
	UserStatus        Status
}

func buildQString[T any](withList []T) string {
	var out string
	for i := range withList {
		if i != len(withList)-1 {
			out += "?, "
		} else {
			out += "?"
		}
	}
	return out
}

func Delete(id int64) error {
	transact, err := Db.Begin()
	if err != nil {
		return err
	}
	transact.Exec(`DELETE FROM entryInfo WHERE itemId = ?`, id)
	transact.Exec(`DELETE FROM metadata WHERE itemId = ?`, id)
	transact.Exec(`DELETE FROM userViewingInfo WHERE itemId = ?`, id)
	return transact.Commit()
}

func Search(mainSearchInfo EntryInfoSearch) ([]InfoEntry, error) {
	query := sqlbuilder.NewSelectBuilder()
	query.Select("entryInfo.*").From("entryInfo").Join("userViewingInfo", "entryInfo.itemId == userViewingInfo.itemId")
	var queries []string
	// query := `SELECT * FROM entryInfo WHERE true`

	if len(mainSearchInfo.Format) > 0 {
		formats := []interface{}{
			mainSearchInfo.Format,
		}
		queries = append(queries, query.In("format", sqlbuilder.Flatten(formats)...))
	}

	if mainSearchInfo.LocationSearch != "" {
		queries = append(queries, query.Like("location", mainSearchInfo.LocationSearch))
	}
	if mainSearchInfo.TitleSearch != "" {
		queries = append(queries, query.Like("en_title", mainSearchInfo.TitleSearch))
	}
	if mainSearchInfo.NativeTitleSearch != "" {
		queries = append(queries, query.Like("native_title", mainSearchInfo.NativeTitleSearch))
	}
	if mainSearchInfo.PurchasePriceGt != 0 {
		queries = append(queries, query.GT("purchasePrice", mainSearchInfo.PurchasePriceGt))
	}
	if mainSearchInfo.PurchasePriceLt != 0 {
		queries = append(queries, query.LT("purchasePrice", mainSearchInfo.PurchasePriceLt))
	}
	if len(mainSearchInfo.InTags) > 0 {
		cols := []interface{}{
			mainSearchInfo.InTags,
		}
		queries = append(queries, query.In("collection", sqlbuilder.Flatten(cols)...))
	}
	if len(mainSearchInfo.HasParent) > 0 {
		pars := []interface{}{
			mainSearchInfo.HasParent,
		}
		queries = append(queries, query.In("parentId", sqlbuilder.Flatten(pars)...))
	}
	if len(mainSearchInfo.CopyIds) > 0 {
		cos := []interface{}{
			mainSearchInfo.CopyIds,
		}
		queries = append(queries, query.In("copyOf", sqlbuilder.Flatten(cos)...))
	}
	if len(mainSearchInfo.Type) > 0 {
		tys := []interface{}{
			mainSearchInfo.Type,
		}
		queries = append(queries, query.In("type", sqlbuilder.Flatten(tys)...))
	}

	if mainSearchInfo.IsAnime != 0 {
		queries = append(queries, query.Equal("isAnime", mainSearchInfo.IsAnime-1))
	}

	if mainSearchInfo.UserStatus != "" {
		queries = append(queries, query.Equal("userViewingInfo.status", mainSearchInfo.UserStatus))
	}

	query = query.Where(queries...)

	finalQuery, args := query.Build()
	rows, err := Db.Query(
		finalQuery,
		args...,
	)

	var out []InfoEntry

	if err != nil {
		return out, err
	}

	defer rows.Close()
	i := 0
	for rows.Next() {
		i += 1
		var row InfoEntry
		err = row.ReadEntry(rows)
		if err != nil {
			println(err.Error())
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func ListCollections() ([]string, error) {
	var out []string
	rows, err := Db.Query(`SELECT en_title FROM entryInfo WHERE type = 'Collection'`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		collection := ""
		err := rows.Scan(&collection)
		if err != nil {
			return out, err
		}
		out = append(out, collection)
	}
	return out, nil
}

func GetCopiesOf(id int64) ([]InfoEntry, error) {
	var out []InfoEntry
	rows, err := Db.Query("SELECT * FROM entryInfo WHERE copyOf = ?", id)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

func mkRows(rows *sql.Rows) ([]InfoEntry, error) {
	var out []InfoEntry
	defer rows.Close()
	for rows.Next() {
		var entry InfoEntry
		err := entry.ReadEntry(rows)
		if err != nil {
			return out, err
		}
		out = append(out, entry)
	}
	return out, nil
}

func GetChildren(id int64) ([]InfoEntry, error) {
	var out []InfoEntry
	rows, err := Db.Query("SELECT * FROM entryInfo where parentId = ?", id)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

func getDescendants(id int64, recurse uint64, maxRecurse uint64) ([]InfoEntry, error) {
	var out []InfoEntry
	if recurse > maxRecurse {
		return out, nil
	}

	children, err := GetChildren(id)
	if err != nil {
		return out, err
	}

	for _, item := range children {
		out = append(out, item)
		newItems, err := getDescendants(item.ItemId, recurse+1, maxRecurse)
		if err != nil {
			continue
		}
		out = append(out, newItems...)
	}
	return out, nil
}

func GetDescendants(id int64) ([]InfoEntry, error) {
	return getDescendants(id, 0, 10)
}
