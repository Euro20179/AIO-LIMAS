package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	db "aiolimas/db"
	"aiolimas/settings"
	"aiolimas/types"
)

func CopyUserViewingEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	userEntry := parsedParams["src-id"].(db_types.UserViewingEntry)
	libraryEntry := parsedParams["dest-id"].(db_types.InfoEntry)

	oldId := userEntry.ItemId

	err := db.MoveUserViewingEntry(parsedParams["uid"].(int64), &userEntry, libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to reassociate entry\n%s", err.Error())
		return
	}

	err = db.ClearUserEventEntries(parsedParams["uid"].(int64), libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to clear event information\n%s", err.Error())
		return
	}

	events, err := db.GetEvents(parsedParams["uid"].(int64), oldId)
	if err != nil {
		wError(w, 500, "Failed to get events for item\n%s", err.Error())
		return
	}

	err = db.MoveUserEventEntries(parsedParams["uid"].(int64), events, libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to copy events\n%s", err.Error())
		return
	}

	success(w)
}

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)

	timezone := pp.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	if !entry.CanBegin() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is already being viewed, or has not been planned, could not start\n")
		return
	}

	if err := db.Begin(pp["uid"].(int64), timezone, &entry); err != nil {
		wError(w, 500, "Could not begin show\n%s", err.Error())
		return
	}

	err := db.UpdateUserViewingEntry(pp["uid"].(int64), &entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d started\n", entry.ItemId)
}

func FinishMedia(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	entry := parsedParams["id"].(db_types.UserViewingEntry)
	timezone := parsedParams.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	if !entry.CanFinish() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is not currently being viewed, cannot finish it\n")
		return
	}

	rating := parsedParams["rating"].(float64)
	entry.UserRating = rating

	if err := db.Finish(parsedParams["uid"].(int64), timezone, &entry); err != nil {
		wError(w, 500, "Could not finish media\n%s", err.Error())
		return
	}

	err := db.UpdateUserViewingEntry(parsedParams["uid"].(int64), &entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d finished\n", entry.ItemId)
}

func PlanMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)
	timezone := pp.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	if !entry.CanPlan() {
		wError(w, 400, "%d can not be planned\n", entry.ItemId)
		return
	}

	db.Plan(pp["uid"].(int64), timezone, &entry)
	err := db.UpdateUserViewingEntry(pp["uid"].(int64), &entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func DropMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)
	timezone := pp.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	db.Drop(pp["uid"].(int64), timezone, &entry)
	err := db.UpdateUserViewingEntry(pp["uid"].(int64), &entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func PauseMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)
	timezone := pp.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	if !entry.CanPause() {
		wError(w, 400, "%d cannot be dropped\n", entry.ItemId)
		return
	}

	db.Pause(pp["uid"].(int64), timezone, &entry)

	err := db.UpdateUserViewingEntry(pp["uid"].(int64), &entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func ResumeMedia(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)
	timezone := pp.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	if !entry.CanResume() {
		wError(w, 400, "%d cannot be resumed\n", entry.ItemId)
		return
	}

	db.Resume(pp["uid"].(int64), timezone, &entry)
	err := db.UpdateUserViewingEntry(pp["uid"].(int64), &entry)
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

	err = db.UpdateUserViewingEntry(parsedParams["uid"].(int64), &user)
	if err != nil {
		wError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(parsedParams["uid"].(int64), user.ItemId)
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

func outputUserEntry(item db_types.UserViewingEntry, w http.ResponseWriter) error{
	j, err := item.ToJson()
	if err != nil {
		println(err.Error())
		return err
	}
	w.Write(j)
	w.Write([]byte("\n"))
	return nil
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
		outputUserEntry(row, w)
	}
	w.Write([]byte("\n"))
}

func GetUserEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.UserViewingEntry)
	item, err := db.GetUserEntry(pp["uid"].(int64), entry.ItemId)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	outputUserEntry(item, w)
}

func UserEntries(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	items, err := db.AllUserEntries(pp["uid"].(int64))
	if err != nil {
		wError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}
	for _, item := range items {
		outputUserEntry(item, w)
	}
}

func ListEvents(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	events, err := db.GetEvents(parsedParams["uid"].(int64), -1)
	if err != nil {
		wError(w, 500, "Could not fetch events\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	for _, event := range events {
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

	events, err := db.GetEvents(parsedParams["uid"].(int64), id.ItemId)
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
	err := db.DeleteEvent(parsedParams["uid"].(int64), id.ItemId, timestamp, after)
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
	timezone := parsedParams.Get("timezone", settings.Settings.DefaultTimeZone).(string)

	err := db.RegisterUserEvent(parsedParams["uid"].(int64), db_types.UserViewingEvent{
		ItemId: id.ItemId,
		Timestamp: uint64(ts),
		After: uint64(after),
		Event: name,
		TimeZone: timezone,
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

	err := db.UpdateUserViewingEntry(parsedParams["uid"].(int64), &user)
	if err != nil {
		wError(w, 500, "Could not update user entry\n%s", err.Error())
		return
	}

	success(w)
}
