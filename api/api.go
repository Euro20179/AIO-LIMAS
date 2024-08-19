package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	db "aiolimas/db"
)

func wError(w http.ResponseWriter, status int, format string, args ...any) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)
}

// lets the user add an item in their library
func AddEntry(w http.ResponseWriter, req *http.Request) {
	title := req.URL.Query().Get("title")

	if title == "" {
		w.WriteHeader(400)
		w.Write([]byte("No title provided\n"))
		return
	}

	price := req.URL.Query().Get("price")
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

	format := req.URL.Query().Get("format")
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

	parentQuery := req.URL.Query().Get("parentId")
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

	var entryInfo db.InfoEntry
	entryInfo.Title = title
	entryInfo.PurchasePrice = priceNum
	entryInfo.Location = req.URL.Query().Get("location")
	entryInfo.Format = db.Format(formatInt)
	entryInfo.Parent = parentId

	var metadata db.MetadataEntry
	var userEntry db.UserViewingEntry

	if err := db.AddEntry(&entryInfo, &metadata, &userEntry); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into table\n" + err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
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
		items.Scan(&row.ItemId, &row.Title, &row.Format, &row.Location, &row.PurchasePrice, &row.Collection, &row.Parent)
		j, err := row.ToJson()
		if err != nil {
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
	w.Write([]byte("\n"))
}

// should be an all in one query thingy, that should be able to query based on any matching column in any table
func QueryEntries(w http.ResponseWriter, req *http.Request) {
}

// Scans a folder, and makes each item in it part of a collection named the same as the folder
// all items in it will be passed to the ScanFolderAsEntry db function
func ScanFolderAsCollection(w http.ResponseWriter, req *http.Request) {
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
func ScanFolderAsEntry(w http.ResponseWriter, req *http.Request) {
}

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(400)
		w.Write([]byte("No id given\n"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "%s is not an int", id)
		return
	}

	entry, err := db.GetUserViewEntryById(int64(idInt))
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "There is no entry with id %s\n", id)
		return
	}

	if entry.CanBegin() {
		w.WriteHeader(405)
		fmt.Fprintf(w, "This media is already being viewed, cannot start again\n")
		return
	}

	var startTimes []uint64
	err = json.Unmarshal([]byte(entry.StartDate), &startTimes)
	if err != nil {
		wError(w, 500, "Could not decode start times into int[], %d may be corrupted\n", idInt)
		return
	}

	if err := entry.Begin(); err != nil {
		wError(w, 500, "Could not begin show\n%s", err.Error())
		return
	}
}



