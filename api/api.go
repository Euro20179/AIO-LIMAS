package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
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
	if err != nil {
		util.WError(ctx.W, 500, "Could not list relations\n%s", err.Error())
		return
	}

	data, err := json.Marshal(relations)
	if err != nil {
		util.WError(ctx.W, 500, "Failed to serialize relations\n%s", err.Error())
		return
	}

	ctx.W.WriteHeader(200)
	ctx.W.Write(data)
}

func ListRelationsAsJSONL(ctx RequestContext) {
	relations, err := db.ListRelations(ctx.Uid)
	if err != nil {
		util.WError(ctx.W, 500, "Could not list relations\n%s", err.Error())
		return
	}
	for id, rs := range relations {
		out := struct{
			ItemId int64
			Children []int64
			Requires []int64
			Copies []int64
		} {
			ItemId: id,
			Children: rs.Children,
			Copies: rs.Copies,
			Requires: rs.Requires,
		}

		if out.Children == nil {
			out.Children = []int64{}
		}
		if out.Requires == nil {
			out.Requires = []int64{}
		}
		if out.Copies == nil {
			out.Copies = []int64{}
		}

		res, err := json.Marshal(out)
		if err != nil {
			util.WError(ctx.W, 500, "Failed to serialize relations\n%s", err.Error())
			return
		}
		ctx.W.Write(res)
		ctx.W.Write([]byte("\n"))
	}
}

func ListCollections(ctx RequestContext) {
	w := ctx.W
	collections, err := db.ListType(actx2dctx(ctx), "en_title", db_types.TY_COLLECTION)
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
	libraries, err := db.ListType(actx2dctx(ctx), "itemId", db_types.TY_LIBRARY)
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

func _getAllForEntry(ctx RequestContext, info db_types.InfoEntry) {
	events, err := db.GetEvents(actx2dctx(ctx), info.ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get events\n%s", err.Error())
		return
	}

	user, err := db.GetUserViewEntryById(actx2dctx(ctx), info.ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get user info\n%s", err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(actx2dctx(ctx), info.ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get metadata info\n%s", err.Error())
		return
	}

	trans, err := db.ListTransactions(actx2dctx(ctx), info.ItemId)

	uj, err := user.ToJson()
	if err != nil {
		util.WError(ctx.W, 500, "Could not marshal user info\n%s", err.Error())
		return
	}

	mj, err := meta.ToJson()
	if err != nil {
		util.WError(ctx.W, 500, "Could not marshal metadata info\n%s", err.Error())
		return
	}

	ij, err := info.ToJson()
	if err != nil {
		util.WError(ctx.W, 500, "Could not marshal main entry info\n%s", err.Error())
		return
	}

	ctx.W.Write(uj)
	ctx.W.Write([]byte("\n"))

	ctx.W.Write(mj)
	ctx.W.Write([]byte("\n"))

	ctx.W.Write(ij)
	ctx.W.Write([]byte("\n"))

	writeSQLRowResults(ctx.W, events)
	ctx.W.Write([]byte("TRANSACTIONS\n"))
	writeSQLRowResults(ctx.W, trans)
}

func GetAllForEntry2(ctx RequestContext) {
	info := ctx.PP["id"].(db_types.InfoEntry)
	events, err := db.GetEvents(actx2dctx(ctx), info.ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get events\n%s", err.Error())
		return
	}

	user, err := db.GetUserViewEntryById(actx2dctx(ctx), info.ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get user info\n%s", err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(actx2dctx(ctx), info.ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get metadata info\n%s", err.Error())
		return
	}

	trans, err := db.ListTransactions(actx2dctx(ctx), info.ItemId)

	children, err := db.GetRelation(actx2dctx(ctx), info.ItemId, db_types.R_Child)
	copies, err := db.GetRelation(actx2dctx(ctx), info.ItemId, db_types.R_Copy)
	requirements, err := db.GetRequires(actx2dctx(ctx), info.ItemId)

	if events == nil {
		events = []db_types.UserViewingEvent{}
	}

	final := struct {
		Info         db_types.InfoEntry
		User         db_types.UserViewingEntry
		Meta         db_types.MetadataEntry
		Events       []db_types.UserViewingEvent
		Transactions []db_types.TransactionEntry
		Children     []uint64
		Copies       []uint64
		Requirements []uint64
	}{
		Info:         info,
		Meta:         meta,
		User:         user,
		Events:       events,
		Transactions: trans,
		Copies:       []uint64{},
		Children:     []uint64{},
		Requirements: []uint64{},
	}

	for _, child := range children {
		final.Children = append(final.Children, uint64(child.ItemId))
	}

	for _, copy := range copies {
		final.Copies = append(final.Copies, uint64(copy.ItemId))
	}

	for _, requirement := range requirements {
		final.Requirements = append(final.Requirements, uint64(requirement.ItemId))
	}

	out, err := json.Marshal(final)
	if err != nil {
		util.WError(ctx.W, 500, "Failed to marshal response: %s\n", err.Error())
		return
	}
	ctx.W.WriteHeader(200)
	ctx.W.Write(out)
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

		i, err := db.GetInfoEntryById(actx2dctx(ctx), n)
		if err != nil {
			util.WError(ctx.W, 500, "An error occured while accessing id: '%s': %s", id, err.Error())
			break
		}

		ctx.W.WriteHeader(200)
		_getAllForEntry(ctx, i)
		ctx.W.Write([]byte("\n"))
	}
}

func GetRecommenders(ctx RequestContext) {
	r, err := db.GetRecommendersList(actx2dctx(ctx))
	if err != nil {
		util.WError(ctx.W, 500, "Could not get a list of recommenders\n%s", err.Error())
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

	entry := ctx.PP["id"].(db_types.InfoEntry)
	oldId := entry.ItemId

	err = json.Unmarshal(data, &entry)
	if err != nil {
		util.WError(w, 400, "Could not parse json into entry\n%s", err.Error())
		return
	}

	entry.Uid = ctx.Uid
	entry.ItemId = oldId

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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
		util.WError(ctx.W, 500, "Failed to add requirement\n%s", err.Error())
		return
	}

	success(ctx.W)
}

type RadarrPostWebhook struct {
	Movie struct {
		Id          int
		Title       string
		Year        int
		ReleaseDate string
		FolderPath  string
		TmdbId      int
		Tags        []string
	}
	RemoteMovie struct {
		TmdbId int
		ImdbId string
		Title  string
		year   int
	}
	Release struct {
		Quality           string
		QualityVersion    int
		ReleaseGroup      string
		ReleaseTitle      string
		Indexer           string
		Size              int
		CustomFormatScore int
	}
	EventType      string
	InstanceName   string
	ApplicationUrl string
}

func _radarrAdd(ctx RequestContext, data RadarrPostWebhook) {
	var entryInfo db_types.InfoEntry
	var userEntry db_types.UserViewingEntry

	if slices.Contains(data.Movie.Tags, "no-add") {
		ctx.W.WriteHeader(200)
		return
	}

	us, err := settings.GetUserSettings(ctx.Uid)
	if err != nil {
		util.WError(ctx.W, 500, "Could not update entry\n%s", err.Error())
		return
	}

	if slices.Contains(data.Movie.Tags, "planned") {
		userEntry.Status = db_types.S_PLANNED
	}

	entryInfo.En_Title = data.Movie.Title
	entryInfo.Type = db_types.TY_MOVIE
	entryInfo.Format = db_types.F_DIGITAL
	entryInfo.ItemId = 0
	entryInfo.Location = settings.CondensePathWithLocationAliases(us.LocationAliases, data.Movie.FolderPath)
	if slices.Contains(data.Movie.Tags, "anime") {
		entryInfo.ArtStyle |= db_types.AS_ANIME
	}

	if slices.Contains(data.Movie.Tags, "live-action") {
		entryInfo.ArtStyle |= db_types.AS_LIVE_ACTION
	}

	timezone := us.DefaultTimeZone

	metadata := db_types.MetadataEntry{}
	if !slices.Contains(data.Movie.Tags, "anime") && data.RemoteMovie.ImdbId != "" {
		metadata, err = meta.GetMetadataById(data.RemoteMovie.ImdbId, ctx.Uid, "omdb")
		if err != nil {
			metadata = db_types.MetadataEntry{}
		}
	} else {
		metadata, err = meta.GetMetadata(&meta.GetMetadataInfo{
			Entry:         &entryInfo,
			MetadataEntry: &metadata,
			Uid:           ctx.Uid,
		})
		if err != nil {
			metadata = db_types.MetadataEntry{}
		}
	}

	datapoints := map[string]string{}
	if metadata.Datapoints != "" {
		json.Unmarshal([]byte(metadata.Datapoints), &datapoints)
	}
	datapoints["radarr-id"] = fmt.Sprintf("%d", data.Movie.Id)
	mar_datapoints, _ := json.Marshal(datapoints)
	metadata.Datapoints = string(mar_datapoints)

	if err := db.AddEntry(ctx.Uid, timezone, &entryInfo, &metadata, &userEntry); err != nil {
		util.WError(ctx.W, 500, "Error adding entry\n%s", err.Error())
		return
	}

	j, err := entryInfo.ToJson()
	if err != nil {
		util.WError(ctx.W, 500, "Could not convert new entry to json\n%s", err.Error())
		return
	}

	ctx.W.WriteHeader(201)
	ctx.W.Write(j)
}

func HookRadarr(ctx RequestContext) {
	body, err := io.ReadAll(ctx.Req.Body)

	defer ctx.Req.Body.Close()
	if err != nil {
		util.WError(ctx.W, 500, "Failed to read body\n%s", err.Error())
		return
	}

	data := RadarrPostWebhook{}
	if err := json.Unmarshal(body, &data); err != nil {
		util.WError(ctx.W, 400, "Failed to parse body\n%s", err.Error())
		return
	}

	if data.EventType == "Test" {
		ctx.W.WriteHeader(200)
		ctx.W.Write([]byte("OK"))
		return
	} else if data.EventType == "MovieAdded" {
		_radarrAdd(ctx, data)
	} else {
		util.WError(ctx.W, 422, "Unknown event: %s\n", data.EventType)
	}
}

// lets the user add an item in their library
func AddEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	title := parsedParams["title"].(string)

	priceNum := parsedParams.Get("price", 0.0).(float64)

	formatInt := parsedParams["format"].(db_types.Format)
	format_modifiers := int64(0)

	if mods, exists := parsedParams["format-modifiers"]; exists {
		format_modifiers = mods.(int64)
	}

	if digital, exists := parsedParams["is-digital"]; exists {
		if digital.(bool) {
			format_modifiers |= int64(db_types.F_MOD_DIGITAL)
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
	entryInfo.Native_Title = nativeTitle
	entryInfo.Location = location
	entryInfo.Format = db_types.Format(formatInt)
	entryInfo.ArtStyle = db_types.ArtStyle(style)
	entryInfo.Type = parsedParams["type"].(db_types.MediaTypes)
	entryInfo.Library = libraryId
	entryInfo.Format_Modifiers = uint64(format_modifiers)

	recommendedList := parsedParams.Get("recommended-by", "").(string)
	if recommendedList != "" {
		recommended := strings.Split(recommendedList, ",")
		recommendedText, _ := json.Marshal(recommended)
		entryInfo.RecommendedBy = string(recommendedText)
	}

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

	// "metadata" and "get-metadata" are mutually exclusive, otherwise
	// it makes no sense since get-metadata would just override metadata
	if data := parsedParams.Get("metadata", "").(string); data != "" {
		var newMeta db_types.MetadataEntry
		err := json.Unmarshal([]byte(data), &newMeta)
		if err != nil {
			util.WError(w, 400, "Failed to parse metadata json\n%s", err.Error())
			return
		}
		newMeta.ItemId = entryInfo.ItemId
		newMeta.Uid = ctx.Uid
		metadata = newMeta
	} else if parsedParams.Get("get-metadata", false).(bool) {
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

	if copyOfId != 0 {
		db.AddRelation(ctx.Uid, entryInfo.ItemId, db_types.R_Copy, copyOfId)
	}

	if parentId != 0 {
		db.AddRelation(ctx.Uid, entryInfo.ItemId, db_types.R_Child, parentId)
	}

	if requiresId != 0 {
		db.AddRelation(ctx.Uid, entryInfo.ItemId, db_types.R_Requires, requiresId)
	}

	if tags != "" {
		if err := db.AddTags(ctx.Uid, entryInfo.ItemId, strings.Split(tags, ",")); err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error adding tags table\n" + err.Error()))
			return
		}
	}

	if priceNum > 0 {
		currency := parsedParams.Get("currency", "USD").(string)
		db.CreateTransaction("Purchased", ctx.Uid, entryInfo.ItemId, 0, timezone, priceNum, currency)
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
	entries, err := db.ListEntries(actx2dctx(ctx), sortBy)
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

func QueryEntries4(ctx RequestContext) {
	search := ctx.PP["search"].(string)
	orderBy := ctx.PP.Get("order-by", "").(string)
	println(orderBy)
	results, err := db.Search4(actx2dctx(ctx), search, orderBy)
	if err != nil {
		util.WError(ctx.W, 500, "Could not complete search\n%s", err.Error())
		return
	}

	ctx.W.WriteHeader(200)
	writeSQLRowResults(ctx.W, results)
}

func QueryEntries3(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	search := pp["search"].(string)

	// can be -1 if user does not provide uid
	if ctx.Uid > 0 {
		search += fmt.Sprintf(" & {entryInfo.uid = %d}", ctx.Uid)
	}

	results, err := db.Search3(actx2dctx(ctx), search, pp.Get("order-by", "").(string))
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

	copies, err := db.GetCopiesOf(actx2dctx(ctx), entry.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not get copies of %d\n%s", entry.ItemId, err.Error())
		return
	}
	w.WriteHeader(200)
	writeSQLRowResults(w, copies)
}

func GetEntry(ctx RequestContext) {
	entry := ctx.PP["id"].(db_types.InfoEntry)
	res, err := json.Marshal(&entry)
	if err != nil {
		util.WError(ctx.W, 500, "Could not marshal info item: %s\n", err.Error())
		return
	}
	ctx.W.Write(res)
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

	items, err := db.GetDescendants(actx2dctx(ctx), entry.ItemId)
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
	tree, err := db.BuildEntryTree(db.RequestContext{
		UID:  ctx.Uid,
		Auth: ctx.Authorized,
	}, 0)
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

func EntrySettings(ctx RequestContext) {
	settings, err := db.GetEntrySettings(ctx.PP["id"].(db_types.InfoEntry).ItemId)
	if ctx.Req.Method == "GET" {
		if err != nil {
			util.WError(ctx.W, 500, "Failed to get settings for entry: %s\n", err.Error())
			return
		}

		res, err := json.Marshal(settings)
		if err != nil {
			util.WError(ctx.W, 500, "Failed to marshal settings: %s\n", err.Error())
			return
		}

		ctx.W.WriteHeader(200)
		ctx.W.Write(res)
	} else if ctx.Req.Method == "POST" {
		body, err := io.ReadAll(ctx.Req.Body)
		defer ctx.Req.Body.Close()

		if err != nil {
			util.WError(ctx.W, 500, "Failed to read body: %s\n", err.Error())
			return
		}

		newSettings := db_types.EntrySettings{
			ItemId: settings.ItemId,
			Permissions: settings.Permissions,
		}

		if err = json.Unmarshal(body, &newSettings); err != nil {
			util.WError(ctx.W, 400, "Invalid body: %s\n", err.Error())
			return
		}

		//for now can only be 0 or PERM_READ (1)
		if newSettings.Permissions < 0 || newSettings.Permissions > 1{
			util.WError(ctx.W, 400, "Bad permissions specified\n")
			return
		}

		if err = db.SetEntrySettings(newSettings); err != nil {
			util.WError(ctx.W, 500, "Failed to set settings: %s\n", err.Error())
			return
		}
		success(ctx.W)
	} else {
		ctx.W.WriteHeader(405)
	}
}

func EntryResource(ctx RequestContext) {
	switch ctx.Req.Method {
	case "QUERY":
		v := ctx.PP.Get("v", int64(4)).(int64)
		ctx.PP["search"] = ctx.PP["q"]
		switch v {
		case 3:
			QueryEntries3(ctx)
		case 4:
			QueryEntries4(ctx)
		}
	case "GET":
		kind := ctx.PP["kind"].(string)
		switch kind {
		case "relations":
			ListRelationsAsJSONL(ctx)
		case "events":
			ListEvents(ctx)
		case "transactions":
			ListTransactions(ctx)
		case "info":
			ListEntries(ctx)
		case "meta":
			ListMetadata(ctx)
		case "user":
			UserEntries(ctx)
		}
	}
}

func actionMedia(ctx RequestContext, fn func(RequestContext)) {
	user, err := db.GetUserEntry(actx2dctx(ctx), ctx.PP["id"].(db_types.InfoEntry).ItemId)
	if err != nil {
		util.WError(ctx.W, 500, "Could not get corresponding user viewing info\n%s", err.Error())
		return
	}
	ctx.PP["id"] = user
	fn(ctx)
}

func SpecificEntry(ctx RequestContext) {
	switch ctx.Req.Method {
	case "TREE":
		GetTree(ctx)
	case "DROP":
		actionMedia(ctx, DropMedia)
	case "RESUME":
		actionMedia(ctx, ResumeMedia)
	case "PAUSE":
		actionMedia(ctx, PauseMedia)
	case "PLAN":
		actionMedia(ctx, PlanMedia)
	case "BEGIN":
		actionMedia(ctx, BeginMedia)
	case "WAIT":
		actionMedia(ctx, WaitMedia)
	case "FINISH":
		if _, has := ctx.PP["rating"]; !has {
			util.WError(ctx.W, 400, "?rating not specified with FINISH\n")
			return
		}
		actionMedia(ctx, FinishMedia)
	case "TRANSACT":
		Transact(ctx)
	case "POST":
		AddEntry(ctx)
	case "PATCH":
		SetEntry(ctx)
	case "DELETE":
		DeleteEntry(ctx)
	case "GET":
		kind := ctx.PP.Get("kind", "info").(string)
		switch kind {
		case "relations":
		case "events":
			GetEventsOf(ctx)
		case "transactions":
			ListTransactions(ctx)
		case "children":
			GetDescendants(ctx)
		case "info":
			GetEntry(ctx)
		case "meta":
			meta, err := db.GetMetadataEntryById(actx2dctx(ctx), ctx.PP["id"].(db_types.InfoEntry).ItemId)
			if err != nil {
				util.WError(ctx.W, 500, "Could not get corresponding meta info\n%s", err.Error())
			}
			ctx.PP["id"] = meta
			RetrieveMetadataForEntry(ctx)
		case "user":
			actionMedia(ctx, GetUserEntry)
		}
	}
}

func success(w http.ResponseWriter) {
	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}
