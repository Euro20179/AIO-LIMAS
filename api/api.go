package api

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"

	globals "aiolimas/globals"
)

type EntryInfo struct {
	ItemId        int64
	Title         string
	Format        string
	Location      string
	PurchasePrice float64
	Collection    string
	Parent        int64
}

func (self *EntryInfo) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

// lets the user add an item in their library
func AddEntry(w http.ResponseWriter, req *http.Request) {
	title := req.URL.Query().Get("title")
	if title == "" {
		w.WriteHeader(400)
		w.Write([]byte("No title provided\n"))
		return
	}

	id := rand.Int64()
	if id == 0 {
		w.WriteHeader(500)
		w.Write([]byte("Failed to generate an id"))
		return
	}

	query := fmt.Sprintf("INSERT INTO entryInfo (itemId, title, format, location, purchasePrice, collection) VALUES (%d, '%s', 'digital', 'test', 0, 'test')", id, title)
	_, err := globals.Db.Exec(query)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into entryInfo table\n" + err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}

// simply will list all entries as a json from the entryInfo table
func ListEntries(w http.ResponseWriter, req *http.Request) {
	items, err := globals.Db.Query("SELECT * FROM entryInfo;")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	w.WriteHeader(200)
	for items.Next() {
		var row EntryInfo
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
func ScanFolder() {}

// engagement endpoints
func BeginMedia(w http.ResponseWriter, req *http.Request) {
}
