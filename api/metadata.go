package api

import (
	"aiolimas/db"
	"aiolimas/metadata"
	"encoding/json"
	"net/http"
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
