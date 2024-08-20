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

	//paths
	//<root> general database stuff
	//

	type EndPointMap map[string]func(http.ResponseWriter, *http.Request)

	apiRoot := "/api/v1"
	makeEndpoints(apiRoot, EndPointMap{
		"add-entry":    api.AddEntry,
		"query":        api.QueryEntries,
		"list-entries": api.ListEntries,
		"scan-folder":  api.ScanFolder,
	})

	//for metadata stuff
	makeEndpoints(apiRoot+"/metadata", EndPointMap {
		"fetch": api.FetchMetadataForEntry,
		"retrieve": api.RetrieveMetadataForEntry,
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
	})

	http.ListenAndServe(":8080", nil)
}
