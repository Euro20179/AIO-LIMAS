package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"aiolimas/logging"
	log "aiolimas/logging"
	"aiolimas/search"

	"aiolimas/types"

	"github.com/mattn/go-sqlite3"
)

type RequestContext struct {
	UID  int64 // requested uid
	Auth int64 // authenticated uid
}

const DB_VERSION = 18

var DB *sql.DB

func DbRoot() string {
	aioPath := os.Getenv("AIO_DIR")
	return fmt.Sprintf("%s/", aioPath)
}

func OpenUserDb() (*sql.DB, error) {
	path := DbRoot()

	return sql.Open("sqlite3", path+"all.db")
}

func CkDBVersion() (int64, error) {
	v, err := DB.Query("PRAGMA user_version")
	if err != nil {
		return 0, err
	}
	defer v.Close()

	var version int64 = 0

	v.Next()
	err = v.Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func UpgradeDB(curversion int64) error {
	for i := curversion; i < DB_VERSION; i++ {
		schema, err := os.ReadFile(fmt.Sprintf("./db/schema/v%d-%d.sql", i, i+1))
		if err != nil {
			return err
		}

		println("Upgrading from", i, "to", i+1)

		_, err = DB.Exec(string(schema))
		if err != nil {
			return err
		}

		_, err = DB.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1))
		if err != nil {
			logging.ELog(err)
			return err
		}
	}

	return nil
}

func QueryDB(query string, args ...any) (*sql.Rows, error) {
	return DB.Query(query, args...)
}

func ExecUserDb(uid int64, query string, args ...any) error {
	_, err := DB.Exec(query, args...)
	return err
}

func InitDb() error {
	conn, err := OpenUserDb()
	if err != nil {
		panic(err.Error())
	}
	DB = conn

	DB.SetMaxIdleConns(1)
	DB.SetMaxOpenConns(1)

	sqlite3.Version()

	v, err := CkDBVersion()
	if err != nil {
		panic(err.Error())
	}

	if v == 0 {
		schema, err := os.ReadFile("./db/schema/schema.sql")
		if err != nil {
			return err
		}

		_, err = conn.Exec(string(schema))
		if err != nil {
			logging.ELog(err)
			return err
		}
	}
	v, err = CkDBVersion()
	if err != nil {
		panic(err.Error())
	}
	if v != DB_VERSION {
		UpgradeDB(v)
	}

	return nil
}

func uidWhere(ctx RequestContext, uidvar string, itemidvar string) string {
	whereClause := " "
	if ctx.UID != 0 {
		if ctx.UID != ctx.Auth {
			//if uid != authuid AND uid > 0, select items from uid with READ permission
			//since uid > 0, we dont have to special case authenticated items, since we are not requesting them
			whereClause += fmt.Sprintf(
				`WHERE %s = %d AND ((select permissions from entrySettings where entrySettings.itemid = %s) & %d) = %d`,
				uidvar,
				ctx.UID,
				itemidvar,
				db_types.PERM_READ,
				db_types.PERM_READ,
			)
		} else {
			//if uid == authid, select everything, they are authenticated
			whereClause += fmt.Sprintf(
				`WHERE %s = %d`, uidvar, ctx.UID,
			)
		}
	} else if ctx.Auth != 0{
		//if authenticated, but selecting from all users (uid 0), select uid = authuid (everything from that user) OR, where permission has PERM_READ
		whereClause += fmt.Sprintf(
			`WHERE %s = %d OR ((select permissions from entrySettings where itemid = %s AND permissions & %d) = %d)`,
			uidvar,
			ctx.Auth,
			itemidvar,
			db_types.PERM_READ,
			db_types.PERM_READ,
		)
	} else {
		//if authuid == uid AND uid == 0, select only PERM_READ items
		whereClause += fmt.Sprintf(
			`WHERE ((SELECT permissions from entrySettings WHERE itemid = %s AND permissions & %d) = %d)`,
			itemidvar, db_types.PERM_READ, db_types.PERM_READ,
		)
	}

	return whereClause
}

func BuildEntryTree(ctx RequestContext) (map[int64]db_types.EntryTree, error) {
	out := map[int64]db_types.EntryTree{}

	whereClause := uidWhere(ctx, "entryInfo.uid", "entryInfo.itemid")

	allRows, err := QueryDB(`SELECT * FROM entryInfo ` + whereClause)
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

		out[cur.EntryInfo.ItemId] = cur
	}
	allRows.Close()

	for id, cur := range out {
		var copies, children []db_types.InfoEntry

		cur.Copies = []string{}
		cur.Children = []string{}

		cur.UserInfo, err = GetUserViewEntryById(ctx, cur.EntryInfo.ItemId)
		if err != nil {
			goto next
		}

		cur.MetaInfo, err = GetMetadataEntryById(ctx, cur.EntryInfo.ItemId)
		if err != nil {
			log.ELog(err)
			goto next
		}

		children, err = GetRelation(ctx, cur.EntryInfo.ItemId, db_types.R_Child)
		if err != nil {
			log.ELog(err)
			goto next
		}

		for _, child := range children {
			cur.Children = append(cur.Children, fmt.Sprintf("%d", child.ItemId))
		}

		copies, err = GetCopiesOf(ctx, cur.EntryInfo.ItemId)
		if err != nil {
			log.ELog(err)
			goto next
		}

		for _, c := range copies {
			cur.Copies = append(cur.Copies, fmt.Sprintf("%d", c.ItemId))
		}

	next:
		out[id] = cur
	}

	return out, nil
}

func getById[T db_types.TableRepresentation](ctx RequestContext, id int64, tblName string, out *T) error {
	query := "SELECT * FROM " + tblName + uidWhere(ctx, fmt.Sprintf("%s.uid", tblName), fmt.Sprintf("%s.itemid", tblName))

	rows, err := QueryDB(query, id, ctx.UID)
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

func GetInfoEntryById(ctx RequestContext, id int64) (db_types.InfoEntry, error) {
	var res db_types.InfoEntry
	return res, getById(ctx, id, "entryInfo", &res)
}

func GetUserViewEntryById(ctx RequestContext, id int64) (db_types.UserViewingEntry, error) {
	var res db_types.UserViewingEntry
	return res, getById(ctx, id, "userViewingInfo", &res)
}

func GetUserEventEntryById(ctx RequestContext, id int64) (db_types.UserViewingEvent, error) {
	var res db_types.UserViewingEvent
	return res, getById(ctx, id, "userEventInfo", &res)
}

func GetMetadataEntryById(ctx RequestContext, id int64) (db_types.MetadataEntry, error) {
	var res db_types.MetadataEntry
	return res, getById(ctx, id, "metadata", &res)
}

func ensureUserJsonNotEmpty(user *db_types.UserViewingEntry) {
	if user.Extra == "" {
		user.Extra = "{}"
	}
}

func ensureRecommendedByNotEmpty(info *db_types.InfoEntry) {
	if info.RecommendedBy == "" {
		info.RecommendedBy = "[]"
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

func ListMetadata(ctx RequestContext) ([]db_types.MetadataEntry, error) {
	var items *sql.Rows
	var err error
	qs := "SELECT * FROM metadata " + uidWhere(ctx, "metadata.uid", "metadata.itemid")

	items, err = QueryDB(qs, ctx.UID)
	if err != nil {
		return nil, err
	}

	var out []db_types.MetadataEntry

	defer items.Close()

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

func Search3(ctx RequestContext, searchQuery string, orderby string) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry

	query := `SELECT DISTINCT entryInfo.*
	FROM entryInfo
	JOIN userViewingInfo ON
	entryInfo.itemId == userViewingInfo.itemId
	JOIN metadata ON
	entryInfo.itemId == metadata.itemId
	LEFT JOIN userEventInfo ON
	entryInfo.itemId == userEventInfo.itemId ` +
	uidWhere(ctx, "metadata.uid", "entryinfo.itemid") +
	" and %s"

	safeQuery, err := search.Search2String(searchQuery)
	if err != nil {
		log.ELog(err)
		return out, err
	}

	fullQuery := fmt.Sprintf(query, safeQuery)

	if orderby != "" {
		// TODO: make an option to toggle DESC
		safeOrderBy, err := search.Search2String(fmt.Sprintf("{ORDER BY %s DESC}", orderby))
		if err != nil {
			log.ELog(err)
			return out, err
		}
		fullQuery += " " + safeOrderBy
		log.Info("god order by %s", safeOrderBy)
	}

	log.Info("got query %s", safeQuery)

	rows, err := QueryDB(fullQuery)
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

func Search4(ctx RequestContext, searchQuery string, orderby string) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry

	searchQuery = "%" + searchQuery + "%"

	query := `SELECT DISTINCT entryInfo.*
	FROM entryInfo
	JOIN metadata ON
	entryInfo.itemId == metadata.itemId
	JOIN userViewingInfo ON
	entryInfo.itemId == userViewingInfo.itemId
	` + uidWhere(ctx, "metadata.uid", "entryInfo.itemid") + ` AND (
		En_Title LIKE ? or
		Title LIKE ? or
		entryInfo.Native_Title LIKE ? or
		metadata.Native_Title LIKE ?
	)
	`
	//parens are for if we want to add the uid condition
	//(it needs to happen separately)

	if orderby != "" {
		safeOrder, err := search.Search2String(fmt.Sprintf("{ORDER BY %s DESC}", orderby))
		if err != nil {
			log.ELog(err)
			return out, err
		}
		query += safeOrder
	}

	log.Info("got query %s", query)

	rows, err := QueryDB(query, searchQuery, searchQuery, searchQuery, searchQuery)
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

func ListType(ctx RequestContext, col string, ty db_types.MediaTypes) ([]string, error) {
	var out []string
	whereClause := uidWhere(ctx, "entryInfo.uid", "entryInfo.itemid") +  " AND type = ?"

	rows, err := QueryDB(`SELECT ? FROM entryInfo `+whereClause, col, string(ty), ctx.UID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
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

func GetCopiesOf(ctx RequestContext, id int64) ([]db_types.InfoEntry, error) {
	return GetRelation(ctx, id, db_types.R_Copy)
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

func GetRequires(ctx RequestContext, id int64) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
	whereClause := uidWhere(ctx, "entryInfo.uid", "entryInfo.itemid") + ` AND
		itemId IN (
			SELECT right FROM relations
			WHERE
			relation = 2 AND left = ?
		)`
	queryStr := fmt.Sprintf(`SELECT * FROM entryInfo %s`, whereClause)
	rows, err := QueryDB(queryStr, id)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

func GetRelation(ctx RequestContext, id int64, relation db_types.Relation) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
	whereClause := uidWhere(ctx, "ei.uid", "ei.itemid") +  " AND ei.itemid = ?"
	queryStr := fmt.Sprintf(`
	SELECT * FROM entryInfo
	WHERE
	entryInfo.itemId IN (
		SELECT left FROM relations r JOIN entryInfo ei ON r.right = ei.itemid AND r.relation = %d
		%s
	)`, relation, whereClause)
	rows, err := QueryDB(queryStr, id)
	if err != nil {
		return out, err
	}
	return mkRows(rows)
}

// if id is -1, it lists all events
func GetEvents(ctx RequestContext, id int64) ([]db_types.UserViewingEvent, error) {
	var out []db_types.UserViewingEvent

	whereClause := []string{}
	whereItems := []any{}
	if id > -1 {
		whereClause = append(whereClause, "itemId == ?")
		whereItems = append(whereItems, id)
	}
	whereClause = append(whereClause, strings.TrimPrefix(uidWhere(ctx, "userEventInfo.uid", "userEventInfo.itemid"), " WHERE"))

	whereText := ""
	if len(whereClause) != 0 {
		whereText = "WHERE " + strings.Join(whereClause, " and ")
	}

	var events *sql.Rows
	var err error
	events, err = QueryDB(fmt.Sprintf(`
	SELECT *, rowid from userEventInfo
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

	defer events.Close()

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

//Setting up a uidWhere for this may be tricky but it shouldn't matter
//too much anyway because all an outsider would see is A -> B without knowing what A and B are
func ListRelations(uid int64) (map[int64]db_types.Relations, error) {
	out := map[int64]db_types.Relations{}

	where := ""

	if uid != 0 {
		where = " WHERE uid = ?"
	}

	res, err := QueryDB("SELECT left, relation, right FROM relations"+where, uid)
	if err != nil {
		return out, err
	}

	defer res.Close()

	for res.Next() {
		var row struct {
			Left     int64
			Relation db_types.Relation
			Right    int64
		}

		res.Scan(&row.Left, &row.Relation, &row.Right)

		switch row.Relation {
		case db_types.R_Child:
			{
				r, ok := out[row.Right]
				if !ok {
					r = db_types.Relations{}
				}
				r.Children = append(out[row.Right].Children, row.Left)
				out[row.Right] = r
			}
		case db_types.R_Copy:
			{
				r, ok := out[row.Right]
				if !ok {
					r = db_types.Relations{}
				}

				r.Copies = append(r.Copies, row.Left)

				out[row.Right] = r

				// copies are symetrical, add to both
				r, ok = out[row.Left]
				if !ok {
					r = db_types.Relations{}
				}

				r.Copies = append(r.Copies, row.Right)

				out[row.Left] = r
			}
		case db_types.R_Requires:
			{
				r, ok := out[row.Left]
				if !ok {
					r = db_types.Relations{}
				}

				r.Requires = append(r.Requires, row.Right)
				out[row.Left] = r
			}
		}
	}

	return out, nil
}

// /sort must be valid sql
func ListEntries(ctx RequestContext, sort string) ([]db_types.InfoEntry, error) {
	whereClause := uidWhere(ctx, "entryInfo.uid", "entryInfo.itemid")
	qs := fmt.Sprintf(`
	SELECT entryInfo.*
	FROM
	entryInfo JOIN userViewingInfo
	ON
	entryInfo.itemId = userViewingInfo.itemId
	%s
	ORDER BY %s`, whereClause, sort)

	items, err := QueryDB(qs, ctx.UID)
	if err != nil {
		return nil, err
	}

	var out []db_types.InfoEntry

	defer items.Close()

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

func GetUserEntry(ctx RequestContext, itemId int64) (db_types.UserViewingEntry, error) {
	var row db_types.UserViewingEntry

	whereClause := uidWhere(ctx, "userViewingInfo.uid", "userViewingInfo.itemid") + " AND itemId = ?"

	items, err := QueryDB("SELECT * FROM userViewingInfo " + whereClause, itemId)
	if err != nil {
		return row, err
	}
	defer items.Close()
	if items.Next() {
		err = row.ReadEntry(items)
	} else {
		return row, errors.New("could not get entrf")
	}
	return row, err
}

func AllUserEntries(ctx RequestContext) ([]db_types.UserViewingEntry, error) {
	whereClause := uidWhere(ctx, "userViewingInfo.uid", "userViewingInfo.itemid")
	qs := "SELECT * FROM userViewingInfo " + whereClause
	items, err := QueryDB(qs, ctx.UID)
	if err != nil {
		return nil, err
	}

	defer items.Close()

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

func getDescendants(ctx RequestContext, id int64, recurse uint64, maxRecurse uint64) ([]db_types.InfoEntry, error) {
	var out []db_types.InfoEntry
	if recurse > maxRecurse {
		return out, nil
	}

	children, err := GetRelation(ctx, id, db_types.R_Child)
	if err != nil {
		return out, err
	}

	for _, item := range children {
		out = append(out, item)
		newItems, err := getDescendants(ctx, item.ItemId, recurse+1, maxRecurse)
		if err != nil {
			continue
		}
		out = append(out, newItems...)
	}
	return out, nil
}

func GetDescendants(ctx RequestContext, id int64) ([]db_types.InfoEntry, error) {
	return getDescendants(ctx, id, 0, 10)
}

func GetRecommendersList(ctx RequestContext) ([]string, error) {
	whereClause := uidWhere(ctx, "entryInfo.uid", "entryInfo.itemid") +  " AND recommendedBy != ''"
	rows, err := QueryDB("SELECT DISTINCT json_each.value from entryInfo, json_each(recommendedBy) " + whereClause)
	if err != nil {
		return []string{}, err
	}

	defer rows.Close()

	recommenders := []string{}
	for rows.Next() {
		var r string
		rows.Scan(&r)
		recommenders = append(recommenders, r)
	}
	return recommenders, nil
}

func ListTransactions(ctx RequestContext, itemid int64) ([]db_types.TransactionEntry, error) {
	rows, err := QueryDB(fmt.Sprintf(`
	SELECT rowid, * from transactions %s AND (? = 0 OR itemid = ?)
	`, uidWhere(ctx, "uid", "itemid")), itemid, itemid)
	if err != nil {
		return []db_types.TransactionEntry{}, err
	}

	defer rows.Close()

	out := []db_types.TransactionEntry{}
	for rows.Next() {
		cur := db_types.TransactionEntry{}
		cur.ReadEntry(rows)
		out = append(out, cur)
	}
	return out, nil
}

func GetTransaction(ctx RequestContext, id int64) (db_types.TransactionEntry, error) {
	whereClause := uidWhere(ctx, "transactions.uid", "transactions.itemid") + " AND rowid = ?"
	rows, err := QueryDB("select rowid, * from transactions " + whereClause, id)
	if err != nil {
		return db_types.TransactionEntry{}, err
	}

	defer rows.Close()

	rows.Next()
	e := db_types.TransactionEntry{}
	err = e.ReadEntry(rows)

	return e, err
}

func GetEvent(ctx RequestContext, eventID int64) (db_types.UserViewingEvent, error) {
	whereClause := uidWhere(ctx, "userEventInfo.uid", "userEventInfo.itemid") + " AND rowid = ?"
	rows, err := QueryDB("select *, rowid from userEventInfo " + whereClause, eventID)
	if err != nil {
		return db_types.UserViewingEvent{}, err
	}

	defer rows.Close()

	rows.Next()
	e := db_types.UserViewingEvent{}
	err = e.ReadEntry(rows)

	return e, err
}
