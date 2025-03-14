package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	db "aiolimas/db"
	"aiolimas/types"
)

func CopyUserViewingEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	userEntry := parsedParams["src-id"].(db_types.UserViewingEntry)
	libraryEntry := parsedParams["dest-id"].(db_types.InfoEntry)

	oldId := userEntry.ItemId

	err := db.MoveUserViewingEntry(&userEntry, libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to reassociate entry\n%s", err.Error())
		return
	}

	err = db.ClearUserEventEntries(libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to clear event information\n%s", err.Error())
		return
	}

	events, err := db.GetEvents(oldId)
	if err != nil {
		wError(w, 500, "Failed to get events for item\n%s", err.Error())
		return
	}

	err = db.MoveUserEventEntries(events, libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to copy events\n%s", err.Error())
		return
	}

	success(w)
}

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)

	if !entry.CanBegin() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is already being viewed, or has not been planned, could not start\n")
		return
	}

	if err := db.Begin(&entry); err != nil {
		wError(w, 500, "Could not begin show\n%s", err.Error())
		return
	}

	err := db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d started\n", entry.ItemId)
}

func FinishMedia(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	entry := parsedParams["id"].(db_types.UserViewingEntry)

	if !entry.CanFinish() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is not currently being viewed, cannot finish it\n")
		return
	}

	rating := parsedParams["rating"].(float64)
	entry.UserRating = rating

	if err := db.Finish(&entry); err != nil {
		wError(w, 500, "Could not finish media\n%s", err.Error())
		return
	}

	err := db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d finished\n", entry.ItemId)
}

func PlanMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)

	if !entry.CanPlan() {
		wError(w, 400, "%d can not be planned\n", entry.ItemId)
		return
	}

	db.Plan(&entry)
	err := db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func DropMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)

	db.Drop(&entry)
	err := db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func PauseMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)

	if !entry.CanPause() {
		wError(w, 400, "%d cannot be dropped\n", entry.ItemId)
		return
	}

	db.Pause(&entry)

	err := db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func ResumeMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)

	if !entry.CanResume() {
		wError(w, 400, "%d cannot be resumed\n", entry.ItemId)
		return
	}

	db.Resume(&entry)
	err := db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func SetUserEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		wError(w, 500, "Could not read body\n%s", err.Error())
		return
	}

	var user db_types.UserViewingEntry
	err = json.Unmarshal(data, &user)
	if err != nil {
		wError(w, 400, "Could not parse json\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(&user)
	if err != nil {
		wError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(user.ItemId)
	if err != nil{
		wError(w, 500, "Could not retrieve updated entry\n%s", err.Error())
		return
	}

	outJson, err := json.Marshal(entry)
	if err != nil{
		wError(w, 500, "Could not marshal new user entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(outJson)
}

func outputUserEntries(items *sql.Rows, w http.ResponseWriter) {
	w.WriteHeader(200)
	for items.Next() {
		var row db_types.UserViewingEntry
		err := row.ReadEntry(items)
		if err != nil {
			println(err.Error())
			continue
		}
		j, err := row.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
	w.Write([]byte("\n"))
}

func GetUserEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)
	items, err := db.Db.Query("SELECT * FROM userViewingInfo WHERE itemId = ?;", entry.ItemId)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	defer items.Close()
	outputUserEntries(items, w)
}

func UserEntries(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	items, err := db.Db.Query("SELECT * FROM userViewingInfo")
	if err != nil {
		wError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}
	defer items.Close()
	outputUserEntries(items, w)
}

func ListEvents(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	events, err := db.Db.Query(`
		SELECT * from userEventInfo
		ORDER BY
			CASE timestamp
				WHEN 0 THEN
					userEventInfo.after
				ELSE timestamp
			END`)
	if err != nil {
		wError(w, 500, "Could not fetch events\n%s", err.Error())
		return
	}
	defer events.Close()

	w.WriteHeader(200)
	for events.Next() {
		var event db_types.UserViewingEvent
		err := event.ReadEntry(events)
		if err != nil {
			println(err.Error())
			continue
		}

		j, err := event.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

func GetEventsOf(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	id := parsedParams["id"].(db_types.InfoEntry)

	events, err := db.GetEvents(id.ItemId)
	if err != nil {
		wError(w, 400, "Could not get events\n%s", err.Error())
		return
	}

	w.WriteHeader(200)

	for _, e := range events {
		j, err := e.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

func DeleteEvent(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	id := parsedParams["id"].(db_types.InfoEntry)
	timestamp := parsedParams["timestamp"].(int64)
	after := parsedParams["after"].(int64)
	err := db.DeleteEvent(id.ItemId, timestamp, after)
	if err != nil{
		wError(w, 500, "Could not delete event\n%s", err.Error())
		return
	}
	success(w)
}

func RegisterEvent(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	id := parsedParams["id"].(db_types.InfoEntry)
	ts := parsedParams.Get("timestamp", time.Now().UnixMilli()).(int64)
	after := parsedParams.Get("after", 0).(int64)
	name := parsedParams["name"].(string)

	err := db.RegisterUserEvent(db_types.UserViewingEvent{
		ItemId: id.ItemId,
		Timestamp: uint64(ts),
		After: uint64(after),
		Event: name,
	})
	if err != nil{
		wError(w, 500, "Could not register event\n%s", err.Error())
		return
	}
}

func ModUserEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	user := parsedParams["id"].(db_types.UserViewingEntry)

	user.Notes = parsedParams.Get("notes", user.Notes).(string)
	user.UserRating = parsedParams.Get("rating", user.UserRating).(float64)
	user.ViewCount = parsedParams.Get("view-count", user.ViewCount).(int64)
	user.CurrentPosition = parsedParams.Get("current-position", user.CurrentPosition).(string)
	user.Status = parsedParams.Get("status", user.Status).(db_types.Status)

	err := db.UpdateUserViewingEntry(&user)
	if err != nil {
		wError(w, 500, "Could not update user entry\n%s", err.Error())
		return
	}

	success(w)
}
