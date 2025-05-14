package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aiolimas/util"
	db "aiolimas/db"
	"aiolimas/settings"
	"aiolimas/types"
	"aiolimas/logging"
)

func CopyUserViewingEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	userEntry := parsedParams["src-id"].(db_types.UserViewingEntry)
	libraryEntry := parsedParams["dest-id"].(db_types.InfoEntry)

	oldId := userEntry.ItemId

	err := db.MoveUserViewingEntry(ctx.Uid, &userEntry, libraryEntry.ItemId)
	if err != nil {
		util.WError(w, 500, "Failed to reassociate entry\n%s", err.Error())
		return
	}

	err = db.ClearUserEventEntries(ctx.Uid, libraryEntry.ItemId)
	if err != nil {
		util.WError(w, 500, "Failed to clear event information\n%s", err.Error())
		return
	}

	events, err := db.GetEvents(ctx.Uid, oldId)
	if err != nil {
		util.WError(w, 500, "Failed to get events for item\n%s", err.Error())
		return
	}

	err = db.MoveUserEventEntries(ctx.Uid, events, libraryEntry.ItemId)
	if err != nil {
		util.WError(w, 500, "Failed to copy events\n%s", err.Error())
		return
	}

	success(w)
}

func WaitMedia(ctx RequestContext) {
	entry := ctx.PP["id"].(db_types.UserViewingEntry)

	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(ctx.W, 500, "Could not update entry\n%s", err.Error())
		return
	}

	timezone := ctx.PP.Get("timezone", us.DefaultTimeZone).(string)

	if !entry.CanWait() {
		ctx.W.WriteHeader(405)
		fmt.Fprintf(ctx.W, "This media is not being viewed, could not set status to waiting\n")
		return
	}

	if err := db.Wait(ctx.Uid, timezone, &entry); err != nil {
		util.WError(ctx.W, 500, "Could not begin show\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(ctx.W, 500, "Could not update entry\n%s", err.Error())
		return
	}

	ctx.W.WriteHeader(200)
	fmt.Fprintf(ctx.W, "%d waited\n", entry.ItemId)
}

// engagement endpoints
func BeginMedia(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.UserViewingEntry)

	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	timezone := pp.Get("timezone", us.DefaultTimeZone).(string)

	if !entry.CanBegin() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is already being viewed, could not start\n")
		return
	}

	if err := db.Begin(ctx.Uid, timezone, &entry); err != nil {
		util.WError(w, 500, "Could not begin show\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d started\n", entry.ItemId)
}

func FinishMedia(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	entry := parsedParams["id"].(db_types.UserViewingEntry)

	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := parsedParams.Get("timezone", us.DefaultTimeZone).(string)

	if !entry.CanFinish() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is not currently being viewed, cannot finish it\n")
		return
	}

	rating := parsedParams["rating"].(float64)
	entry.UserRating = rating

	if err := db.Finish(ctx.Uid, timezone, &entry); err != nil {
		util.WError(w, 500, "Could not finish media\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Fprintf(w, "%d finished\n", entry.ItemId)
}

func PlanMedia(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.UserViewingEntry)
	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := pp.Get("timezone", us.DefaultTimeZone).(string)

	if !entry.CanPlan() {
		util.WError(w, 400, "%d can not be planned\n", entry.ItemId)
		return
	}

	db.Plan(ctx.Uid, timezone, &entry)
	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func DropMedia(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.UserViewingEntry)
	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := pp.Get("timezone", us.DefaultTimeZone).(string)

	db.Drop(ctx.Uid, timezone, &entry)
	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func PauseMedia(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.UserViewingEntry)
	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := pp.Get("timezone", us.DefaultTimeZone).(string)

	if !entry.CanPause() {
		util.WError(w, 400, "%d cannot be dropped\n", entry.ItemId)
		return
	}

	db.Pause(ctx.Uid, timezone, &entry)

	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func ResumeMedia(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.UserViewingEntry)
	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := pp.Get("timezone", us.DefaultTimeZone).(string)

	if !entry.CanResume() {
		util.WError(w, 400, "%d cannot be resumed\n", entry.ItemId)
		return
	}

	db.Resume(ctx.Uid, timezone, &entry)
	err = db.UpdateUserViewingEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}

	success(w)
}

func SetUserEntry(ctx RequestContext) {
	w := ctx.W
	req := ctx.Req
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		util.WError(w, 500, "Could not read body\n%s", err.Error())
		return
	}

	var user db_types.UserViewingEntry
	err = json.Unmarshal(data, &user)
	if err != nil {
		util.WError(w, 400, "Could not parse json\n%s", err.Error())
		return
	}

	err = db.UpdateUserViewingEntry(ctx.Uid, &user)
	if err != nil {
		util.WError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(ctx.Uid, user.ItemId)
	if err != nil{
		util.WError(w, 500, "Could not retrieve updated entry\n%s", err.Error())
		return
	}

	outJson, err := json.Marshal(entry)
	if err != nil{
		util.WError(w, 500, "Could not marshal new user entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(outJson)
}

func outputUserEntry(item db_types.UserViewingEntry, w http.ResponseWriter) error{
	j, err := item.ToJson()
	if err != nil {
		logging.ELog(err)
		return err
	}
	w.Write(j)
	w.Write([]byte("\n"))
	return nil
}

func GetUserEntry(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.UserViewingEntry)
	item, err := db.GetUserEntry(ctx.Uid, entry.ItemId)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	outputUserEntry(item, w)
}

func UserEntries(ctx RequestContext) {
	w := ctx.W
	items, err := db.AllUserEntries(ctx.Uid)
	if err != nil {
		util.WError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}
	for _, item := range items {
		outputUserEntry(item, w)
	}
}

func ListEvents(ctx RequestContext) {
	w := ctx.W
	events, err := db.GetEvents(ctx.Uid, -1)
	if err != nil {
		util.WError(w, 500, "Could not fetch events\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	for _, event := range events {
		j, err := event.ToJson()
		if err != nil {
			logging.ELog(err)
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

func GetEventsOf(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	id := parsedParams["id"].(db_types.InfoEntry)

	events, err := db.GetEvents(ctx.Uid, id.ItemId)
	if err != nil {
		util.WError(w, 400, "Could not get events\n%s", err.Error())
		return
	}

	w.WriteHeader(200)

	for _, e := range events {
		j, err := e.ToJson()
		if err != nil {
			logging.ELog(err)
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

func DeleteEvent(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	id := parsedParams["id"].(db_types.InfoEntry)
	timestamp := parsedParams["timestamp"].(int64)
	after := parsedParams["after"].(int64)
	err := db.DeleteEvent(ctx.Uid, id.ItemId, timestamp, after)
	if err != nil{
		util.WError(w, 500, "Could not delete event\n%s", err.Error())
		return
	}
	success(w)
}

func RegisterEvent(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	id := parsedParams["id"].(db_types.InfoEntry)
	ts := parsedParams.Get("timestamp", time.Now().UnixMilli()).(int64)
	after := parsedParams.Get("after", 0).(int64)
	name := parsedParams["name"].(string)
	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := parsedParams.Get("timezone", us.DefaultTimeZone).(string)

	err = db.RegisterUserEvent(ctx.Uid, db_types.UserViewingEvent{
		ItemId: id.ItemId,
		Timestamp: uint64(ts),
		After: uint64(after),
		Event: name,
		TimeZone: timezone,
	})
	if err != nil{
		util.WError(w, 500, "Could not register event\n%s", err.Error())
		return
	}
}

func ModUserEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	user := parsedParams["id"].(db_types.UserViewingEntry)

	user.Notes = parsedParams.Get("notes", user.Notes).(string)
	user.UserRating = parsedParams.Get("rating", user.UserRating).(float64)
	user.ViewCount = parsedParams.Get("view-count", user.ViewCount).(int64)
	user.CurrentPosition = parsedParams.Get("current-position", user.CurrentPosition).(string)
	user.Status = parsedParams.Get("status", user.Status).(db_types.Status)

	err := db.UpdateUserViewingEntry(ctx.Uid, &user)
	if err != nil {
		util.WError(w, 500, "Could not update user entry\n%s", err.Error())
		return
	}

	success(w)
}
