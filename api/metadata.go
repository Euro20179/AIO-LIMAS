package api

import (
	"encoding/json"
	"io"
	"net/http"

	"aiolimas/db"
	"aiolimas/metadata"
)

func FetchMetadataForEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	mainEntry := pp["id"].(db.InfoEntry)

	metadataEntry, err := db.GetMetadataEntryById(mainEntry.ItemId)
	if err != nil {
		wError(w, 500, "%s\n", err.Error())
		return
	}

	providerOverride := req.URL.Query().Get("provider")
	if !metadata.IsValidProvider(providerOverride) {
		providerOverride = ""
	}

	newMeta, err := metadata.GetMetadata(&mainEntry, &metadataEntry, providerOverride)
	if err != nil {
		wError(w, 500, "%s\n", err.Error())
		return
	}
	newMeta.ItemId = mainEntry.ItemId
	err = db.UpdateMetadataEntry(&newMeta)
	if err != nil {
		wError(w, 500, "%s\n", err.Error())
		return
	}

	success(w)
}

func RetrieveMetadataForEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	entry := pp["id"].(db.MetadataEntry)

	data, err := json.Marshal(entry)
	if err != nil {
		wError(w, 500, "%s\n", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func SetMetadataEntry(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil{
		wError(w, 500, "Could not read body\n%s", err.Error())
		return
	}

	var meta db.MetadataEntry
	err = json.Unmarshal(data, &meta)
	if err != nil{
		wError(w, 400, "Could not parse json\n%s", err.Error())
		return
	}

	err = db.UpdateMetadataEntry(&meta)
	if err != nil{
		wError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	entry, err := db.GetUserViewEntryById(meta.ItemId)
	if err != nil{
		wError(w, 500, "Could not retrieve updated entry\n%s", err.Error())
		return
	}

	outJson, err := json.Marshal(entry)
	if err != nil{
		wError(w, 500, "Could not marshal new user entry\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(outJson)
}

func ModMetadataEntry(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	metadataEntry := pp["id"].(db.MetadataEntry)

	metadataEntry.Rating = pp.Get("rating", metadataEntry.Rating).(float64)

	metadataEntry.Description = pp.Get("description", metadataEntry.Description).(string) 

	metadataEntry.ReleaseYear = pp.Get("release-year", metadataEntry.ReleaseYear).(int64)

	metadataEntry.Thumbnail = pp.Get("thumbnail", metadataEntry.Thumbnail).(string)
	metadataEntry.MediaDependant = pp.Get("media-dependant", metadataEntry.MediaDependant).(string)
	metadataEntry.Datapoints = pp.Get("datapoints", metadataEntry.Datapoints).(string)

	err := db.UpdateMetadataEntry(&metadataEntry)
	if err != nil{
		wError(w, 500, "Could not update metadata entry\n%s", err.Error())
		return
	}

	success(w)
}

func ListMetadata(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	items, err := db.Db.Query("SELECT * FROM metadata")
	if err != nil {
		wError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}

	defer items.Close()

	w.WriteHeader(200)
	for items.Next() {
		var row db.MetadataEntry
		err := row.ReadEntry(items)
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

func IdentifyWithSearch(w http.ResponseWriter, req *http.Request, parsedParsms ParsedParams) {
	title := parsedParsms["title"].(string)
	search := metadata.IdentifyMetadata{
		Title: title,
	}

	infoList, provider, err := metadata.Identify(search, parsedParsms["provider"].(string))
	if err != nil {
		wError(w, 500, "Could not identify\n%s", err.Error())
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

func FinalizeIdentification(w http.ResponseWriter, req *http.Request, parsedParams ParsedParams) {
	itemToApplyTo := parsedParams["apply-to"].(db.MetadataEntry)
	id := parsedParams["identified-id"].(string)
	provider := parsedParams["provider"].(string)

	data, err := metadata.GetMetadataById(id, provider)
	if err != nil {
		wError(w, 500, "Could not get metadata\n%s", err.Error())
		return
	}

	data.ItemId = itemToApplyTo.ItemId
	err = db.UpdateMetadataEntry(&data)
	if err != nil {
		wError(w, 500, "Failed to update metadata\n%s", err.Error())
		return
	}
}
