package api

import (
	"aiolimas/db"
	"aiolimas/metadata"
	"encoding/json"
	"net/http"
	"strconv"
)

func FetchMetadataForEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil{
		wError(w, 400, "Could not find entry\n")
		return
	}

	mainEntry, err := db.GetInfoEntryById(entry.ItemId)

	metadataEntry, err := db.GetMetadataEntryById(entry.ItemId)
	if err != nil{
		wError(w, 500, "%s\n", err.Error())
		return
	}

	providerOverride := req.URL.Query().Get("provider")
	if !metadata.IsValidProvider(providerOverride) {
		providerOverride = ""
	}

	newMeta, err := metadata.GetMetadata(&mainEntry, &metadataEntry, providerOverride)
	if err != nil{
		wError(w, 500, "%s\n", err.Error())
		return
	}
	db.UpdateMetadataEntry(&newMeta)

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}

func RetrieveMetadataForEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil{
		wError(w, 400, "Could not find entry\n")
		return
	}

	metadataEntry, err := db.GetMetadataEntryById(entry.ItemId)
	if err != nil{
		wError(w, 500, "%s\n", err.Error())
		return
	}

	data, err := json.Marshal(metadataEntry)
	if err != nil{
		wError(w, 500, "%s\n", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func SetMetadataForEntry(w http.ResponseWriter, req *http.Request) {
	entry, err := verifyIdAndGetUserEntry(w, req)
	if err != nil{
		wError(w, 400, "Could not find entry\n")
		return
	}

	metadataEntry, err := db.GetMetadataEntryById(entry.ItemId)
	if err != nil{
		wError(w, 500, "%s\n", err.Error())
		return
	}

	rating := req.URL.Query().Get("rating")
	ratingF, err := strconv.ParseFloat(rating, 64)
	if err != nil{
		ratingF = metadataEntry.Rating
	}
	metadataEntry.Rating = ratingF

	description := req.URL.Query().Get("description")
	if description != "" {
		metadataEntry.Description = description
	}

	releaseYear := req.URL.Query().Get("release-year")
	releaseYearInt, err := strconv.ParseInt(releaseYear, 10, 64)
	if err != nil{
		releaseYearInt = metadataEntry.ReleaseYear
	}
	metadataEntry.ReleaseYear = releaseYearInt

	thumbnail := req.URL.Query().Get("thumbnail")
	if thumbnail != "" {
		metadataEntry.Thumbnail = thumbnail
	}

	mediaDependantJson := req.URL.Query().Get("media-dependant")
	if mediaDependantJson != "" {
		metadataEntry.MediaDependant = mediaDependantJson
	}

	datapointsJson := req.URL.Query().Get("datapoints")
	if datapointsJson != "" {
		metadataEntry.Datapoints = datapointsJson
	}

	db.UpdateMetadataEntry(&metadataEntry)

	w.WriteHeader(200)
	w.Write([]byte("Success\n"))
}

func ListMetadata(w http.ResponseWriter, req *http.Request) {
	items, err := db.Db.Query("SELECT * FROM metadata")
	if err != nil{
		wError(w, 500, "Could not fetch data\n%s", err.Error())
		return
	}

	w.WriteHeader(200)
	for items.Next() {
		var row db.MetadataEntry
		err := row.ReadEntry(items)
		if err != nil{
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
