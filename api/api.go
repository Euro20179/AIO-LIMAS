package api

import (
	"fmt"
	"net/http"
	"strconv"

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

	price := req.URL.Query().Get("price")
	priceNum := 0.0
	if price != "" {
		var err error
		priceNum, err = strconv.ParseFloat(price, 64)
		if err != nil{
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
		if err != nil{
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
		if err != nil{
			w.WriteHeader(400)
			fmt.Fprintf(w, "Invalid parent id: %s\n" + err.Error(), i)
			return
		}
		p, err := db.GetItemById(int64(i))
		if err != nil{
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
