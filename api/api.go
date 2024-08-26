package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	db "aiolimas/db"
	meta "aiolimas/metadata"
	"aiolimas/util"
)

func wError(w http.ResponseWriter, status int, format string, args ...any) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)
}

func ListCollections(w http.ResponseWriter, req *http.Request) {
	collections, err := db.ListCollections()
	if err != nil {
		wError(w, 500, "Could not get collections\n%s", err.Error())
		return
	}
	w.WriteHeader(200)
	for _, col := range collections {
		fmt.Fprintf(w, "%s\n", col)
	}
}

func ModEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}

	info, err := db.GetInfoEntryById(entry.ItemId)
	if err != nil {
		wError(w, 500, "%s\n", err.Error())
		return
	}

	query := req.URL.Query()

	title := query.Get("en-title")
	if title != "" {
		info.En_Title = title
	}

	nativeTitle := query.Get("native-title")
	if nativeTitle != "" {
		info.Native_Title = nativeTitle
	}

	format := query.Get("format")
	if format != "" {
		formatI, err := strconv.ParseInt(format, 10, 64)
		if err != nil {
			wError(w, 400, "Invalid format %s\n%s", format, err.Error())
			return
		}
		if !db.IsValidFormat(formatI) {
			wError(w, 400, "Invalid format %d\n", formatI)
			return
		}

		info.Format = db.Format(formatI)
	}

	parentId := query.Get("parent-id")
	if parentId != "" {
		parentIdI, err := strconv.ParseInt(parentId, 10, 64)
		if err != nil {
			wError(w, 400, "Invalid parent id %s\n%s", parentId, err.Error())
			return
		}

		if _, err := db.GetInfoEntryById(parentIdI); err != nil {
			wError(w, 400, "Non existant parent %d\n%s", parentIdI, err.Error())
			return
		}
		info.Parent = parentIdI
	}

	copyId := query.Get("copy-id")
	if copyId != "" {
		copyIdI, err := strconv.ParseInt(copyId, 10, 64)
		if err != nil {
			wError(w, 400, "Invalid copy id %s\n%s", copyId, err.Error())
			return
		}

		if _, err := db.GetInfoEntryById(copyIdI); err != nil {
			wError(w, 400, "Non existant item %d\n%s", copyIdI, err.Error())
			return
		}
		info.CopyOf = copyIdI
	}

	price := query.Get("price")
	if price != "" {
		priceF, err := strconv.ParseFloat(price, 64)
		if err != nil {
			wError(w, 400, "Invalid price %s\n%s", price, err.Error())
			return
		}
		info.PurchasePrice = priceF
	}

	location := query.Get("location")
	if location != "" {
		info.Location = location
	}

	tags := query.Get("tags")
	if tags != "" {
		info.Collection = tags
	}

	err = db.UpdateInfoEntry(&info)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	success(w)
}

// lets the user add an item in their library
/*
PARAMETERS:
title: string
price: float64
location: string
parentId: int64
format: Format
copyOf: int64
*/
func AddEntry(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	title := query.Get("title")
	if title == "" {
		w.WriteHeader(400)
		w.Write([]byte("No title provided\n"))
		return
	}

	price := query.Get("price")
	priceNum := 0.0
	if price != "" {
		var err error
		priceNum, err = strconv.ParseFloat(price, 64)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "%s is not an float\n", price)
			return
		}
	}

	format := query.Get("format")
	formatInt := int64(-1)
	if format != "" {
		var err error
		formatInt, err = strconv.ParseInt(format, 10, 64)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "%s is not an int\n", format)
			return
		}
		if !db.IsValidFormat(formatInt) {
			w.WriteHeader(400)
			fmt.Fprintf(w, "%d is not a valid format\n", formatInt)
			return
		}
	}

	parentQuery := query.Get("parentId")
	var parentId int64 = 0
	if parentQuery != "" {
		i, err := strconv.Atoi(parentQuery)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid parent id: %s\n"+err.Error(), i)
			return
		}
		p, err := db.GetInfoEntryById(int64(i))
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Parent does not exist: %d\n", i)
			return
		}
		parentId = p.ItemId
	}

	copyOf := query.Get("copyOf")
	var copyOfId int64 = 0
	if copyOf != "" {
		i, err := strconv.Atoi(copyOf)
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid copy id: %s\n"+err.Error(), i)
			return
		}
		p, err := db.GetInfoEntryById(int64(i))
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Item does not exist: %d\n", i)
			return
		}
		copyOfId = p.ItemId
	}

	ty := query.Get("type")

	isAnime := query.Get("is-anime")
	anime := isAnime == "true"

	var entryInfo db.InfoEntry
	entryInfo.En_Title = title
	entryInfo.PurchasePrice = priceNum
	entryInfo.Native_Title = query.Get("native-title")
	entryInfo.Collection = query.Get("tags")
	entryInfo.Location = query.Get("location")
	entryInfo.Format = db.Format(formatInt)
	entryInfo.Parent = parentId
	entryInfo.IsAnime = anime
	entryInfo.CopyOf = copyOfId
	if db.IsValidType(ty) {
		entryInfo.Type = db.MediaTypes(ty)
	} else {
		wError(w, 400, "%s is not a valid type\n", ty)
		return
	}

	var metadata db.MetadataEntry

	var userEntry db.UserViewingEntry

	userRating := query.Get("user-rating")
	if userRating != "" {
		ur, err := strconv.ParseFloat(userRating, 64)
		if err != nil {
			wError(w, 400, "%s is not a valid user rating\n%s", userRating, err.Error())
			return
		}
		userEntry.UserRating = ur
	}

	status := query.Get("user-status")
	if status != "" {
		if !db.IsValidStatus(status) {
			wError(w, 400, "%s is not a valid status\n", status)
			return
		}
		userEntry.Status = db.Status(status)
	}

	startDates := query.Get("user-start-dates")
	if startDates != "" {
		var startTimes []uint64
		err := json.Unmarshal([]byte(startDates), &startTimes)
		if err != nil {
			wError(w, 400, "Invalid start dates %s\n%s", startDates, err.Error())
			return
		}
	} else {
		startDates = "[]"
	}
	userEntry.StartDate = startDates

	endDates := query.Get("user-end-dates")
	if endDates != "" {
		var endTimes []uint64
		err := json.Unmarshal([]byte(endDates), &endTimes)
		if err != nil {
			wError(w, 400, "Invalid start dates %s\n%s", endDates, err.Error())
			return
		}
	} else {
		endDates = "[]"
	}
	userEntry.EndDate = endDates

	viewCount := query.Get("user-view-count")
	if viewCount != "" {
		vc, err := strconv.ParseInt(viewCount, 10, 64)
		if err != nil {
			wError(w, 400, "Invalid view count %s\n%s", viewCount, err.Error())
			return
		}
		userEntry.ViewCount = vc
	}

	userEntry.Notes = query.Get("user-notes")

	if query.Get("get-metadata") == "true" {
		providerOverride := query.Get("metadata-provider")
		if !meta.IsValidProvider(providerOverride) {
			providerOverride = ""
		}
		var err error
		metadata, err = meta.GetMetadata(&entryInfo, &metadata, providerOverride)
		if err != nil {
			wError(w, 500, "Could not get metadata\n%s", err.Error())
			return
		}
	}

	if err := db.AddEntry(&entryInfo, &metadata, &userEntry); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into table\n" + err.Error()))
		return
	}

	success(w)
}

// simply will list all entries as a json from the entryInfo table
func ListEntries(w http.ResponseWriter, req *http.Request) {
	items, err := db.Db.Query("SELECT * FROM entryInfo;")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	w.WriteHeader(200)
	for items.Next() {
		var row db.InfoEntry
		err = row.ReadEntry(items)
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

// should be an all in one query thingy, that should be able to query based on any matching column in any table
func QueryEntries(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	title := query.Get("title")
	nativeTitle := query.Get("native-title")
	location := query.Get("location")
	purchaseGt := query.Get("purchase-gt")
	purchaselt := query.Get("purchase-lt")
	formats := query.Get("formats")
	tags := query.Get("tags")
	types := query.Get("types")
	parents := query.Get("parent-ids")
	isAnime := query.Get("is-anime")
	copyIds := query.Get("copy-ids")

	pgt := 0.0
	plt := 0.0
	var fmts []db.Format
	var pars []int64
	var cos []int64
	var tys []db.MediaTypes
	tagsSplit := strings.Split(tags, ",")
	var collects []string
	for _, c := range tagsSplit {
		if c != "" {
			collects = append(collects, c)
		}
	}

	if util.IsNumeric([]byte(purchaseGt)) {
		pgt, _ = strconv.ParseFloat(purchaseGt, 64)
	}
	if util.IsNumeric([]byte(purchaselt)) {
		plt, _ = strconv.ParseFloat(purchaselt, 64)
	}

	for _, format := range strings.Split(formats, ",") {
		if util.IsNumeric([]byte(format)) {
			f, _ := strconv.ParseInt(format, 10, 64)
			if db.IsValidFormat(f) {
				fmts = append(fmts, db.Format(f))
			}
		}
	}

	for _, ty := range strings.Split(types, ",") {
		if db.IsValidType(ty) {
			tys = append(tys, db.MediaTypes(ty))
		}
	}

	for _, par := range strings.Split(parents, ",") {
		if util.IsNumeric([]byte(par)) {
			p, _ := strconv.ParseInt(par, 10, 64)
			pars = append(pars, p)
		}
	}

	for _, co := range strings.Split(copyIds, ",") {
		if util.IsNumeric([]byte(co)) {
			c, _ := strconv.ParseInt(co, 10, 64)
			cos = append(cos, c)
		}
	}

	var entrySearch db.EntryInfoSearch
	entrySearch.TitleSearch = title
	entrySearch.NativeTitleSearch = nativeTitle
	entrySearch.LocationSearch = location
	entrySearch.PurchasePriceGt = pgt
	entrySearch.PurchasePriceLt = plt
	entrySearch.InTags = collects
	entrySearch.Format = fmts
	entrySearch.Type = tys
	entrySearch.HasParent = pars
	entrySearch.CopyIds = cos
	switch isAnime {
	case "true":
		entrySearch.IsAnime = 2
	case "false":
		entrySearch.IsAnime = 1
	default:
		entrySearch.IsAnime = 0
	}

	rows, err := db.Search(entrySearch)
	if err != nil {
		wError(w, 500, "%s\n", err.Error())
		return
	}
	w.WriteHeader(200)
	for _, row := range rows {
		j, err := row.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

func GetCopies(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n")
		return
	}

	copies, err := db.GetCopiesOf(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not get copies of %d\n%s", entry.ItemId, err.Error())
		return
	}
	w.WriteHeader(200)
	for _, row := range copies {
		j, err := row.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

// Scans a folder as an entry, all folders within will be treated as children
// this will work well because even if a folder structure exists as:
// Friends -> S01 -> E01 -> 01.mkv
//
//	-> E02 -> 02.mkv
//
// We will end up with the following entries, Friends -> S01(Friends) -> E01(S01) -> 01.mkv(E01)
// despite what seems as duplication is actually fine, as the user may want some extra stuff associated with E01, if they structure it this way
// on rescan, we can check if the location doesn't exist, or is empty, if either is true, it will be deleted from the database
// **ONLY entryInfo rows will be deleted, as the user may have random userViewingEntries that are not part of their library**
// metadata also stays because it can be used to display the userViewingEntries nicer
// also on rescan, we can check if the title exists in entryInfo or metadata, if it does, we can reuse that id
func ScanFolder(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Query().Get("path")
	if path == "" {
		wError(w, 400, "No path given\n")
		return
	}

	collection := req.URL.Query().Get("collection-id")

	errs := db.ScanFolder(path, collection)

	if len(errs) != 0 {
		w.WriteHeader(500)
		for _, err := range errs {
			fmt.Fprintf(w, "%s\n", err.Error())
		}
		return
	}

	success(w)
}

func Stream(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n%s", err.Error())
		return
	}

	info, err := db.GetInfoEntryById(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not get info entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	http.ServeFile(w, req, info.Location)
}

func DeleteEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil {
		wError(w, 400, "Could not find entry\n%s", err.Error())
		return
	}
	err = db.Delete(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not delete entry\n%s", err.Error())
		return
	}
	success(w)
}

func GetDescendants(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil{
		wError(w, 400, "Could not find entry\n%s", err.Error())
		return
	}

	items, err := db.GetDescendants(entry.ItemId)
	if err != nil{
		wError(w, 500, "Could not get items\n%s", err.Error())
		return
	}
	w.WriteHeader(200)

	for _, item := range items {
		j, err := item.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
	w.Write([]byte("\n"))
}

func TotalCostOf(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil{
		wError(w, 400, "Could not find entry\n%s", err.Error())
		return
	}
	info, err := db.GetInfoEntryById(entry.ItemId)
	if err != nil{
		wError(w, 500, "Could not get price info\n%s", err.Error())
		return
	}
	desc, err := db.GetDescendants(entry.ItemId)
	if err != nil{
		wError(w, 500, "Could not get descendants\n%s", err.Error())
		return
	}

	cost := info.PurchasePrice
	for _, item := range desc {
		cost += item.PurchasePrice
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, "%f", cost)
}

func verifyIdQueryParam(req *http.Request) (int64, error) {
	id := req.URL.Query().Get("id")
	if id == "" {
		return 0, errors.New("No id given\n")
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is not an int\n", id)
	}
	return idInt, nil
}

func verifyIdAndGetUserEntry(w http.ResponseWriter, req *http.Request) (db.UserViewingEntry, error) {
	var out db.UserViewingEntry
	id, err := verifyIdQueryParam(req)
	if err != nil {
		return out, err
	}
	entry, err := db.GetUserViewEntryById(id)
	if err != nil {
		wError(w, 400, "There is no entry with id %d\n", id)
		return out, err
	}

	return entry, nil
}

func success(w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}
