package api

import (
	"encoding/json"
	"io"

	"aiolimas/db"
	"aiolimas/logging"
	"aiolimas/metadata"
	"aiolimas/types"
	"aiolimas/util"
)

func FetchLocation(ctx RequestContext) {
	m := ctx.PP["id"].(db_types.MetadataEntry)
	provider := ctx.PP.Get("provider", "").(string)
	providerIDOverride := ctx.PP.Get("provider-id", "").(string)

	providerID := m.ProviderID
	if providerIDOverride != "" {
		providerID = providerIDOverride
	}

	if provider == "" {
		info, _ := db.GetInfoEntryById(ctx.Uid, m.ItemId)
		provider = metadata.DetermineBestLocationProvider(&info, &m)
	}

	if provider == "" {
		util.WError(ctx.W, 400, "could not determine a good location provider")
		return
	}

	location, err := metadata.GetLocation(providerID, ctx.Uid, provider)
	if err != nil {
		util.WError(ctx.W, 500, "could not get location: %s", err)
		return
	}

	entry, _ := db.GetInfoEntryById(ctx.Uid, m.ItemId)
	entry.Location = location

	err = db.UpdateInfoEntry(ctx.Uid, &entry)
	if err != nil {
		util.WError(ctx.W, 500, "failed to update entry location: %s", err)
		return
	}
	ctx.W.WriteHeader(200)
	ctx.W.Write([]byte(location))
}

func FetchMetadataForEntry(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W
	req := ctx.Req
	mainEntry := pp["id"].(db_types.InfoEntry)

	metadataEntry, err := db.GetMetadataEntryById(ctx.Uid, mainEntry.ItemId)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}

	providerOverride := req.URL.Query().Get("provider")
	if !metadata.IsValidProvider(providerOverride) {
		providerOverride = ""
	}

	newMeta, err := metadata.GetMetadata(&metadata.GetMetadataInfo{
		Entry:         &mainEntry,
		MetadataEntry: &metadataEntry,
		Override:      providerOverride,
		Uid:           ctx.Uid,
	})
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}
	newMeta.ItemId = mainEntry.ItemId
	newMeta.Uid = mainEntry.Uid
	err = db.UpdateMetadataEntry(ctx.Uid, &newMeta)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}
	err = db.UpdateInfoEntry(ctx.Uid, &mainEntry)
	if err != nil {
		util.WError(w, 500, "%s\n", err.Error())
		return
	}

	data, err := newMeta.ToJson()
	if err != nil {
		util.WError(ctx.W, 500, "could not serialize new metadata (new metadata saved): %s", err)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
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
	w := ctx.W
	req := ctx.Req
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		util.WError(w, 500, "Could not read body\n%s", err.Error())
		return
	}

	var meta db_types.MetadataEntry
	err = json.Unmarshal(data, &meta)
	if err != nil {
		util.WError(w, 400, "Could not parse json\n%s", err.Error())
		return
	}

	err = db.UpdateMetadataEntry(ctx.Uid, &meta)
	if err != nil {
		util.WError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(ctx.Uid, meta.ItemId)
	if err != nil {
		util.WError(w, 500, "Could not retrieve updated entry\n%s", err.Error())
		return
	}

	outJson, err := json.Marshal(entry)
	if err != nil {
		util.WError(w, 500, "Could not marshal new user entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(outJson)
}

func SetThumbnail(ctx RequestContext) {
	body, err := io.ReadAll(ctx.Req.Body)
	if err != nil {
		util.WError(ctx.W, 500, "Could not read request body\n%s", err.Error())
	}

	entry := ctx.PP["id"].(db_types.MetadataEntry)
	entry.Thumbnail = string(body)
	db.UpdateMetadataEntry(ctx.Uid, &entry)
	success(ctx.W)
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

	err := db.UpdateMetadataEntry(ctx.Uid, &metadataEntry)
	if err != nil {
		util.WError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	success(w)
}

func ListMetadata(ctx RequestContext) {
	w := ctx.W
	items, err := db.ListMetadata(ctx.Uid)
	if err != nil {
		util.WError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	for _, item := range items {
		j, err := item.ToJson()
		if err != nil {
			logging.ELog(err)
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
		Title:  title,
		ForUid: ctx.Uid,
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
			logging.ELog(err)
			continue
		}
		w.Write(text)
		w.Write([]byte("\n"))
	}
}

func FinalizeIdentification(ctx RequestContext) {
	parsedParams := ctx.PP
	w := ctx.W
	itemToApplyTo := parsedParams.Get("apply-to", db_types.MetadataEntry{ItemId: 0, Title: "DOES NOT EXIST"}).(db_types.MetadataEntry)
	id := parsedParams["identified-id"].(string)
	provider := parsedParams["provider"].(string)

	data, err := metadata.GetMetadataById(id, ctx.Uid, provider)
	if err != nil {
		util.WError(w, 500, "Could not get metadata\n%s", err.Error())
		return
	}

	if itemToApplyTo.ItemId != 0 {
		data.ItemId = itemToApplyTo.ItemId
		data.Uid = itemToApplyTo.Uid
		err = db.UpdateMetadataEntry(ctx.Uid, &data)
		if err != nil {
			util.WError(w, 500, "Failed to update metadata\n%s", err.Error())
			return
		}
	}

	j, err := data.ToJson()
	if err != nil {
		util.WError(w, 500, "Failed to serialize metadata\n%s", err.Error());
		return
	}
	w.Write(j)
}
