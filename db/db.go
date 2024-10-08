package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func ensureProviderColumns(db *sql.DB) error {
	_, err := db.Exec("ALTER TABLE metadata ADD COLUMN provider default ''")
	if err != nil {
		return err
	}
	_, err = db.Exec("ALTER TABLE metadata ADD COLUMN providerID default ''")
	if err != nil {
		return err
	}
	return nil
}

func ensureRatingMax(db *sql.DB) error {
	_, err := db.Exec("ALTER TABLE metadata ADD COLUMN ratingMax default 0")
	return err
}

func ensureCurrentPosition(db *sql.DB) error {
	_, err := db.Exec("ALTER TABLE userViewingInfo ADD COLUMN currentPosition default ''")
	if err != nil {
		return err
	}
	return nil
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
			 parentId INTEGER,
			 isAnime INTEGER,
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
			dataPoints TEXT,
			native_title TEXT,
			title TEXT,
			ratingMax NUMERIC,
			provider TEXT,
			providerID TEXT
		)
`)
	if err != nil {
		panic("Failed to create metadata table\n" + err.Error())
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS userViewingInfo (
			itemId INTEGER,
			status TEXT,
			viewCount INTEGER,
			userRating NUMERIC,
			notes TEXT,
			currentPosition TEXT
		)
	`)
	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS userEventInfo (
			itemId INTEGER,
			timestamp INTEGER,
			after INTEGER,
			event TEXT
		)
	`)
	if err != nil {
		panic("Failed to create user status/mal/letterboxd table\n" + err.Error())
	}

	err = ensureProviderColumns(conn)
	if err != nil {
		println(err.Error())
	}
	err = ensureCurrentPosition(conn)
	if err != nil {
		println(err.Error())
	}

	err = ensureRatingMax(conn)
	if err != nil {
		println(err.Error())
	}

	Db = conn
}

func getById[T TableRepresentation](id int64, tblName string, out *T) error {
	query := "SELECT * FROM " + tblName + " WHERE itemId = ?;"

	rows, err := Db.Query(query, id)
	if err != nil {
		return err
	}

	defer rows.Close()

	hasEntry := rows.Next()
	if !hasEntry {
		return fmt.Errorf("Could not find id %d", id)
	}

	newEntry, err := (*out).ReadEntryCopy(rows)
	if err != nil {
		return err
	}

	*out = newEntry.(T)

	return nil
}

func GetInfoEntryById(id int64) (InfoEntry, error) {
	var res InfoEntry
	return res, getById(id, "entryInfo", &res)
}

func GetUserViewEntryById(id int64) (UserViewingEntry, error) {
	var res UserViewingEntry
	return res, getById(id, "userViewingInfo", &res)
}

func GetUserEventEntryById(id int64) (UserViewingEvent, error) {
	var res UserViewingEvent
	return res, getById(id, "userEventInfo", &res)
}

func GetMetadataEntryById(id int64) (MetadataEntry, error) {
	var res MetadataEntry
	return res, getById(id, "metadata", &res)
}

// **WILL ASSIGN THE ENTRYINFO.ID**
func AddEntry(entryInfo *InfoEntry, metadataEntry *MetadataEntry, userViewingEntry *UserViewingEntry) error {
	id := rand.Int64()

	entryInfo.ItemId = id
	metadataEntry.ItemId = id
	userViewingEntry.ItemId = id

	entries := map[string]TableRepresentation{
		"entryInfo":       *entryInfo,
		"metadata":        *metadataEntry,
		"userViewingInfo": *userViewingEntry,
	}

	for entryName, entry := range entries {
		entryData := StructNamesToDict(entry)

		var entryArgs []any
		questionMarks := ""
		entryQ := `INSERT INTO ` + entryName + ` (`
		for k, v := range entryData {
			entryQ += k + ","
			entryArgs = append(entryArgs, v)
			questionMarks += "?,"
		}
		// cut off trailing comma
		entryQ = entryQ[:len(entryQ)-1] + ")"

		entryQ += "VALUES(" + questionMarks[:len(questionMarks)-1] + ")"
		_, err := Db.Exec(entryQ, entryArgs...)
		if err != nil {
			return err
		}
	}

	if userViewingEntry.Status != Status("") {
		err := RegisterUserEvent(UserViewingEvent{
			ItemId:    userViewingEntry.ItemId,
			Timestamp: uint64(time.Now().UnixMilli()),
			Event:     string(userViewingEntry.Status),
			After:     0,
		})
		if err != nil {
			return err
		}
	}

	err := RegisterBasicUserEvent("Added", metadataEntry.ItemId)
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
		info.ParentId = parent
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

func RegisterUserEvent(event UserViewingEvent) error {
	_, err := Db.Exec(`
		INSERT INTO userEventInfo (itemId, timestamp, event, after)
		VALUES (?, ?, ?, ?)
	`, event.ItemId, event.Timestamp, event.Event, event.After)
	return err
}

func RegisterBasicUserEvent(event string, itemId int64) error {
	var e UserViewingEvent
	e.Event = event
	e.Timestamp = uint64(time.Now().UnixMilli())
	e.ItemId = itemId
	return RegisterUserEvent(e)
}

func UpdateUserViewingEntry(entry *UserViewingEntry) error {
	return updateTable(*entry, "userViewingInfo")
}

func MoveUserViewingEntry(oldEntry *UserViewingEntry, newId int64) error {
	oldEntry.ItemId = newId
	return UpdateUserViewingEntry(oldEntry)
}

func MoveUserEventEntries(eventList []UserViewingEvent, newId int64) error {
	for _, e := range eventList {
		e.ItemId = newId
		err := RegisterUserEvent(e)
		if err != nil {
			return err
		}
	}
	return nil
}

func ClearUserEventEntries(id int64) error {
	_, err := Db.Exec(`
		DELETE FROM userEventInfo
		WHERE itemId = ?
	`, id)
	if err != nil {
		return err
	}
	return nil
}

func updateTable(tblRepr TableRepresentation, tblName string) error {
	updateStr := `UPDATE ` + tblName + ` SET `

	data := StructNamesToDict(tblRepr)

	var updateArgs []any

	for k, v := range data {
		updateArgs = append(updateArgs, v)

		updateStr += k + "= ?,"
	}

	// needs itemid for checking which item to update
	updateArgs = append(updateArgs, tblRepr.Id())

	// remove final trailing comma
	updateStr = updateStr[:len(updateStr)-1]
	updateStr += "\nWHERE itemId = ?"

	_, err := Db.Exec(updateStr, updateArgs...)
	if err != nil {
		return err
	}

	return nil
}

func UpdateMetadataEntry(entry *MetadataEntry) error {
	return updateTable(*entry, "metadata")
}

func UpdateInfoEntry(entry *InfoEntry) error {
	return updateTable(*entry, "entryInfo")
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
	UserStatus        []Status
	UserRatingGt      float64
	UserRatingLt      float64
	ReleasedGE        int64
	ReleasedLE        int64
}

func Delete(id int64) error {
	transact, err := Db.Begin()
	if err != nil {
		return err
	}
	transact.Exec(`DELETE FROM entryInfo WHERE itemId = ?`, id)
	transact.Exec(`DELETE FROM metadata WHERE itemId = ?`, id)
	transact.Exec(`DELETE FROM userViewingInfo WHERE itemId = ?`, id)
	transact.Exec(`DELETE FROM userEventInfo WHERE itemId = ?`, id)
	return transact.Commit()
}

func Search(mainSearchInfo EntryInfoSearch) ([]InfoEntry, error) {
	query := sqlbuilder.NewSelectBuilder()
	query.Select("entryInfo.*").From("entryInfo").Join("userViewingInfo", "entryInfo.itemId == userViewingInfo.itemId").Join("metadata", "entryInfo.itemId == metadata.itemId")

	var queries []string
	// query := `SELECT * FROM entryInfo WHERE true`

	if len(mainSearchInfo.Format) > 0 {
		formats := []interface{}{
			mainSearchInfo.Format,
		}
		queries = append(queries, query.In("format", sqlbuilder.Flatten(formats)...))
	}

	if mainSearchInfo.ReleasedLE == mainSearchInfo.ReleasedGE && mainSearchInfo.ReleasedGE != 0 {
		queries = append(queries, query.EQ("releaseYear", mainSearchInfo.ReleasedGE))
	} else if mainSearchInfo.ReleasedGE < mainSearchInfo.ReleasedLE {
		queries = append(queries, query.And(
			query.GE("releaseYear", mainSearchInfo.ReleasedGE),
			query.LE("releaseYear", mainSearchInfo.ReleasedLE),
		))
	} else if mainSearchInfo.ReleasedLE != 0 {
		queries = append(queries, query.LE("releaseYear", mainSearchInfo.ReleasedLE))
	} else if mainSearchInfo.ReleasedGE != 0 {
		queries = append(queries, query.GE("releaseYear", mainSearchInfo.ReleasedGE))
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
	if mainSearchInfo.UserRatingGt != 0 {
		queries = append(queries, query.GT("userRating", mainSearchInfo.UserRatingGt))
	}
	if mainSearchInfo.UserRatingLt != 0 {
		queries = append(queries, query.LT("userRating", mainSearchInfo.UserRatingLt))
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

	if len(mainSearchInfo.UserStatus) != 0 {
		items := []interface{}{
			mainSearchInfo.UserStatus,
		}
		queries = append(queries, query.In("userViewingInfo.status", sqlbuilder.Flatten(items)...))
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

type DataChecker int

const (
	DATA_GT      DataChecker = iota
	DATA_LT      DataChecker = iota
	DATA_LE      DataChecker = iota
	DATA_GE      DataChecker = iota
	DATA_EQ      DataChecker = iota
	DATA_NE      DataChecker = iota
	DATA_LIKE    DataChecker = iota
	DATA_IN      DataChecker = iota
	DATA_NOTIN   DataChecker = iota
	DATA_NOTLIKE DataChecker = iota
)

func Str2DataChecker(in string) DataChecker {
	switch strings.ToUpper(in) {
	case "GT":
		return DATA_GT
	case "LT":
		return DATA_LT
	case "LE":
		return DATA_LE
	case "GE":
		return DATA_GE
	case "EQ":
		return DATA_EQ
	case "NE":
		return DATA_NE
	case "LIKE":
		return DATA_LIKE
	case "IN":
		return DATA_IN
	case "NOTIN":
		return DATA_NOTIN
	case "NOTLIKE":
		return DATA_NOTLIKE
	}
	return DATA_EQ
}

type SearchData struct {
	DataName  string
	DataValue any
	Checker   DataChecker
}
type SearchQuery []SearchData

func Search2(searchQuery SearchQuery) ([]InfoEntry, error) {
	query := sqlbuilder.NewSelectBuilder()
	query.Select("entryInfo.*").From("entryInfo").Join("userViewingInfo", "entryInfo.itemId == userViewingInfo.itemId").Join("metadata", "entryInfo.itemId == metadata.itemId")

	var queries []string

	for _, searchData := range searchQuery {
		name := searchData.DataName
		value := searchData.DataValue
		if name == "" {
			continue
		}

		switch searchData.Checker {
		case DATA_GT:
			queries = append(queries, query.GT(name, value))
		case DATA_GE:
			queries = append(queries, query.GE(name, value))
		case DATA_LT:
			queries = append(queries, query.LT(name, value))
		case DATA_LE:
			queries = append(queries, query.LE(name, value))
		case DATA_EQ:
			queries = append(queries, query.EQ(name, value))
		case DATA_NE:
			queries = append(queries, query.NE(name, value))
		case DATA_IN:
			flattenedValue := []interface{}{
				value,
			}
			queries = append(queries, query.In(name, sqlbuilder.Flatten(flattenedValue)...))
		case DATA_NOTIN:
			flattenedValue := []interface{}{
				value,
			}
			queries = append(queries, query.NotIn(name, sqlbuilder.Flatten(flattenedValue)...))
		case DATA_LIKE:
			queries = append(queries, query.Like(name, value))
		case DATA_NOTLIKE:
			queries = append(queries, query.NotLike(name, value))
		}
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

func DeleteEvent(id int64, timestamp int64, after int64) error {
	_, err := Db.Exec(`
		DELETE FROM userEventInfo
		WHERE 
			itemId == ? and timestamp == ? and after == ?
	`, id, timestamp, after)
	return err
}

func GetEvents(id int64) ([]UserViewingEvent, error) {
	var out []UserViewingEvent
	events, err := Db.Query(`
		SELECT * from userEventInfo
		WHERE
			itemId == ?
		ORDER BY
			CASE timestamp
				WHEN 0 THEN
					userEventInfo.after
				ELSE timestamp
			END`, id)
	if err != nil {
		return out, err
	}

	defer events.Close()

	for events.Next() {
		var event UserViewingEvent
		err := event.ReadEntry(events)
		if err != nil {
			println(err.Error())
			continue
		}
		out = append(out, event)
	}
	return out, nil
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
