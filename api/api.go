package api;
import (
	"net/http"
	"math/rand/v2"
	"fmt"

	globals "aiolimas/globals"
)

// lets the user add an item in their library
func AddEntry(w http.ResponseWriter, req *http.Request) {
	title := req.URL.Query().Get("title")
	if title == "" {
		w.WriteHeader(400)
		w.Write([]byte("No title provided\n"))
		return
	}
	id := rand.Uint64()

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

//should be an all in one query thingy, that should be able to query based on any matching column in any table
func QueryEntries(w http.ResponseWriter, req *http.Request) {

}

// scans a folder and adds all items to the library as best it can
// it will not add it to any collections
func ScanFolder() {}
