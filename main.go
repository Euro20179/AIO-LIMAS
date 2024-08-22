package main

import (
	"flag"
	"net/http"

	api "aiolimas/api"
	db "aiolimas/db"
)

func makeEndpoints(root string, endPoints map[string]func(http.ResponseWriter, *http.Request)) {
	for name, fn := range endPoints {
		http.HandleFunc(root+"/"+name, fn)
	}
}

func main() {
	dbPathPtr := flag.String("db-path", "./all.db", "Path to the database file")
	flag.Parse()
	db.InitDb(*dbPathPtr)

	type EndPointMap map[string]func(http.ResponseWriter, *http.Request)

	apiRoot := "/api/v1"

	//for db management type stuff
	makeEndpoints(apiRoot, EndPointMap{
		"add-entry":    api.AddEntry,
		"mod-entry":    api.ModEntry,
		"query":        api.QueryEntries,
		"list-entries": api.ListEntries,
		"scan-folder":  api.ScanFolder,
		"stream-entry": api.Stream,
	})

	//for metadata stuff
	makeEndpoints(apiRoot+"/metadata", EndPointMap {
		"fetch": api.FetchMetadataForEntry,
		"retrieve": api.RetrieveMetadataForEntry,
		"set": api.SetMetadataForEntry,
		"list-entries": api.ListMetadata,
	})

	// for stuff relating to user viewing info
	// such as user rating, user beginning/ending a media, etc
	// stuff that would normally be managed by strack
	makeEndpoints(apiRoot+"/engagement", EndPointMap{
		"begin-media":  api.BeginMedia,
		"finish-media": api.FinishMedia,
		"plan-media": api.PlanMedia,
		"drop-media": api.DropMedia,
		"pause-media": api.PauseMedia,
		"resume-media": api.ResumeMedia,
		"set-note": api.SetNote,
		"get-entry": api.GetUserEntry,
		"list-entries": api.UserEntries,
	})

	http.ListenAndServe(":8080", nil)
}
