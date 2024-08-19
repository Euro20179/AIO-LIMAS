package api

import (
	"fmt"
	"net/http"
	"strconv"

	db "aiolimas/db"
)

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request) {
	id, err := verifyIdQueryParam(req)
	if err != nil{
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
	if err != nil{
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d started\n", id)
}

func FinishMedia(w http.ResponseWriter, req *http.Request) {
	id, err := verifyIdQueryParam(req)
	if err != nil{
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
	if err != nil{
		wError(w, 400, "Not a number %s Be sure to provide a rating\n", rating)
		return
	}
	entry.UserRating = ratingN

	if err := entry.Finish(); err != nil {
		wError(w, 500, "Could not finish media\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(&entry)
	if err != nil{
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d finished\n", id)
}

func PlanMedia(w http.ResponseWriter, req *http.Request) {
	id, err := verifyIdQueryParam(req)
	if err != nil{
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
	if err != nil{
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}

func DropMedia(w http.ResponseWriter, req *http.Request) {
	id, err := verifyIdQueryParam(req)
	if err != nil{
		wError(w, 400, err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(id)
	if err != nil{
		wError(w, 400, "There is no entry with id %d\n", id)
		return
	}

	if !entry.CanDrop() {
		wError(w, 400, "%d cannot be planned\n", entry.ItemId)
		return
	}

	entry.Plan()
	err = db.UpdateUserViewingEntry(&entry)
	if err != nil{
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}
