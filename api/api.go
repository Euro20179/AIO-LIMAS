package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func GetAllForEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	info := parsedParams["id"].(db.InfoEntry)

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

	for _, event := range events {
		ej, err := event.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(ej)
		w.Write([]byte("\n"))
	}
}

func SetEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		wError(w, 500, "Could not ready body\n%s", err.Error())
		return
	}

	var entry db.InfoEntry
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
	info := parsedParams["id"].(db.InfoEntry)

	title, exists := parsedParams["en-title"].(string)
	if exists {
		info.En_Title = title
	}

	nativeTitle, exists := parsedParams["native-title"].(string)
	if exists {
		info.Native_Title = nativeTitle
	}

	format, exists := parsedParams["format"].(db.Format)
	if exists {
		info.Format = format
	}

	parent, exists := parsedParams["parent-id"].(db.InfoEntry)
	if exists {
		info.ParentId = parent.ItemId
	}

	if orphan, exists := parsedParams["become-orphan"].(bool); exists && orphan {
		info.ParentId = 0
	}

	if original, exists := parsedParams["become-original"].(bool); exists && original {
		info.CopyOf = 0
	}

	if itemCopy, exists := parsedParams["copy-id"].(db.InfoEntry); exists {
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

	if isAnime, exists := parsedParams["is-anime"].(bool); exists {
		info.IsAnime = isAnime
	}

	info.Type = parsedParams.Get("type", info.Type).(db.MediaTypes)

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

	formatInt := parsedParams["format"].(db.Format)

	if digital, exists := parsedParams["is-digital"]; exists {
		if digital.(bool) {
			formatInt |= db.F_MOD_DIGITAL
		}
	}

	var parentId int64 = 0
	if parent, exists := parsedParams["parentId"]; exists {
		parentId = parent.(db.InfoEntry).ItemId
	}

	var copyOfId int64 = 0

	if c, exists := parsedParams["copyOf"]; exists {
		copyOfId = c.(db.InfoEntry).ItemId
	}

	isAnime := false
	if anime, exists := parsedParams["is-anime"]; exists {
		isAnime = anime.(bool)
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

	var entryInfo db.InfoEntry
	entryInfo.En_Title = title
	entryInfo.PurchasePrice = priceNum
	entryInfo.Native_Title = nativeTitle
	entryInfo.Collection = tags
	entryInfo.Location = location
	entryInfo.Format = db.Format(formatInt)
	entryInfo.ParentId = parentId
	entryInfo.IsAnime = isAnime
	entryInfo.CopyOf = copyOfId
	entryInfo.Type = parsedParams["type"].(db.MediaTypes)

	var metadata db.MetadataEntry

	var userEntry db.UserViewingEntry

	if userRating, exists := parsedParams["user-rating"]; exists {
		userEntry.UserRating = userRating.(float64)
	}
	if status, exists := parsedParams["user-status"]; exists {
		userEntry.Status = status.(db.Status)
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

func QueryEntries2(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	names := pp["names"].([]string)
	values := pp["values"].([]string)
	checkers := pp["checkers"].([]db.DataChecker)

	var query db.SearchQuery

	for i, name := range names {
		data := db.SearchData{
			DataName: name,
			Checker:  checkers[i],
		}
		if checkers[i] == db.DATA_NOTIN || checkers[i] == db.DATA_IN {
			values = strings.Split(values[i], ":")
			data.DataValue = values
		} else {
			data.DataValue = values[i]
		}
		query = append(query, data)
	}

	results, err := db.Search2(query)
	if err != nil{
		wError(w, 500, "Could not complete search\n%s", err.Error())
		return
	}
	w.WriteHeader(200)
	for _, row := range results {
		j, err := row.ToJson()
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(j)
		w.Write([]byte("\n"))
	}
}

// should be an all in one query thingy, that should be able to query based on any matching column in any table
func QueryEntries(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	title := parsedParams.Get("title", "").(string)
	nativeTitle := parsedParams.Get("native-title", "").(string)
	location := parsedParams.Get("location", "").(string)
	pgt := parsedParams.Get("purchase-gt", 0.0).(float64)
	plt := parsedParams.Get("purchase-lt", 0.0).(float64)
	formats := parsedParams.Get("formats", "").(string)
	tags := parsedParams.Get("tags", "").(string)
	types := parsedParams.Get("types", "").(string)
	parents := parsedParams.Get("parent-ids", "").(string)
	copyIds := parsedParams.Get("copy-ids", "").(string)
	status := parsedParams.Get("user-status", []db.Status{}).([]db.Status)
	userRatingGt := parsedParams.Get("user-rating-gt", 0.0).(float64)
	userRatingLt := parsedParams.Get("user-rating-lt", 0.0).(float64)
	isAnime := parsedParams.Get("is-anime", int64(0)).(int64)

	releasedGe := parsedParams.Get("released-ge", int64(0)).(int64)
	releasedLe := parsedParams.Get("released-le", int64(0)).(int64)

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
	entrySearch.UserRatingLt = userRatingLt
	entrySearch.UserRatingGt = userRatingGt
	entrySearch.InTags = collects
	entrySearch.Format = fmts
	entrySearch.Type = tys
	entrySearch.HasParent = pars
	entrySearch.CopyIds = cos
	entrySearch.UserStatus = status
	entrySearch.IsAnime = int(isAnime)
	entrySearch.ReleasedGE = releasedGe
	entrySearch.ReleasedLE = releasedLe

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

func GetCopies(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db.InfoEntry)

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

func Stream(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	entry := parsedParams["id"].(db.InfoEntry)

	http.ServeFile(w, req, entry.Location)
}

func DeleteEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db.InfoEntry)
	err := db.Delete(entry.ItemId)
	if err != nil {
		wError(w, 500, "Could not delete entry\n%s", err.Error())
		return
	}
	success(w)
}

func GetDescendants(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db.InfoEntry)

	items, err := db.GetDescendants(entry.ItemId)
	if err != nil {
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
	info := pp["id"].(db.InfoEntry)
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
