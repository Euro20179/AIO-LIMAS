package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aiolimas/search"

	"aiolimas/types"

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

func BuildEntryTree() (map[int64]db_types.EntryTree, error) {
	out := map[int64]db_types.EntryTree{}

	allRows, err := Db.Query(`SELECT * FROM entryInfo`)
	if err != nil {
		return out, err
	}

	for allRows.Next() {
		var cur db_types.EntryTree

		err := cur.EntryInfo.ReadEntry(allRows)
		if err != nil {
			println(err.Error())
			continue
		}
		cur.UserInfo, err = GetUserViewEntryById(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		cur.MetaInfo, err = GetMetadataEntryById(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		children, err := GetChildren(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		for _, child := range children {
			cur.Children = append(cur.Children, fmt.Sprintf("%d", child.ItemId))
		}

		copies, err := GetCopiesOf(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		for _, c := range copies {
			cur.Copies = append(cur.Copies, fmt.Sprintf("%d", c.ItemId))
		}

		out[cur.EntryInfo.ItemId] = cur
	}
	//
	// for id, cur := range out {
	// 	children, err := GetChildren(id)
	// 	if err != nil{
	// 		println(err.Error())
	// 		continue
	// 	}
	// 	for _, child := range children {
	// 		cur.Children = append(cur.Children, child.ItemId)
	// 	}
	// }

	return out, nil
}

func Begin(entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent("Viewing", entry.ItemId)
	if err != nil {
		return err
	}

	if entry.Status != db_types.S_FINISHED {
		entry.Status = db_types.S_VIEWING
	} else {
		entry.Status = db_types.S_REVIEWING
	}

	return nil
}

func Finish(entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent("Finished", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_FINISHED
	entry.ViewCount += 1

	return nil
}

func Plan(entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent("Planned", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_PLANNED

	return nil
}

func Resume(entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent("ReViewing", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_REVIEWING
	return nil
}

func Drop(entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent("Dropped", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_DROPPED

	return nil
}

func Pause(entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent("Paused", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_PAUSED

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
			 artStyle INTEGER,
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

	Db = conn
}

func getById[T db_types.TableRepresentation](id int64, tblName string, out *T) error {
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

func GetInfoEntryById(id int64) (db_types.InfoEntry, error) {
	var res db_types.InfoEntry
	return res, getById(id, "entryInfo", &res)
}

func GetUserViewEntryById(id int64) (db_types.UserViewingEntry, error) {
	var res db_types.UserViewingEntry
	return res, getById(id, "userViewingInfo", &res)
}

func GetUserEventEntryById(id int64) (db_types.UserViewingEvent, error) {
	var res db_types.UserViewingEvent
	return res, getById(id, "userEventInfo", &res)
}

func GetMetadataEntryById(id int64) (db_types.MetadataEntry, error) {
	var res db_types.MetadataEntry
	return res, getById(id, "metadata", &res)
}

// **WILL ASSIGN THE ENTRYINFO.ID**
func AddEntry(entryInfo *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry, userViewingEntry *db_types.UserViewingEntry) error {
	id := rand.Int64()

	entryInfo.ItemId = id
	metadataEntry.ItemId = id
	userViewingEntry.ItemId = id

	entries := map[string]db_types.TableRepresentation{
		"entryInfo":       *entryInfo,
		"metadata":        *metadataEntry,
		"userViewingInfo": *userViewingEntry,
	}

	for entryName, entry := range entries {
		entryData := db_types.StructNamesToDict(entry)

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

	if userViewingEntry.Status != db_types.Status("") {
		err := RegisterUserEvent(db_types.UserViewingEvent{
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
		var info db_types.InfoEntry
		var userEntry db_types.UserViewingEntry
		var metadata db_types.MetadataEntry
		name := entry.Name()

		fullPath := filepath.Join(path, entry.Name())
		info.En_Title = name
		info.ParentId = parent
		info.Format = db_types.F_DIGITAL
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

func RegisterUserEvent(event db_types.UserViewingEvent) error {
	_, err := Db.Exec(`
		INSERT INTO userEventInfo (itemId, timestamp, event, after)
		VALUES (?, ?, ?, ?)
	`, event.ItemId, event.Timestamp, event.Event, event.After)
	return err
}

func RegisterBasicUserEvent(event string, itemId int64) error {
	var e db_types.UserViewingEvent
	e.Event = event
	e.Timestamp = uint64(time.Now().UnixMilli())
	e.ItemId = itemId
	return RegisterUserEvent(e)
}

func UpdateUserViewingEntry(entry *db_types.UserViewingEntry) error {
	return updateTable(*entry, "userViewingInfo")
}

func MoveUserViewingEntry(oldEntry *db_types.UserViewingEntry, newId int64) error {
	oldEntry.ItemId = newId
	return UpdateUserViewingEntry(oldEntry)
}

func MoveUserEventEntries(eventList []db_types.UserViewingEvent, newId int64) error {
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

func updateTable(tblRepr db_types.TableRepresentation, tblName string) error {
	updateStr := `UPDATE ` + tblName + ` SET `

	data := db_types.StructNamesToDict(tblRepr)

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

func UpdateMetadataEntry(entry *db_types.MetadataEntry) error {
	return updateTable(*entry, "metadata")
}

func UpdateInfoEntry(entry *db_types.InfoEntry) error {
	return updateTable(*entry, "entryInfo")
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
	DATA_OR      DataChecker = iota
	DATA_AND     DataChecker = iota
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

type LogicType int

const (
	LOGIC_AND LogicType = iota
	LOGIC_OR  LogicType = iota
)

type SearchData struct {
	DataName  string
	DataValue []string
	Checker   DataChecker
	LogicType LogicType
}

type SearchQuery []SearchData

func colValToCorrectType(name string, value string) (any, error) {

	u := func(val string) (uint64, error) {
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return 0, err
		}
		return n, nil
	}

	f := func(val string) (float64, error) {
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return n, nil
	}

	switch name {
	case "artStyle":
		return u(value)
	case "parentId":
		return u(value)
	case "itemId":
		return u(value)
	case "copyOf":
		return u(value)
	case "viewCount":
		return u(value)
	case "format":
		return u(value)
	case "purchasePrice":
		return f(value)
	case "generalRating":
		return f(value)
	case "userRating":
		return f(value)
	}

	// if the user types a numeric value, assume they meant it to be of type float
	converted, err := f(value)
	if err == nil {
		return converted, nil
	}

	return value, nil
}

func searchData2Query(query *sqlbuilder.SelectBuilder, previousExpr string, searchData SearchData) string {
	name := searchData.DataName
	origValue := searchData.DataValue
	logicType := searchData.LogicType

	if name == "" {
		panic("Name cannot be empty when turning searchData into query")
	}

	var coercedValues []any
	for _, val := range origValue {
		newVal, err := colValToCorrectType(name, val)
		if err != nil {
			println(err.Error())
			continue
		}
		coercedValues = append(coercedValues, newVal)
	}

	logicFN := query.And
	if logicType == LOGIC_OR {
		logicFN = query.Or
	}

	exprFn := query.EQ

	switch searchData.Checker {
	case DATA_GT:
		exprFn = query.GT
	case DATA_GE:
		exprFn = query.GE
	case DATA_LT:
		exprFn = query.LT
	case DATA_LE:
		exprFn = query.LE
	case DATA_EQ:
		exprFn = query.EQ
	case DATA_NE:
		exprFn = query.NE
	case DATA_IN:
		flattenedValue := []interface{}{
			coercedValues,
		}
		exprFn = func(field string, value interface{}) string {
			return query.In(name, sqlbuilder.Flatten(flattenedValue)...)
		}
	case DATA_NOTIN:
		flattenedValue := []interface{}{
			coercedValues,
		}
		exprFn = func(field string, value interface{}) string {
			return query.NotIn(name, sqlbuilder.Flatten(flattenedValue)...)
		}
	case DATA_LIKE:
		exprFn = query.Like
	case DATA_NOTLIKE:
		exprFn = query.NotLike
	}

	newPrevious := exprFn(name, coercedValues[0])

	var newExpr string
	if previousExpr == "" {
		newExpr = newPrevious
	} else {
		newExpr = logicFN(previousExpr, newPrevious)
	}

	return newExpr
}

func Search3(searchQuery string) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry

	query := "SELECT entryInfo.* FROM entryInfo JOIN userViewingInfo ON entryInfo.itemId == userViewingInfo.itemId JOIN metadata ON entryInfo.itemId == metadata.itemId WHERE %s"

	safeQuery, err := search.Search2String(searchQuery)
	if err != nil {
		return out, err
	}
	fmt.Fprintf(os.Stderr, "Got query %s\n", safeQuery)

	rows, err := Db.Query(fmt.Sprintf(query, safeQuery))
	if err != nil {
		return out, err
	}

	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var row db_types.InfoEntry
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

func GetCopiesOf(id int64) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
	rows, err := Db.Query("SELECT * FROM entryInfo WHERE copyOf = ?", id)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

func mkRows(rows *sql.Rows) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
	defer rows.Close()
	for rows.Next() {
		var entry db_types.InfoEntry
		err := entry.ReadEntry(rows)
		if err != nil {
			return out, err
		}
		out = append(out, entry)
	}
	return out, nil
}

func GetChildren(id int64) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
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

func GetEvents(id int64) ([]db_types.UserViewingEvent, error) {
	var out []db_types.UserViewingEvent
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
		var event db_types.UserViewingEvent
		err := event.ReadEntry(events)
		if err != nil {
			println(err.Error())
			continue
		}
		out = append(out, event)
	}
	return out, nil
}

func getDescendants(id int64, recurse uint64, maxRecurse uint64) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
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

func GetDescendants(id int64) ([]db_types.InfoEntry, error) {
	return getDescendants(id, 0, 10)
}
