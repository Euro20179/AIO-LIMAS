package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	db "aiolimas/db"
)

func CopyUserViewingEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	userEntry := parsedParams["src-id"].(db.UserViewingEntry)
	libraryEntry := parsedParams["dest-id"].(db.InfoEntry)

	oldId := userEntry.ItemId

	err := db.CopyUserViewingEntry(&userEntry, libraryEntry.ItemId)
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

	err = db.CopyUserEventEntries(events, libraryEntry.ItemId)
	if err != nil {
		wError(w, 500, "Failed to copy events\n%s", err.Error())
		return
	}
	// TODO: FIXME

	success(w)
}

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request) {
	id, err := verifyIdQueryParam(req)
	if err != nil {
		wError(w, 400, err.Error())
	}

	entry, err := db.GetUserViewEntryById(id)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "There is no entry with id %d\n", id)
		return
	}

	if !entry.CanBegin() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is already being viewed, or has not been planned, could not start\n")
		return
	}

	if err := entry.Begin(); err != nil {
		wError(w, 500, "Could not begin show\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d started\n", id)
}

func FinishMedia(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	entry := parsedParams["id"].(db.UserViewingEntry)

	if !entry.CanFinish() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is not currently being viewed, cannot finish it\n")
		return
	}

	rating := parsedParams["rating"].(float64)
	entry.UserRating = rating

	if err := entry.Finish(); err != nil {
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

func PlanMedia(w http.ResponseWriter, req *http.Request) {
	id, err := verifyIdQueryParam(req)
	if err != nil {
		wError(w, 400, err.Error())
	}

	entry, err := db.GetUserViewEntryById(id)
	if err != nil {
		wError(w, 400, "There is no entry with id %d\n", id)
		return
	}

	if !entry.CanPlan() {
		wError(w, 400, "%d can not be planned\n", entry.ItemId)
		return
	}

	entry.Plan()
	err = db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func DropMedia(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}

	if !entry.CanDrop() {
		wError(w, 400, "%d cannot be planned\n", entry.ItemId)
		return
	}

	entry.Drop()
	err = db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func PauseMedia(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}

	if !entry.CanPause() {
		wError(w, 400, "%d cannot be dropped\n", entry.ItemId)
		return
	}

	entry.Pause()

	err = db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func ResumeMedia(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}

	if !entry.CanResume() {
		wError(w, 400, "%d cannot be resumed\n", entry.ItemId)
		return
	}

	entry.Resume()
	err = db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func SetNote(w http.ResponseWriter, req *http.Request) {
	note := req.URL.Query().Get("note")

	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}

	entry.Notes = note
	err = db.UpdateUserViewingEntry(&entry)
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

	var user db.UserViewingEntry
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
		var row db.UserViewingEntry
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

func GetUserEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}
	items, err := db.Db.Query("SELECT * FROM userViewingInfo WHERE itemId = ?;", entry.ItemId)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	defer items.Close()
	outputUserEntries(items, w)
}

func UserEntries(w http.ResponseWriter, req *http.Request) {
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
		var event db.UserViewingEvent
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
	id := parsedParams["id"].(db.InfoEntry)

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
	id := parsedParams["id"].(db.InfoEntry)
	timestamp := parsedParams["timestamp"].(int64)
	after := parsedParams["after"].(int64)
	err := db.DeleteEvent(id.ItemId, timestamp, after)
	if err != nil{
		wError(w, 500, "Could not delete event\n%s", err.Error())
		return
	}
	success(w)
}

func ModUserEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	user := parsedParams["id"].(db.UserViewingEntry)

	user.Notes = parsedParams.Get("notes", user.Notes).(string)
	user.UserRating = parsedParams.Get("rating", user.UserRating).(float64)
	user.ViewCount = parsedParams.Get("view-count", user.ViewCount).(int64)
	user.CurrentPosition = parsedParams.Get("current-position", user.CurrentPosition).(string)
	user.Status = parsedParams.Get("status", user.Status).(db.Status)

	err := db.UpdateUserViewingEntry(&user)
	if err != nil {
		wError(w, 500, "Could not update user entry\n%s", err.Error())
		return
	}

	success(w)
}
