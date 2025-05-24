package db

import (
	"database/sql"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aiolimas/logging"
	log "aiolimas/logging"
	"aiolimas/search"
	"aiolimas/settings"

	"aiolimas/types"

	"github.com/mattn/go-sqlite3"
)

const DB_VERSION = 2

func DbRoot() string {
	aioPath := os.Getenv("AIO_DIR")
	return fmt.Sprintf("%s/", aioPath)
}

func OpenUserDb() (*sql.DB, error) {
	path := DbRoot()

	return sql.Open("sqlite3", path+"all.db")
}

func CkDBVersion() error {
	DB, err := OpenUserDb()
	if err != nil {
		return err
	}

	defer DB.Close()

	DB.Exec("CREATE TABLE IF NOT EXISTS DBInfo (version INTEGER DEFAULT 0)")

	v, err := DB.Query("SELECT version FROM DBInfo")
	if err != nil {
		return err
	}

	var version int64 = 0

	if !v.Next() {
		logging.Info("COULD NOT DETERMINE DB VERSION, USING VERSION 0")
		var cont int64
		print("type 1 if you are SURE that this is correct: ")
		fmt.Scanln(&cont)
		if cont != 1 {
			panic("Could not determine db veresion")
		}
	} else {
		err = v.Scan(&version)
		if err != nil {
			return err
		}
	}

	v.Close()

	for i := version; i < DB_VERSION; i++ {
		schema, err := os.ReadFile(fmt.Sprintf("./db/schema/v%d-%d.sql", i, i+1))
		if err != nil {
			return err
		}

		println("Upgrading from", i, "to", i+1)

		_, err = DB.Exec(string(schema))
		if err != nil {
			return err
		}
	}

	return nil
}

func QueryDB(query string, args ...any) (*sql.Rows, error) {
	Db, err := OpenUserDb()
	if err != nil {
		log.ELog(err)
		return nil, err
	}
	defer Db.Close()

	return Db.Query(query, args...)
}

func ExecUserDb(uid int64, query string, args ...any) error {
	Db, err := OpenUserDb()
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	_, err = Db.Exec(query, args...)
	return err
}

func BuildEntryTree(uid int64) (map[int64]db_types.EntryTree, error) {
	out := map[int64]db_types.EntryTree{}

	whereClause := ""
	if uid != 0 {
		whereClause = "WHERE entryInfo.uid = ?"
	}

	allRows, err := QueryDB(`SELECT * FROM entryInfo `, whereClause, uid)
	if err != nil {
		log.ELog(err)
		return out, err
	}

	for allRows.Next() {
		var cur db_types.EntryTree

		err := cur.EntryInfo.ReadEntry(allRows)
		if err != nil {
			log.ELog(err)
			continue
		}
		cur.UserInfo, err = GetUserViewEntryById(uid, cur.EntryInfo.ItemId)
		if err != nil {
			log.ELog(err)
			continue
		}

		cur.MetaInfo, err = GetMetadataEntryById(uid, cur.EntryInfo.ItemId)
		if err != nil {
			log.ELog(err)
			continue
		}

		children, err := GetChildren(uid, cur.EntryInfo.ItemId)
		if err != nil {
			log.ELog(err)
			continue
		}

		for _, child := range children {
			cur.Children = append(cur.Children, fmt.Sprintf("%d", child.ItemId))
		}

		copies, err := GetCopiesOf(uid, cur.EntryInfo.ItemId)
		if err != nil {
			log.ELog(err)
			continue
		}

		for _, c := range copies {
			cur.Copies = append(cur.Copies, fmt.Sprintf("%d", c.ItemId))
		}

		out[cur.EntryInfo.ItemId] = cur
	}

	return out, nil
}

func Wait(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Waiting", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_WAITING
	return nil
}

func Begin(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Viewing", entry.ItemId)
	if err != nil {
		return err
	}

	if entry.Status == db_types.S_FINISHED {
		entry.Status = db_types.S_REVIEWING
	} else {
		entry.Status = db_types.S_VIEWING
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
	err := RegisterBasicUserEvent(uid, timezone, "Resuming", entry.ItemId)
	if err != nil {
		return err
	}

	if entry.ViewCount == 0 {
		entry.Status = db_types.S_VIEWING
	} else {
		entry.Status = db_types.S_REVIEWING
	}
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

func InitDb() error {
	err := CkDBVersion()
	if err != nil {
		panic(err.Error())
	}
	conn, err := OpenUserDb()
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	sqlite3.Version()

	schema, err := os.ReadFile("./db/schema/schema.sql")
	if err != nil {
		return err
	}

	_, err = conn.Exec(string(schema))
	if err != nil {
		logging.ELog(err)
		return err
	}

	return nil
}

func getById[T db_types.TableRepresentation](uid int64, id int64, tblName string, out *T) error {
	query := "SELECT * FROM " + tblName + " WHERE itemId = ?"
	if uid != 0 {
		query += " and " + tblName + ".uid = ?;"
	}

	rows, err := QueryDB(query, id, uid)
	if err != nil {
		return err
	}

	defer rows.Close()

	hasEntry := rows.Next()
	if !hasEntry {
		return fmt.Errorf("could not find id %d", id)
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
	var items *sql.Rows
	var err error
	qs := "SELECT * FROM metadata WHERE metadata.uid = ?"
	if uid == 0 {
		qs = "SELECT * FROM metadata"
	}

	items, err = QueryDB(qs, uid)
	if err != nil {
		return nil, err
	}

	var out []db_types.MetadataEntry

	i := 0
	for items.Next() {
		i++
		var row db_types.MetadataEntry
		err := row.ReadEntry(items)
		if err != nil {
			log.ELog(err)
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

// TODO: remove timezone parameter from this function, maybe combine it witih userViewingEntry since that also keeps track of the timezone
// **WILL ASSIGN THE ENTRYINFO.ID**
// if timezone is empty, it will not add an Added event
// if entryInfo has an id, that id will be used
func AddEntry(uid int64, timezone string, entryInfo *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry, userViewingEntry *db_types.UserViewingEntry) error {
	id := entryInfo.ItemId
	if id == 0 {
		id = rand.Int64()
	}

	Db, err := OpenUserDb()
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	entryInfo.Uid = uid
	entryInfo.ItemId = id
	metadataEntry.Uid = uid
	metadataEntry.ItemId = id
	userViewingEntry.Uid = uid
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

		// add uid last
		entryArgs = append(entryArgs, uid)
		entryQ += "uid"
		questionMarks += "?"

		// add final paren
		entryQ = entryQ + ")"

		entryQ += "VALUES(" + questionMarks + ")"
		_, err := Db.Exec(entryQ, entryArgs...)
		if err != nil {
			return err
		}
	}

	if userViewingEntry.Status != db_types.Status("") && timezone != "" {
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

	if timezone != "" {
		err = RegisterBasicUserEvent(uid, timezone, "Added", metadataEntry.ItemId)
		if err != nil {
			return err
		}
	}

	us, err := settings.GetUserSettings(uid)
	if err != nil {
		return err
	}
	// This should happen after the added event, because well, it was added, this file is a luxury thing
	if us.WriteIdFile {
		err = WriteLocationFile(entryInfo, us.LocationAliases)
	}
	if err != nil {
		fmt.Printf("Error updating location file: %s\n", err.Error())
	}

	return nil
}

func RegisterUserEvent(uid int64, event db_types.UserViewingEvent) error {
	return ExecUserDb(uid, `
		INSERT INTO userEventInfo (uid, itemId, timestamp, event, after, timezone, beforeTS)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, uid, event.ItemId, event.Timestamp, event.Event, event.After, event.TimeZone, event.Before)
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
	ExecUserDb(uid, `
		DELETE FROM userEventInfo
		WHERE itemId = ? and uid = ?
	`, id, uid)

	return nil
}

func updateTable(uid int64, tblRepr db_types.TableRepresentation, tblName string) error {
	updateStr := `UPDATE ` + tblName + ` SET `

	data := db_types.StructNamesToDict(tblRepr)

	updateArgs := []any{}

	for k, v := range data {
		updateArgs = append(updateArgs, v)

		updateStr += k + "= ?,"
	}

	// append the user id
	updateArgs = append(updateArgs, uid)
	// needs itemid for checking which item to update
	updateArgs = append(updateArgs, tblRepr.Id())

	// remove final trailing comma
	updateStr = updateStr[:len(updateStr)-1]
	updateStr += "\nWHERE " + tblName + ".uid = ? and itemId = ?"

	return ExecUserDb(uid, updateStr, updateArgs...)
}

func UpdateMetadataEntry(uid int64, entry *db_types.MetadataEntry) error {
	ensureMetadataJsonNotEmpty(entry)
	return updateTable(uid, *entry, "metadata")
}

func WriteLocationFile(entry *db_types.InfoEntry, aliases map[string]string) error {
	location := settings.ExpandPathWithLocationAliases(aliases, entry.Location)

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

	err = os.WriteFile(aioIdPath, fmt.Appendf(nil, "%d", entry.ItemId), 0o644)
	if err != nil {
		return err
	}
	return nil
}

func UpdateInfoEntry(uid int64, entry *db_types.InfoEntry) error {
	us, err := settings.GetUserSettings(uid)
	if err != nil {
		return err
	}
	if us.WriteIdFile {
		err := WriteLocationFile(entry, us.LocationAliases)
		if err != nil {
			fmt.Printf("Error updating location file: %s\n", err.Error())
		}
	}

	return updateTable(uid, *entry, "entryInfo")
}

func Delete(uid int64, id int64) error {
	Db, err := OpenUserDb()
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

	transact.Exec(`DELETE FROM entryInfo WHERE itemId = ? and entryInfo.uid = ?`, id, uid)
	transact.Exec(`DELETE FROM metadata WHERE itemId = ? and metadata.uid = ?`, id, uid)
	transact.Exec(`DELETE FROM userViewingInfo WHERE itemId = ? and userViewingInfo.uid = ?`, id, uid)
	transact.Exec(`DELETE FROM userEventInfo WHERE itemId = ? and userEventInfo.uid = ?`, id, uid)

	return transact.Commit()
}

func DeleteByUID(uid int64) error {
	Db, err := OpenUserDb()
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	transact, err := Db.Begin()
	if err != nil {
		return err
	}

	transact.Exec(`DELETE FROM entryInfo WHERE entryInfo.uid = ?`, uid)
	transact.Exec(`DELETE FROM metadata WHERE metadata.uid = ?`, uid)
	transact.Exec(`DELETE FROM userViewingInfo WHERE userViewingInfo.uid = ?`, uid)
	transact.Exec(`DELETE FROM userEventInfo WHERE userEventInfo.uid = ?`, uid)

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

func Search3(searchQuery string) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry

	query := `SELECT DISTINCT entryInfo.*
				FROM entryInfo
				JOIN userViewingInfo ON
					entryInfo.itemId == userViewingInfo.itemId
				JOIN metadata ON
					entryInfo.itemId == metadata.itemId
				LEFT JOIN userEventInfo ON
					entryInfo.itemId == userEventInfo.itemId
				WHERE %s`

	safeQuery, err := search.Search2String(searchQuery)
	if err != nil {
		log.ELog(err)
		return out, err
	}

	log.Info("got query %s", safeQuery)

	rows, err := QueryDB(fmt.Sprintf(query, safeQuery))
	if err != nil {
		return out, err
	}

	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var row db_types.InfoEntry
		err = row.ReadEntry(rows)
		if err != nil {
			log.ELog(err)
			continue
		}
		out = append(out, row)
	}

	return out, nil
}

func ListType(uid int64, col string, ty db_types.MediaTypes) ([]string, error) {
	var out []string
	whereClause := "WHERE type = ?"
	if uid != 0 {
		whereClause += " and entryInfo.uid = ?"
	}
	rows, err := QueryDB(`SELECT ? FROM entryInfo `+whereClause, col, string(ty), uid)
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
	var out []db_types.InfoEntry
	whereClause := "copyOf = ?"
	if uid != 0 {
		whereClause += " and entryInfo.uid = ?"
	}
	rows, err := QueryDB("SELECT * FROM entryInfo WHERE "+whereClause, id, uid)
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
	var out []db_types.InfoEntry
	whereClause := "parentId = ?"
	if uid != 0 {
		whereClause += " and entryInfo.uid = ?"
	}
	rows, err := QueryDB("SELECT * FROM entryInfo where "+whereClause, id, uid)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

func DeleteEvent(uid int64, id int64, timestamp int64, after int64) error {
	return ExecUserDb(uid, `
		DELETE FROM userEventInfo
		WHERE 
			itemId == ? and timestamp == ? and after == ? and userEventInfo.uid = ?
	`, id, timestamp, after, uid)
}

// if id is -1, it lists all events
func GetEvents(uid int64, id int64) ([]db_types.UserViewingEvent, error) {
	var out []db_types.UserViewingEvent

	whereClause := []string{}
	whereItems := []any{}
	if id > -1 {
		whereClause = append(whereClause, "itemId == ?")
		whereItems = append(whereItems, id)
	}
	if uid != 0 {
		whereClause = append(whereClause, "userEventInfo.uid = ?")
		whereItems = append(whereItems, uid)
	}

	whereText := ""
	if len(whereClause) != 0 {
		whereText = "WHERE " + strings.Join(whereClause, " and ")
	}

	var events *sql.Rows
	var err error
	events, err = QueryDB(fmt.Sprintf(`
	SELECT * from userEventInfo
	%s
	ORDER BY
		CASE timestamp
			WHEN 0 THEN
				userEventInfo.after
			ELSE timestamp
		END`, whereText), whereItems...)
	if err != nil {
		return out, err
	}

	for events.Next() {
		var event db_types.UserViewingEvent
		err := event.ReadEntry(events)
		if err != nil {
			log.ELog(err)
			continue
		}
		out = append(out, event)
	}
	return out, nil
}

// /sort must be valid sql
func ListEntries(uid int64, sort string) ([]db_types.InfoEntry, error) {
	whereClause := ""
	if uid != 0 {
		whereClause = "WHERE entryInfo.uid = ?"
	}
	qs := fmt.Sprintf(`
		SELECT entryInfo.*
		FROM
			entryInfo JOIN userViewingInfo
		ON
			entryInfo.itemId = userViewingInfo.itemId
		%s
		ORDER BY %s`, whereClause, sort)

	items, err := QueryDB(qs, uid)
	if err != nil {
		return nil, err
	}

	var out []db_types.InfoEntry

	for items.Next() {
		var row db_types.InfoEntry
		err = row.ReadEntry(items)
		if err != nil {
			log.ELog(err)
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func GetUserEntry(uid int64, itemId int64) (db_types.UserViewingEntry, error) {
	var row db_types.UserViewingEntry

	items, err := QueryDB("SELECT * FROM userViewingInfo WHERE itemId = ? and userViewingInfo.uid = ?;", itemId, uid)
	if err != nil {
		return row, err
	}
	items.Next()
	err = row.ReadEntry(items)
	return row, err
}

func AllUserEntries(uid int64) ([]db_types.UserViewingEntry, error) {
	qs := "SELECT * FROM userViewingInfo WHERE userViewingInfo.uid = ?"
	if uid == 0 {
		qs = "SELECT * FROM userViewingInfo"
	}
	items, err := QueryDB(qs, uid)
	if err != nil {
		return nil, err
	}

	var out []db_types.UserViewingEntry
	for items.Next() {
		var row db_types.UserViewingEntry
		err := row.ReadEntry(items)
		if err != nil {
			log.ELog(err)
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
	tagsString := strings.Join(tags, "\x1F\x1F")
	return ExecUserDb(uid, "UPDATE entryInfo SET collection = (collection || char(31) || ? || char(31)) WHERE itemId = ? and entryInfo.uid = ?", tagsString, id, uid)
}

func DelTags(uid int64, id int64, tags []string) error {
	Db, err := OpenUserDb()
	if err != nil {
		panic(err.Error())
	}
	defer Db.Close()

	for _, tag := range tags {
		if tag == "" {
			continue
		}

		_, err = Db.Exec("UPDATE entryInfo SET collection = replace(collection, char(31) || ? || char(31), '') WHERE itemId = ? and entryInfo.uid = ?", tag, id, uid)
		if err != nil {
			return err
		}
	}
	return nil
}
