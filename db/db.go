package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiolimas/search"
	"aiolimas/settings"

	"aiolimas/types"

	"github.com/mattn/go-sqlite3"
)

func UserRoot(uid int64) string {
	aioPath := os.Getenv("AIO_DIR")
	return fmt.Sprintf("%s/users/%d/", aioPath, uid)
}

func OpenUserDb(uid int64) (*sql.DB, error) {
	path := UserRoot(uid)
	return sql.Open("sqlite3", path+"all.db")
}

func BuildEntryTree(uid int64) (map[int64]db_types.EntryTree, error) {
	out := map[int64]db_types.EntryTree{}

	Db, err := OpenUserDb(uid)
	if err != nil {
		return out, err
	}
	defer Db.Close()

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
		cur.UserInfo, err = GetUserViewEntryById(uid, cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		cur.MetaInfo, err = GetMetadataEntryById(uid, cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		children, err := GetChildren(uid, cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		for _, child := range children {
			cur.Children = append(cur.Children, fmt.Sprintf("%d", child.ItemId))
		}

		copies, err := GetCopiesOf(uid, cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		for _, c := range copies {
			cur.Copies = append(cur.Copies, fmt.Sprintf("%d", c.ItemId))
		}

		out[cur.EntryInfo.ItemId] = cur
	}

	return out, nil
}

func Begin(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Viewing", entry.ItemId)
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

func Finish(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Finished", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_FINISHED
	entry.ViewCount += 1

	return nil
}

func Plan(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Planned", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_PLANNED

	return nil
}

func Resume(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "ReViewing", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_REVIEWING
	return nil
}

func Drop(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Dropped", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_DROPPED

	return nil
}

func Pause(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Paused", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_PAUSED

	return nil
}

func InitDb(uid int64) error {
	conn, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	sqlite3.Version()
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
			copyOf INTEGER,
			artStyle INTEGER,
			library INTEGER
		)`)
	if err != nil {
		return err
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
			title TEXT,
			native_title TEXT,
			ratingMax NUMERIC,
			provider TEXT,
			providerID TEXT
		)
`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS userViewingInfo (
			itemId INTEGER,
			status TEXT,
			viewCount INTEGER,
			userRating NUMERIC,
			notes TEXT,
			currentPosition TEXT,
			extra TEXT
		)
	`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS userEventInfo (
			itemId INTEGER,
			timestamp INTEGER,
			after INTEGER,
			event TEXT,
			timezone TEXT
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func getById[T db_types.TableRepresentation](uid int64, id int64, tblName string, out *T) error {
	query := "SELECT * FROM " + tblName + " WHERE itemId = ?;"

	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

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

func GetInfoEntryById(uid int64, id int64) (db_types.InfoEntry, error) {
	var res db_types.InfoEntry
	return res, getById(uid, id, "entryInfo", &res)
}

func GetUserViewEntryById(uid int64, id int64) (db_types.UserViewingEntry, error) {
	var res db_types.UserViewingEntry
	return res, getById(uid, id, "userViewingInfo", &res)
}

func GetUserEventEntryById(uid int64, id int64) (db_types.UserViewingEvent, error) {
	var res db_types.UserViewingEvent
	return res, getById(uid, id, "userEventInfo", &res)
}

func GetMetadataEntryById(uid int64, id int64) (db_types.MetadataEntry, error) {
	var res db_types.MetadataEntry
	return res, getById(uid, id, "metadata", &res)
}

func ensureUserJsonNotEmpty(user *db_types.UserViewingEntry) {
	if user.Extra == "" {
		user.Extra = "{}"
	}
}

func ensureMetadataJsonNotEmpty(metadata *db_types.MetadataEntry) {
	if metadata.MediaDependant == "" {
		metadata.MediaDependant = "{}"
	}
	if metadata.Datapoints == "" {
		metadata.Datapoints = "{}"
	}
}

func ListMetadata(uid int64) ([]db_types.MetadataEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	items, err := Db.Query("SELECT * FROM metadata")
	if err != nil {
		return nil, err
	}

	var out []db_types.MetadataEntry

	for items.Next() {
		var row db_types.MetadataEntry
		err := row.ReadEntry(items)
		if err != nil {
			println(err.Error())
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

// TODO: remove timezone parameter from this function, maybe combine it witih userViewingEntry since that also keeps track of the timezone
// **WILL ASSIGN THE ENTRYINFO.ID**
func AddEntry(uid int64, timezone string, entryInfo *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry, userViewingEntry *db_types.UserViewingEntry) error {
	id := rand.Int64()

	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	entryInfo.ItemId = id
	metadataEntry.ItemId = id
	userViewingEntry.ItemId = id

	ensureMetadataJsonNotEmpty(metadataEntry)
	ensureUserJsonNotEmpty(userViewingEntry)

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
		err := RegisterUserEvent(uid, db_types.UserViewingEvent{
			ItemId:    userViewingEntry.ItemId,
			Timestamp: uint64(time.Now().UnixMilli()),
			Event:     string(userViewingEntry.Status),
			TimeZone:  timezone,
			After:     0,
		})
		if err != nil {
			return err
		}
	}

	err = RegisterBasicUserEvent(uid, timezone, "Added", metadataEntry.ItemId)
	if err != nil {
		return err
	}

	// This should happen after the added event, because well, it was added, this file is a luxury thing
	err = WriteLocationFile(entryInfo)
	if err != nil {
		fmt.Printf("Error updating location file: %s\n", err.Error())
	}

	return nil
}

func RegisterUserEvent(uid int64, event db_types.UserViewingEvent) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	_, err = Db.Exec(`
		INSERT INTO userEventInfo (itemId, timestamp, event, after, timezone)
		VALUES (?, ?, ?, ?, ?)
	`, event.ItemId, event.Timestamp, event.Event, event.After, event.TimeZone)
	return err
}

func RegisterBasicUserEvent(uid int64, timezone string, event string, itemId int64) error {
	var e db_types.UserViewingEvent
	e.Event = event
	e.Timestamp = uint64(time.Now().UnixMilli())
	e.ItemId = itemId
	e.TimeZone = timezone
	return RegisterUserEvent(uid, e)
}

func UpdateUserViewingEntry(uid int64, entry *db_types.UserViewingEntry) error {
	ensureUserJsonNotEmpty(entry)
	return updateTable(uid, *entry, "userViewingInfo")
}

func MoveUserViewingEntry(uid int64, oldEntry *db_types.UserViewingEntry, newId int64) error {
	oldEntry.ItemId = newId
	return UpdateUserViewingEntry(uid, oldEntry)
}

func MoveUserEventEntries(uid int64, eventList []db_types.UserViewingEvent, newId int64) error {
	for _, e := range eventList {
		e.ItemId = newId
		err := RegisterUserEvent(uid, e)
		if err != nil {
			return err
		}
	}
	return nil
}

func ClearUserEventEntries(uid int64, id int64) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	_, err = Db.Exec(`
		DELETE FROM userEventInfo
		WHERE itemId = ?
	`, id)
	if err != nil {
		return err
	}

	return nil
}

func updateTable(uid int64, tblRepr db_types.TableRepresentation, tblName string) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

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

	_, err = Db.Exec(updateStr, updateArgs...)
	if err != nil {
		return err
	}

	return nil
}

func UpdateMetadataEntry(uid int64, entry *db_types.MetadataEntry) error {
	ensureMetadataJsonNotEmpty(entry)
	return updateTable(uid, *entry, "metadata")
}

func WriteLocationFile(entry *db_types.InfoEntry) error {
	if settings.Settings.WriteIdFile {
		location := entry.Location
		for k, v := range settings.Settings.LocationAliases {
			location = strings.Replace(location, "${"+k+"}", v, 1)
		}

		var aioIdPath string
		stat, err := os.Stat(location)
		if err == nil && !stat.IsDir() {
			dir := filepath.Dir(location)
			aioIdPath = filepath.Join(dir, ".AIO-ID")
		} else if err != nil {
			return err
		} else {
			aioIdPath = filepath.Join(location, ".AIO-ID")
		}

		err = os.WriteFile(aioIdPath, []byte(fmt.Sprintf("%d", entry.ItemId)), 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

func UpdateInfoEntry(uid int64, entry *db_types.InfoEntry) error {
	err := WriteLocationFile(entry)
	if err != nil {
		fmt.Printf("Error updating location file: %s\n", err.Error())
	}

	return updateTable(uid, *entry, "entryInfo")
}

func Delete(uid int64, id int64) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	transact, err := Db.Begin()
	if err != nil {
		return err
	}

	// item might have associated thumbnail, remove it
	aioPath := os.Getenv("AIO_DIR")
	thumbPath := fmt.Sprintf("%s/thumbnails/item-%d", aioPath, id)
	if _, err := os.Stat(thumbPath); err == nil {
		os.Remove(thumbPath)
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

func Search3(uid int64, searchQuery string) ([]db_types.InfoEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

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

func ListType(uid int64, col string, ty db_types.MediaTypes) ([]string, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	var out []string
	rows, err := Db.Query(`SELECT ? FROM entryInfo WHERE type = ?`, col, string(ty))
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

func GetCopiesOf(uid int64, id int64) ([]db_types.InfoEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

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

func GetChildren(uid int64, id int64) ([]db_types.InfoEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	var out []db_types.InfoEntry
	rows, err := Db.Query("SELECT * FROM entryInfo where parentId = ?", id)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

func DeleteEvent(uid int64, id int64, timestamp int64, after int64) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	_, err = Db.Exec(`
		DELETE FROM userEventInfo
		WHERE 
			itemId == ? and timestamp == ? and after == ?
	`, id, timestamp, after)
	return err
}

// /if id is -1, it lists all events
func GetEvents(uid int64, id int64) ([]db_types.UserViewingEvent, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	var out []db_types.UserViewingEvent

	var events *sql.Rows
	// if an id is given
	if id > -1 {
		events, err = Db.Query(`
		SELECT * from userEventInfo
		WHERE
			itemId == ?
		ORDER BY
			CASE timestamp
				WHEN 0 THEN
					userEventInfo.after
				ELSE timestamp
			END`, id)
		// otherweise the caller wants all events
	} else {
		events, err = Db.Query(`
		SELECT * from userEventInfo
		ORDER BY
			CASE timestamp
				WHEN 0 THEN
					userEventInfo.after
				ELSE timestamp
			END`, id)
	}
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

// /sort must be valid sql
func ListEntries(uid int64, sort string) ([]db_types.InfoEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	items, err := Db.Query(fmt.Sprintf(`
		SELECT entryInfo.*
		FROM
			entryInfo JOIN userViewingInfo
		ON
			entryInfo.itemId = userViewingInfo.itemId
		ORDER BY %s`, sort))
	if err != nil {
		return nil, err
	}

	var out []db_types.InfoEntry

	for items.Next() {
		var row db_types.InfoEntry
		err = row.ReadEntry(items)
		if err != nil {
			println(err.Error())
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func GetUserEntry(uid int64, itemId int64) (db_types.UserViewingEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	var row db_types.UserViewingEntry

	items, err := Db.Query("SELECT * FROM userViewingInfo WHERE itemId = ?;", itemId)
	if err != nil {
		return row, err
	}
	items.Next()
	err = row.ReadEntry(items)
	return row, err
}

func AllUserEntries(uid int64) ([]db_types.UserViewingEntry, error) {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	items, err := Db.Query("SELECT * FROM userViewingInfo")
	if err != nil {
		return nil, err
	}

	var out []db_types.UserViewingEntry
	for items.Next() {
		var row db_types.UserViewingEntry
		err := row.ReadEntry(items)
		if err != nil {
			println(err.Error())
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func getDescendants(uid int64, id int64, recurse uint64, maxRecurse uint64) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
	if recurse > maxRecurse {
		return out, nil
	}

	children, err := GetChildren(uid, id)
	if err != nil {
		return out, err
	}

	for _, item := range children {
		out = append(out, item)
		newItems, err := getDescendants(uid, item.ItemId, recurse+1, maxRecurse)
		if err != nil {
			continue
		}
		out = append(out, newItems...)
	}
	return out, nil
}

func GetDescendants(uid int64, id int64) ([]db_types.InfoEntry, error) {
	return getDescendants(uid, id, 0, 10)
}

func AddTags(uid int64, id int64, tags []string) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	tagsString := strings.Join(tags, "\x1F\x1F")
	_, err = Db.Exec("UPDATE entryInfo SET collection = (collection || char(31) || ? || char(31)) WHERE itemId = ?", tagsString, id)
	return err
}

func DelTags(uid int64, id int64, tags []string) error {
	Db, err := OpenUserDb(uid)
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	for _, tag := range tags {
		if tag == "" {
			continue
		}

		_, err = Db.Exec("UPDATE entryInfo SET collection = replace(collection, char(31) || ? || char(31), '')", tag)
		if err != nil {
			return err
		}
	}
	return nil
}
