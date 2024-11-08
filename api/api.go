package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	db "aiolimas/db"
	"aiolimas/types"
	meta "aiolimas/metadata"
)

func wError(w http.ResponseWriter, status int, format string, args ...any) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)
}

func ListCollections(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
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

func DownloadDB(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	dir := os.Getenv("AIO_DB_PATH")
	if dir == "" {
		panic("$AIO_DB_PATH should not be empty")
	}

	http.ServeFile(w, req, dir)
}

func GetAllForEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	info := parsedParams["id"].(db_types.InfoEntry)

	events, err := db.GetEvents(info.ItemId)
	if err != nil {
		wError(w, 500, "Could not get events\n%s", err.Error())
		return
	}

	user, err := db.GetUserViewEntryById(info.ItemId)
	if err != nil {
		wError(w, 500, "Could not get user info\n%s", err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(info.ItemId)
	if err != nil {
		wError(w, 500, "Could not get metadata info\n%s", err.Error())
		return
	}

	uj, err := user.ToJson()
	if err != nil {
		wError(w, 500, "Could not marshal user info\n%s", err.Error())
		return
	}

	mj, err := meta.ToJson()
	if err != nil {
		wError(w, 500, "Could not marshal metadata info\n%s", err.Error())
		return
	}

	ij, err := info.ToJson()
	if err != nil {
		wError(w, 500, "Could not marshal main entry info\n%s", err.Error())
		return
	}

	w.WriteHeader(200)

	w.Write(uj)
	w.Write([]byte("\n"))

	w.Write(mj)
	w.Write([]byte("\n"))

	w.Write(ij)
	w.Write([]byte("\n"))

	writeSQLRowResults(w, events)
}

func SetEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		wError(w, 500, "Could not ready body\n%s", err.Error())
		return
	}

	var entry db_types.InfoEntry
	err = json.Unmarshal(data, &entry)
	if err != nil {
		wError(w, 400, "Could not parse json into entry\n%s", err.Error())
		return
	}

	err = db.UpdateInfoEntry(&entry)
	if err != nil {
		wError(w, 500, "Could not update info entry\n%s", err.Error())
		return
	}
	success(w)
}

func ModEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	info := parsedParams["id"].(db_types.InfoEntry)

	title, exists := parsedParams["en-title"].(string)
	if exists {
		info.En_Title = title
	}

	nativeTitle, exists := parsedParams["native-title"].(string)
	if exists {
		info.Native_Title = nativeTitle
	}

	format, exists := parsedParams["format"].(db_types.Format)
	if exists {
		info.Format = format
	}

	parent, exists := parsedParams["parent-id"].(db_types.InfoEntry)
	if exists {
		info.ParentId = parent.ItemId
	}

	if orphan, exists := parsedParams["become-orphan"].(bool); exists && orphan {
		info.ParentId = 0
	}

	if original, exists := parsedParams["become-original"].(bool); exists && original {
		info.CopyOf = 0
	}

	if itemCopy, exists := parsedParams["copy-id"].(db_types.InfoEntry); exists {
		info.CopyOf = itemCopy.ItemId
	}

	if price, exists := parsedParams["price"].(float64); exists {
		info.PurchasePrice = price
	}

	if location, exists := parsedParams["location"].(string); exists {
		info.Location = location
	}

	if tags, exists := parsedParams["tags"].(string); exists {
		info.Collection = tags
	}

	info.ArtStyle = db_types.ArtStyle(parsedParams.Get("art-style", uint(0)).(uint))
	info.Type = parsedParams.Get("type", info.Type).(db_types.MediaTypes)

	err := db.UpdateInfoEntry(&info)
	if err != nil {
		wError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	success(w)
}

// lets the user add an item in their library
func AddEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	title := parsedParams["title"].(string)

	priceNum := parsedParams.Get("price", 0.0).(float64)

	formatInt := parsedParams["format"].(db_types.Format)

	if digital, exists := parsedParams["is-digital"]; exists {
		if digital.(bool) {
			formatInt |= db_types.F_MOD_DIGITAL
		}
	}

	var parentId int64 = 0
	if parent, exists := parsedParams["parentId"]; exists {
		parentId = parent.(db_types.InfoEntry).ItemId
	}

	var copyOfId int64 = 0

	if c, exists := parsedParams["copyOf"]; exists {
		copyOfId = c.(db_types.InfoEntry).ItemId
	}

	style := parsedParams.Get("art-style", uint(0)).(uint)

	if parsedParams.Get("is-anime", false).(bool) {
		style &= uint(db_types.AS_ANIME)
	}

	nativeTitle := ""
	if title, exists := parsedParams["native-title"]; exists {
		nativeTitle = title.(string)
	}

	location := ""
	if l, exists := parsedParams["location"]; exists {
		location = l.(string)
	}

	tags := ""
	if t, exists := parsedParams["tags"]; exists {
		tags = t.(string)
	}

	var entryInfo db_types.InfoEntry
	entryInfo.En_Title = title
	entryInfo.PurchasePrice = priceNum
	entryInfo.Native_Title = nativeTitle
	entryInfo.Collection = tags
	entryInfo.Location = location
	entryInfo.Format = db_types.Format(formatInt)
	entryInfo.ParentId = parentId
	entryInfo.ArtStyle = db_types.ArtStyle(style)
	entryInfo.CopyOf = copyOfId
	entryInfo.Type = parsedParams["type"].(db_types.MediaTypes)

	var metadata db_types.MetadataEntry

	var userEntry db_types.UserViewingEntry

	if userRating, exists := parsedParams["user-rating"]; exists {
		userEntry.UserRating = userRating.(float64)
	}
	if status, exists := parsedParams["user-status"]; exists {
		userEntry.Status = status.(db_types.Status)
	}

	userEntry.ViewCount = parsedParams.Get("user-view-count", int64(0)).(int64)

	userEntry.Notes = parsedParams.Get("user-notes", "").(string)

	if parsedParams.Get("get-metadata", false).(bool) {
		providerOverride := parsedParams.Get("metadata-provider", "").(string)
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

	j, err := entryInfo.ToJson()
	if err != nil {
		wError(w, 500, "Could not convert new entry to json\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(j)
}

// simply will list all entries as a json from the entryInfo table
func ListEntries(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	sortBy, _ := parsedParams.Get("sort-by", "userRating").(string)
	items, err := db.Db.Query(fmt.Sprintf(`
		SELECT entryInfo.*
		FROM
			entryInfo JOIN userViewingInfo
		ON
			entryInfo.itemId = userViewingInfo.itemId
		ORDER BY %s`, sortBy))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}
	w.WriteHeader(200)
	for items.Next() {
		var row db_types.InfoEntry
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

func QueryEntries3(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	search := pp["search"].(string)

	results, err := db.Search3(search)
	if err != nil {
		wError(w, 500, "Could not complete search\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	writeSQLRowResults(w, results)
}

func GetCopies(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.InfoEntry)

	copies, err := db.GetCopiesOf(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not get copies of %d\n%s", entry.ItemId, err.Error())
		return
	}
	w.WriteHeader(200)
	writeSQLRowResults(w, copies)
}

func Stream(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	entry := parsedParams["id"].(db_types.InfoEntry)

	http.ServeFile(w, req, entry.Location)
}

func DeleteEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.InfoEntry)
	err := db.Delete(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not delete entry\n%s", err.Error())
		return
	}
	success(w)
}

func GetDescendants(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db_types.InfoEntry)

	items, err := db.GetDescendants(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not get items\n%s", err.Error())
		return
	}
	w.WriteHeader(200)

	writeSQLRowResults(w, items)
	w.Write([]byte("\n"))
}

func GetTree(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	tree, err := db.BuildEntryTree()
	if err != nil {
		wError(w, 500, "Could not build tree\n%s", err.Error())
		return
	}
	jStr, err := json.Marshal(tree)
	if err != nil {
		wError(w, 500, "Could not marshal tree\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(jStr)
}

// TODO: allow this to accept multiple ids
func TotalCostOf(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	info := pp["id"].(db_types.InfoEntry)
	desc, err := db.GetDescendants(info.ItemId)
	if err != nil {
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

func verifyIdAndGetUserEntry(w http.ResponseWriter, req *http.Request) (db_types.UserViewingEntry, error) {
	var out db_types.UserViewingEntry
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
