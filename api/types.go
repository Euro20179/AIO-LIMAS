package api

import (
	"aiolimas/db"
	"encoding/json"
	"net/http"
)

func ListFormats(w http.ResponseWriter, req *http.Request) {
	text, err := json.Marshal(db.ListFormats())
	if err != nil{
		wError(w, 500, "Could not encode formats\n%s", err.Error())
		return
	}
	w.WriteHeader(200)
	w.Write(text)
}
