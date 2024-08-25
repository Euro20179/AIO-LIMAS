package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	db "aiolimas/db"
)

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
		fmt.Fprintf(w, "This media is already being viewed, cannot start again\n")
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

func FinishMedia(w http.ResponseWriter, req *http.Request) {
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

	if !entry.CanFinish() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is not currently being viewed, cannot finish it\n")
		return
	}

	rating := req.URL.Query().Get("rating")
	ratingN, err := strconv.ParseFloat(rating, 64)
	if err != nil {
		wError(w, 400, "Not a number %s Be sure to provide a rating\n", rating)
		return
	}
	entry.UserRating = ratingN

	if err := entry.Finish(); err != nil {
		wError(w, 500, "Could not finish media\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d finished\n", id)
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

	entry.Plan()
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

func outputUserEntries(items *sql.Rows, w http.ResponseWriter) {
	w.WriteHeader(200)
	for items.Next() {
		var row db.UserViewingEntry
		err := row.ReadEntry(items)
		if err != nil{
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
	if err != nil{
		wError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}
	defer items.Close()
	outputUserEntries(items, w)
}
