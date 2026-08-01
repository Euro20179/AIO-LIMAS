package db

import (
	"aiolimas/types"
	"errors"
	"time"
	"os"
	"fmt"
	"strings"
)

func Wait(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Waiting", entry.ItemId)
	if err != nil {
		return err
	}

	entry.Status = db_types.S_WAITING
	return nil
}

func Begin(uid int64, timezone string, entry *db_types.UserViewingEntry) error {
	err := RegisterBasicUserEvent(uid, timezone, "Started", entry.ItemId)
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


// TODO: remove timezone parameter from this function, maybe combine it witih userViewingEntry since that also keeps track of the timezone
// **WILL ASSIGN THE ENTRYINFO.ID**
// if timezone is empty, it will not add an Added event
// if entryInfo has an id, that id will be used
func AddEntry(uid int64, timezone string, entryInfo *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry, userViewingEntry *db_types.UserViewingEntry) error {
	id := entryInfo.ItemId
	if id == 0 {
		res, err := QueryDB("SELECT max(itemid) FROM entryInfo")
		if err != nil || !res.Next() {
			return errors.New("failed to add entry, could not determine id")
		}
		res.Scan(&id)
		res.Close()
		id += 1
	}

	entryInfo.Uid = uid
	entryInfo.ItemId = id
	metadataEntry.Uid = uid
	metadataEntry.ItemId = id
	userViewingEntry.Uid = uid
	userViewingEntry.ItemId = id

	ensureMetadataJsonNotEmpty(metadataEntry)
	ensureUserJsonNotEmpty(userViewingEntry)
	ensureRecommendedByNotEmpty(entryInfo)

	entries := map[string]db_types.TableRepresentation{
		"entryInfo":       *entryInfo,
		"metadata":        *metadataEntry,
		"userViewingInfo": *userViewingEntry,
	}

	for entryName, entry := range entries {
		entryData := db_types.StructNamesToDict(entry, map[string]string{})

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
		_, err := DB.Exec(entryQ, entryArgs...)
		if err != nil {
			return err
		}
	}

	if userViewingEntry.Status != db_types.Status("") && timezone != "" {
		eName := string(userViewingEntry.Status)
	 	switch(eName) {
			case "Viewing":
				eName = "Started"
			case "ReViewing":
				eName = "Started"
		}
		err := RegisterUserEvent(uid, db_types.UserViewingEvent{
			ItemId:    userViewingEntry.ItemId,
			Timestamp: int64(time.Now().UnixMilli()),
			Event:     eName,
			TimeZone:  timezone,
			After:     0,
		})
		if err != nil {
			return err
		}
	}

	if timezone != "" {
		event := "Added"
		err := RegisterBasicUserEvent(uid, timezone, event, metadataEntry.ItemId)
		if err != nil {
			return err
		}
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
	e.Timestamp = int64(time.Now().UnixMilli())
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

	data := db_types.StructNamesToDict(tblRepr, map[string]string{})

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

	err := ExecUserDb(uid, updateStr, updateArgs...)
	return err
}

func updateRowidTable(uid int64, rowid int64, tblRepr db_types.TableRepresentation, tblName string, replacements map[string]string) error {
	set := ""
	data := db_types.StructNamesToDict(tblRepr, replacements)
	updateArgs := []any{}

	for k, v := range data {
		if k == "eventId" || k == "transactionId" {
			continue
		}
		updateArgs = append(updateArgs, v)

		set += k + "= ?,"
	}
	updateArgs = append(updateArgs, rowid)
	set = set[:len(set)-1]
	return ExecUserDb(uid, `UPDATE ` + tblName + ` SET ` + set + ` WHERE rowid = ?`, updateArgs...)
}

func UpdateEvent(uid int64, event *db_types.UserViewingEvent) error {
	return updateRowidTable(uid, event.EventId, *event, "userEventInfo", map[string]string { "Before": "beforets"})
}

func DeleteTransaction(uid int64, id int64) error {
	return ExecUserDb(uid, `DELETE FROM transactions WHERE rowid = ?`, id)
}

func UpdateTransaction(uid int64, transaction *db_types.TransactionEntry) error {
	return updateRowidTable(uid, transaction.TransactionId, *transaction, "transactions", map[string]string{})
}

func UpdateMetadataEntry(uid int64, entry *db_types.MetadataEntry) error {
	ensureMetadataJsonNotEmpty(entry)
	return updateTable(uid, *entry, "metadata")
}

func UpdateInfoEntry(uid int64, entry *db_types.InfoEntry) error {
	ensureRecommendedByNotEmpty(entry)
	return updateTable(uid, *entry, "entryInfo")
}

func Delete(uid int64, id int64) error {
	transact, err := DB.Begin()
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
	transact.Exec(`DELETE FROM relations WHERE left = ? or right = ?`, id, id)
	transact.Exec(`DELETE FROM transactions WHERE itemid = ?`, id)

	return transact.Commit()
}

func DeleteByUID(uid int64) error {
	transact, err := DB.Begin()
	if err != nil {
		return err
	}

	transact.Exec(`DELETE FROM entryInfo WHERE entryInfo.uid = ?`, uid)
	transact.Exec(`DELETE FROM metadata WHERE metadata.uid = ?`, uid)
	transact.Exec(`DELETE FROM userViewingInfo WHERE userViewingInfo.uid = ?`, uid)
	transact.Exec(`DELETE FROM userEventInfo WHERE userEventInfo.uid = ?`, uid)
	transact.Exec(`DELETE FROM relations WHERE userEventInfo.uid = ?`, uid)
	transact.Exec(`DELETE FROM transactions WHERE userEventInfo.uid = ?`, uid)

	return transact.Commit()
}

func DeleteEvent(uid int64, id int64, timestamp int64, after int64, before int64) error {
	return ExecUserDb(uid, `
		DELETE FROM userEventInfo
		WHERE 
			itemId == ? and timestamp == ? and after == ? and beforeTS == ? and userEventInfo.uid = ?
	`, id, timestamp, after, before, uid)
}

func DeletEventV2(uid int64, id int64) error {
	return ExecUserDb(uid, `DELETE FROM userEventInfo WHERE rowid == ?`, id)
}

func BecomeOriginal(uid int64, itemid int64) error{
	return ExecUserDb(uid, `
		DELETE FROM relations WHERE left = ? or right = ? and relation = ?
	`, itemid, itemid, db_types.R_Copy)
}

func SetParent(uid int64, itemid int64, parent int64) error {
	if uid == 0 {
		return errors.New("uid cannot be 0 to set a parent")
	}

	if err := BecomeOrphan(uid, itemid); err != nil {
		return err
	}

	return ExecUserDb(uid, `
		INSER INTO relations (uid, left, relation, right)
		VALUES
		(?, ?, ?, ?)
	`, uid, itemid, db_types.R_Child, parent)
}

func SetCopy(uid int64, itemid int64, copyof int64) error {
	if uid == 0 {
		return errors.New("uid cannot be 0 to set a copy")
	}

	err := BecomeOriginal(uid, itemid)
	if err != nil{
		return err
	}

	return ExecUserDb(uid, `
		INSERT INTO relations (uid, left, relation, right)
		VALUES
		(?, ?, ?, ?)
	`, uid, itemid, db_types.R_Copy, copyof)
}

func BecomeOrphan(uid int64, itemid int64) error {
	return ExecUserDb(uid, `
		DELETE FROM relations WHERE left = ? and relation = ?
	`, itemid, db_types.R_Child)
}

func AddRelation(uid int64, left int64, relation db_types.Relation, right int64) error {
	if uid == 0 {
		return errors.New("uid cannot be 0 to add a relation")
	}
	return ExecUserDb(uid, `
		INSERT INTO relations (uid, left, relation, right)
		VALUES (?, ?, ?, ?)
`, uid, left, relation, right)
}

func DelRelation(uid int64, left int64, relation db_types.Relation, right int64) error {
	return ExecUserDb(uid, `
		DELETE FROM relations WHERE left = ? and relation = ? and right = ?
`, left, relation, right)
}

func AddTags(uid int64, id int64, tags []string) error {
	tagsString := strings.Join(tags, "\x1F\x1F")
	return ExecUserDb(uid, "UPDATE entryInfo SET collection = (collection || char(31) || ? || char(31)) WHERE itemId = ? and entryInfo.uid = ?", tagsString, id, uid)
}

func DelTags(uid int64, id int64, tags []string) error {
	for _, tag := range tags {
		if tag == "" {
			continue
		}

		_, err := DB.Exec("UPDATE entryInfo SET collection = replace(collection, char(31) || ? || char(31), '') WHERE itemId = ? and entryInfo.uid = ?", tag, id, uid)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateTransaction(transactionType db_types.Transaction, uid int64, itemId int64, eventId int64, timezone string, price float64, currency string) error {
	if uid == 0 {
		return errors.New("uid cannot be 0 for creating a transaction")
	}

	if eventId == 0 {
		if err := RegisterBasicUserEvent(uid, timezone, string(transactionType), itemId); err != nil {
			return err;
		}
		return ExecUserDb(uid, `
			INSERT INTO transactions VALUES (
				?,
				?,
				(SELECT MAX(rowid) FROM userEventInfo WHERE uid = ? AND itemId = ?),
				?,
				?
			)
		`, uid, itemId, uid, itemId, price, currency)
	}

	return ExecUserDb(uid, `
		INSERT INTO transactions VALUES (
			?,
			?,
			?,
			?,
			?
	`, uid, itemId, eventId, price, currency)
}

