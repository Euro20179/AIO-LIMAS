package api

import (
	"net/http"

	db "aiolimas/db"
)

// lets the user add an item in their library
func AddEntry(w http.ResponseWriter, req *http.Request) {
	title := req.URL.Query().Get("title")
	if title == "" {
		w.WriteHeader(400)
		w.Write([]byte("No title provided\n"))
		return
	}

	var entryInfo db.InfoEntry
	entryInfo.Title = title

	var metadata db.MetadataEntry
	var userEntry db.UserViewingEntry

	if err := db.AddEntry(&entryInfo, &metadata, &userEntry); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into entryInfo table\n" + err.Error()))
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
		if err != nil{
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

// scans a folder and adds all items to the library as best it can
// it will not add it to any collections
func ScanFolder(w http.ResponseWriter, req *http.Request) {

}

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request) {
}
