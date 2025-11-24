package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	db "aiolimas/db"
	"aiolimas/logging"
	meta "aiolimas/metadata"
	"aiolimas/settings"
	"aiolimas/types"
	"aiolimas/util"
)

func ListRelations(ctx RequestContext) {
	relations, err := db.ListRelations(ctx.Uid)

	if err != nil{
		util.WError(ctx.W, 500, "Could not list relations\n%s", err.Error())
		return
	}

	data, err := json.Marshal(relations)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to serialize relations\n%s", err.Error())
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write(data)
}

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
	dbPath := fmt.Sprintf("%s/all.db", db.DbRoot())

	http.ServeFile(ctx.W, ctx.Req, dbPath)
}

func _getAllForEntry(w http.ResponseWriter, uid int64, info db_types.InfoEntry) {
	events, err := db.GetEvents(uid, info.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get events\n%s", err.Error())
		return
	}

	user, err := db.GetUserViewEntryById(uid, info.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get user info\n%s", err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(uid, info.ItemId)
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

	w.Write(uj)
	w.Write([]byte("\n"))

	w.Write(mj)
	w.Write([]byte("\n"))

	w.Write(ij)
	w.Write([]byte("\n"))

	writeSQLRowResults(w, events)
}

func GetAllForEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W

	info := parsedParams["id"].(db_types.InfoEntry)
	w.WriteHeader(200)
	_getAllForEntry(ctx.W, ctx.Uid, info)
}

func GetAllForEntries(ctx RequestContext) {
	ids := ctx.PP["ids"].([]string)

	if ctx.Uid == -1 {
		ctx.Uid = 0
	}

	for _, id := range ids {
		n, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			util.WError(ctx.W, 400, "Invalid id: '%s'", id)
			break
		}

		i, err := db.GetInfoEntryById(ctx.Uid, n)
		if err != nil {
			util.WError(ctx.W, 500, "An error occured while accessing id: '%s': %s", id, err)
			break
		}

		ctx.W.WriteHeader(200)
		_getAllForEntry(ctx.W, ctx.Uid, i)
		ctx.W.Write([]byte("\n"))
	}
}

func GetRecommenders(ctx RequestContext) {
	uid := ctx.Uid

	r, err := db.GetRecommendersList(uid)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get a list of recommenders\n%s", err)
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write([]byte(strings.Join(r, "\x1F")))
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
		if err := db.SetParent(ctx.Uid, info.ItemId, parent.ItemId); err != nil {
			logging.ELog(err)
			util.WError(ctx.W, 500, "Failed to set parent\n%s", err.Error())
		}
	}

	if orphan, exists := parsedParams["become-orphan"].(bool); exists && orphan {
		if err := db.BecomeOrphan(ctx.Uid, info.ItemId); err != nil {
			logging.ELog(err)
			util.WError(ctx.W, 500, "Failed to make orphan\n%s", err.Error())
		}
	}

	if original, exists := parsedParams["become-original"].(bool); exists && original {
		if err := db.BecomeOriginal(ctx.Uid, info.ItemId); err != nil {
			logging.ELog(err)
			util.WError(ctx.W, 500, "Failed to make orignal\n%s", err.Error())
		}
	}

	if itemCopy, exists := parsedParams["copy-id"].(db_types.InfoEntry); exists {
		if err := db.SetCopy(ctx.Uid, info.ItemId, itemCopy.ItemId); err != nil {
			logging.ELog(err)
		}
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

func DelChild(ctx RequestContext) {
	uid := ctx.Uid
	parent := ctx.PP["parent"].(db_types.InfoEntry)
	child := ctx.PP["child"].(db_types.InfoEntry)

	err := db.DelRelation(uid, child.ItemId, db_types.R_Child, parent.ItemId)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to delete child\n%s", err.Error())
		return
	}

	success(ctx.W)
}

func DelCopy(ctx RequestContext) {
	uid := ctx.Uid
	cpy := ctx.PP["copy"].(db_types.InfoEntry)
	cpyOf := ctx.PP["copyof"].(db_types.InfoEntry)

	err := db.DelRelation(uid, cpy.ItemId, db_types.R_Copy, cpyOf.ItemId)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to delete copy\n%s", err.Error())
		return
	}

	success(ctx.W)
}

func DelRequires(ctx RequestContext) {
	uid := ctx.Uid
	item := ctx.PP["itemid"].(db_types.InfoEntry)
	requires := ctx.PP["requires"].(db_types.InfoEntry)

	err := db.DelRelation(uid, item.ItemId, db_types.R_Requires, requires.ItemId)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to delete requirement\n%s", err.Error())
		return
	}

	success(ctx.W)
}

func AddChild(ctx RequestContext) {
	uid := ctx.Uid
	parent := ctx.PP["parent"].(db_types.InfoEntry)
	child := ctx.PP["child"].(db_types.InfoEntry)

	err := db.AddRelation(uid, child.ItemId, db_types.R_Child, parent.ItemId)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to add child\n%s", err.Error())
		return
	}

	success(ctx.W)
}

func AddCopy(ctx RequestContext) {
	uid := ctx.Uid
	cpy := ctx.PP["copy"].(db_types.InfoEntry)
	cpyOf := ctx.PP["copyof"].(db_types.InfoEntry)

	err := db.AddRelation(uid, cpy.ItemId, db_types.R_Copy, cpyOf.ItemId)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to add copy\n%s", err.Error())
		return
	}

	success(ctx.W)
}

func AddRequires(ctx RequestContext) {
	uid := ctx.Uid
	item := ctx.PP["itemid"].(db_types.InfoEntry)
	requires := ctx.PP["requires"].(db_types.InfoEntry)

	err := db.AddRelation(uid, item.ItemId, db_types.R_Requires, requires.ItemId)

	if err != nil{
		util.WError(ctx.W, 500, "Failed to add requirement\n%s", err.Error())
		return
	}

	success(ctx.W)
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

	var requiresId int64 = 0
	if r, exists := parsedParams["requires"]; exists {
		requiresId = r.(db_types.InfoEntry).ItemId
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
	entryInfo.ItemId = 0
	entryInfo.En_Title = title
	entryInfo.PurchasePrice = priceNum
	entryInfo.Native_Title = nativeTitle
	entryInfo.Location = location
	entryInfo.Format = db_types.Format(formatInt)
	entryInfo.ArtStyle = db_types.ArtStyle(style)
	entryInfo.Type = parsedParams["type"].(db_types.MediaTypes)
	entryInfo.Library = libraryId
	entryInfo.Requires = requiresId
	entryInfo.RecommendedBy = parsedParams.Get("recommended-by", "").(string)

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

	if copyOfId != 0 {
		db.AddRelation(userEntry.Uid, entryInfo.ItemId, db_types.R_Copy, copyOfId)
	}

	if parentId != 0 {
		db.AddRelation(userEntry.Uid, entryInfo.ItemId, db_types.R_Child, parentId)
	}

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
		newMeta.Uid = ctx.Uid
		metadata = newMeta
	}

	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil {
		util.WError(w, 500, "Could not update entry\n%s", err.Error())
		return
	}
	timezone := parsedParams.Get("timezone", us.DefaultTimeZone).(string)

	if err := db.AddEntry(ctx.Uid, timezone, &entryInfo, &metadata, &userEntry); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Error adding into table\n" + err.Error()))
		return
	}

	if tags != "" {
		if err := db.AddTags(ctx.Uid, entryInfo.ItemId, strings.Split(tags, ",")); err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error adding tags table\n" + err.Error()))
			return
		}
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
			logging.ELog(err)
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

	// can be -1 if user does not provide uid
	if ctx.Uid > 0 {
		search += fmt.Sprintf(" & {entryInfo.uid = %d}", ctx.Uid)
	}

	results, err := db.Search3(search, pp.Get("order-by", "").(string))
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

	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil {
		logging.ELog(err)
		return
	}

	fullPath := settings.ExpandPathWithLocationAliases(us.LocationAliases, entry.Location)

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
		logging.ELog(err)
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
		logging.ELog(err)
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
