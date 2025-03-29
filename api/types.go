package api

import (
	"encoding/json"

	"aiolimas/util"
	"aiolimas/types"
)

func ListFormats(ctx RequestContext) {
	w := ctx.W
	text, err := json.Marshal(db_types.ListFormats())
	if err != nil{
		util.WError(w, 500, "Could not encode formats\n%s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(text)
}

func ListTypes(ctx RequestContext) {
	w := ctx.W
	text, err := json.Marshal(db_types.ListMediaTypes())
	if err != nil{
		util.WError(w, 500, "Could not encode types\n%s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(text)
}

func ListArtStyles(ctx RequestContext) {
	w := ctx.W
	text, err := json.Marshal(db_types.ListArtStyles())
	if err != nil{
		util.WError(w, 500, "Could not encode artstyles\n%s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(text)
}
