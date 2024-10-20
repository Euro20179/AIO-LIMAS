package api

import (
	"aiolimas/db"
	"encoding/json"
	"net/http"
)

func ListFormats(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	text, err := json.Marshal(db.ListFormats())
	if err != nil{
		wError(w, 500, "Could not encode formats\n%s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(text)
}

func ListTypes(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	text, err := json.Marshal(db.ListMediaTypes())
	if err != nil{
		wError(w, 500, "Could not encode types\n%s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(text)
}

func ListArtStyles(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	text, err := json.Marshal(db.ListArtStyles())
	if err != nil{
		wError(w, 500, "Could not encode artstyles\n%s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(text)
}
