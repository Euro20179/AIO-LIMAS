package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	db "aiolimas/db"
	meta "aiolimas/metadata"
	"aiolimas/settings"
	"aiolimas/types"
	"aiolimas/util"
)

func ListCollections(ctx RequestContext) {
	w := ctx.W
	collections, err := db.ListType(ctx.Uid, "en_title", db_types.TY_COLLECTION)
	if err != nil {
		util.WError(w, 500, "Could not get collections\n%s", err.Error())
		return
	}
	w.WriteHeader(200)
	for _, col := range collections {
		fmt.Fprintf(w, "%s\n", col)
	}
}

func ListLibraries(ctx RequestContext) {
	w := ctx.W
	libraries, err := db.ListType(ctx.Uid, "itemId", db_types.TY_LIBRARY)
	if err != nil {
		util.WError(w, 500, "Could not get collections\n%s", err.Error())
		return
	}
	ctx.W.WriteHeader(200)
	for _, col := range libraries {
		fmt.Fprintf(w, "%s\n", col)
	}
}

func AddTags(ctx RequestContext) {
	entry := ctx.PP["id"].(db_types.InfoEntry)
	newTags := ctx.PP["tags"].([]string)
	uid := ctx.Uid

	if err := db.AddTags(uid, entry.ItemId, newTags); err != nil {
		ctx.W.WriteHeader(500)
		ctx.W.Write([]byte("Could not add tags"))
		return
	}

	success(ctx.W)
}

func DeleteTags(ctx RequestContext) {
	entry := ctx.PP["id"].(db_types.InfoEntry)
	newTags := ctx.PP["tags"].([]string)
	uid := ctx.Uid

	if err := db.DelTags(uid, entry.ItemId, newTags); err != nil {
		ctx.W.WriteHeader(500)
		ctx.W.Write([]byte("Could not add tags"))
		return
	}

	success(ctx.W)
}

func DownloadDB(ctx RequestContext) {
	dbPath := fmt.Sprintf("%s/all.db", db.UserRoot(ctx.Uid))

	http.ServeFile(ctx.W, ctx.Req, dbPath)
}

func GetAllForEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W

	info := parsedParams["id"].(db_types.InfoEntry)

	events, err := db.GetEvents(ctx.Uid, info.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get events\n%s", err.Error())
		return
	}

	user, err := db.GetUserViewEntryById(ctx.Uid, info.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get user info\n%s", err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(ctx.Uid, info.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get metadata info\n%s", err.Error())
		return
	}

	uj, err := user.ToJson()
	if err != nil {
		util.WError(w, 500, "Could not marshal user info\n%s", err.Error())
		return
	}

	mj, err := meta.ToJson()
	if err != nil {
		util.WError(w, 500, "Could not marshal metadata info\n%s", err.Error())
		return
	}

	ij, err := info.ToJson()
	if err != nil {
		util.WError(w, 500, "Could not marshal main entry info\n%s", err.Error())
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

func SetEntry(ctx RequestContext) {
	w := ctx.W
	req := ctx.Req

	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		util.WError(w, 500, "Could not ready body\n%s", err.Error())
		return
	}

	var entry db_types.InfoEntry
	err = json.Unmarshal(data, &entry)
	if err != nil {
		util.WError(w, 400, "Could not parse json into entry\n%s", err.Error())
		return
	}

	err = db.UpdateInfoEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(w, 500, "Could not update info entry\n%s", err.Error())
		return
	}
	success(w)
}

func ModEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
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

	err := db.UpdateInfoEntry(ctx.Uid, &info)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	success(w)
}

// lets the user add an item in their library
func AddEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
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

	var libraryId int64 = 0
	if l, exists := parsedParams["libraryId"]; exists {
		libraryId = l.(db_types.InfoEntry).ItemId
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
	entryInfo.Library = libraryId

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
		newMeta, err := meta.GetMetadata(&meta.GetMetadataInfo{
			Entry:         &entryInfo,
			MetadataEntry: &metadata,
			Override:      providerOverride,
			Uid:           ctx.Uid,
		})
		if err != nil {
			util.WError(w, 500, "Could not get metadata\n%s", err.Error())
			return
		}

		newMeta.ItemId = entryInfo.ItemId
		metadata = newMeta
	}


	us, err := settings.GetUserSettigns(ctx.Uid)
	if err != nil{
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := parsedParams.Get("timezone", us.DefaultTimeZone).(string)

	if err := db.AddEntry(ctx.Uid, timezone, &entryInfo, &metadata, &userEntry); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into table\n" + err.Error()))
		return
	}

	j, err := entryInfo.ToJson()
	if err != nil {
		util.WError(w, 500, "Could not convert new entry to json\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(j)
}

// simply will list all entries as a json from the entryInfo table
func ListEntries(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	sortBy, _ := parsedParams.Get("sort-by", "userRating").(string)
	entries, err := db.ListEntries(ctx.Uid, sortBy)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Could not query entries\n" + err.Error()))
		return
	}

	w.WriteHeader(200)
	for _, row := range entries {
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

func QueryEntries3(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	search := pp["search"].(string)

	results, err := db.Search3(ctx.Uid, search)
	if err != nil {
		util.WError(w, 500, "Could not complete search\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	writeSQLRowResults(w, results)
}

func GetCopies(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.InfoEntry)

	copies, err := db.GetCopiesOf(ctx.Uid, entry.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get copies of %d\n%s", entry.ItemId, err.Error())
		return
	}
	w.WriteHeader(200)
	writeSQLRowResults(w, copies)
}

func Stream(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	req := ctx.Req
	entry := parsedParams["id"].(db_types.InfoEntry)
	subFile := parsedParams.Get("subfile", "").(string)

	subFile, err := url.QueryUnescape(subFile)
	if err != nil {
		subFile = ""
	}

	newLocation := os.ExpandEnv(entry.Location)

	fullPath := newLocation

	if subFile != "" {
		fullPath += "/" + subFile
	}

	stat, err := os.Stat(fullPath)
	if err == nil && stat.IsDir() {
		files, err := os.ReadDir(fullPath)
		if err != nil {
			return
		}
		w.Write([]byte("#EXTM3U\n"))
		for _, file := range files {
			path := url.QueryEscape(file.Name())
			var data string
			if subFile != "" {
				data = fmt.Sprintf("stream-entry?id=%d&subfile=%s\n", entry.ItemId, subFile+"/"+path)
			} else {
				data = fmt.Sprintf("stream-entry?id=%d&subfile=%s\n", entry.ItemId, path)
			}
			w.Write([]byte(data))
		}
	} else if err != nil {
		println(err.Error())
		w.WriteHeader(500)
		w.Write([]byte("ERROR"))
	} else {
		http.ServeFile(w, req, fullPath)
	}
}

func DeleteEntry(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.InfoEntry)
	err := db.Delete(ctx.Uid, entry.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not delete entry\n%s", err.Error())
		return
	}
	success(w)
}

func GetDescendants(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.InfoEntry)

	items, err := db.GetDescendants(ctx.Uid, entry.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get items\n%s", err.Error())
		return
	}
	w.WriteHeader(200)

	writeSQLRowResults(w, items)
	w.Write([]byte("\n"))
}

func GetTree(ctx RequestContext) {
	w := ctx.W
	tree, err := db.BuildEntryTree(ctx.Uid)
	if err != nil {
		util.WError(w, 500, "Could not build tree\n%s", err.Error())
		return
	}
	jStr, err := json.Marshal(tree)
	if err != nil {
		util.WError(w, 500, "Could not marshal tree\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(jStr)
}

// TODO: allow this to accept multiple ids
func TotalCostOf(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	info := pp["id"].(db_types.InfoEntry)
	desc, err := db.GetDescendants(ctx.Uid, info.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get descendants\n%s", err.Error())
		return
	}

	cost := info.PurchasePrice
	for _, item := range desc {
		cost += item.PurchasePrice
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, "%f", cost)
}

func success(w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}
