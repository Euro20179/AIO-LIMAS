package api

import (
	"encoding/json"
	"io"

	"aiolimas/util"
	"aiolimas/db"
	"aiolimas/metadata"
	"aiolimas/types"
)

func FetchMetadataForEntry(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	req := ctx.Req
	mainEntry := pp["id"].(db_types.InfoEntry)

	metadataEntry, err := db.GetMetadataEntryById(pp["uid"].(int64), mainEntry.ItemId)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}

	providerOverride := req.URL.Query().Get("provider")
	if !metadata.IsValidProvider(providerOverride) {
		providerOverride = ""
	}

	newMeta, err := metadata.GetMetadata(&mainEntry, &metadataEntry, providerOverride)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}
	newMeta.ItemId = mainEntry.ItemId
	err = db.UpdateMetadataEntry(pp["uid"].(int64), &newMeta)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}
	err = db.UpdateInfoEntry(pp["uid"].(int64), &mainEntry)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}

	success(w)
}

func RetrieveMetadataForEntry(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	entry := pp["id"].(db_types.MetadataEntry)

	data, err := json.Marshal(entry)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func SetMetadataEntry(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	req := ctx.Req
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil{
		util.WError(w, 500, "Could not read body\n%s", err.Error())
		return
	}

	var meta db_types.MetadataEntry
	err = json.Unmarshal(data, &meta)
	if err != nil{
		util.WError(w, 400, "Could not parse json\n%s", err.Error())
		return
	}

	err = db.UpdateMetadataEntry(parsedParams["uid"].(int64), &meta)
	if err != nil{
		util.WError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(parsedParams["uid"].(int64), meta.ItemId)
	if err != nil{
		util.WError(w, 500, "Could not retrieve updated entry\n%s", err.Error())
		return
	}

	outJson, err := json.Marshal(entry)
	if err != nil{
		util.WError(w, 500, "Could not marshal new user entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(outJson)
}

func ModMetadataEntry(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	metadataEntry := pp["id"].(db_types.MetadataEntry)

	metadataEntry.Rating = pp.Get("rating", metadataEntry.Rating).(float64)

	metadataEntry.Description = pp.Get("description", metadataEntry.Description).(string) 

	metadataEntry.ReleaseYear = pp.Get("release-year", metadataEntry.ReleaseYear).(int64)

	metadataEntry.Thumbnail = pp.Get("thumbnail", metadataEntry.Thumbnail).(string)
	metadataEntry.MediaDependant = pp.Get("media-dependant", metadataEntry.MediaDependant).(string)
	metadataEntry.Datapoints = pp.Get("datapoints", metadataEntry.Datapoints).(string)

	err := db.UpdateMetadataEntry(pp["uid"].(int64), &metadataEntry)
	if err != nil{
		util.WError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	success(w)
}

func ListMetadata(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	items, err := db.ListMetadata(pp["uid"].(int64))
	if err != nil {
		util.WError(w, 500, "Could not fetch data\n%s", err.Error())
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

func IdentifyWithSearch(ctx RequestContext) {
	parsedParsms := ctx.PP
	w := ctx.W

	title := parsedParsms["title"].(string)
	search := metadata.IdentifyMetadata{
		Title: title,
	}

	infoList, provider, err := metadata.Identify(search, parsedParsms["provider"].(string))
	if err != nil {
		util.WError(w, 500, "Could not identify\n%s", err.Error())
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(provider))
	w.Write([]byte("\x02")) // start of text
	for _, entry := range infoList {
		text, err := json.Marshal(entry)
		if err != nil {
			println(err.Error())
			continue
		}
		w.Write(text)
		w.Write([]byte("\n"))
	}
}

func FinalizeIdentification(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	itemToApplyTo := parsedParams["apply-to"].(db_types.MetadataEntry)
	id := parsedParams["identified-id"].(string)
	provider := parsedParams["provider"].(string)

	data, err := metadata.GetMetadataById(id, provider)
	if err != nil {
		util.WError(w, 500, "Could not get metadata\n%s", err.Error())
		return
	}

	data.ItemId = itemToApplyTo.ItemId
	err = db.UpdateMetadataEntry(parsedParams["uid"].(int64), &data)
	if err != nil {
		util.WError(w, 500, "Failed to update metadata\n%s", err.Error())
		return
	}
}
