package main

import (
	"flag"
	"net/http"

	api "aiolimas/api"
	db "aiolimas/db"
	"aiolimas/webservice"
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

	addEntry := api.ApiEndPoint{
		Handler: api.AddEntry,
		QueryParams: api.QueryParams{
			"title":        api.MkQueryInfo(api.P_NotEmpty, true),
			"type":         api.MkQueryInfo(api.P_EntryType, true),
			"format":       api.MkQueryInfo(api.P_EntryFormat, true),
			"price":        api.MkQueryInfo(api.P_Float64, false),
			"is-digital":   api.MkQueryInfo(api.P_Bool, false),
			"is-anime":     api.MkQueryInfo(api.P_Bool, false),
			"parentId":     api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"copyOf":       api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"native-title": api.MkQueryInfo(api.P_True, false),
			"tags":         api.MkQueryInfo(api.P_True, false),
			"location":     api.MkQueryInfo(api.P_True, false),

			"get-metadata":      api.MkQueryInfo(api.P_Bool, false),
			"metadata-provider": api.MkQueryInfo(api.P_MetaProvider, false),

			"user-rating":     api.MkQueryInfo(api.P_Float64, false),
			"user-status":     api.MkQueryInfo(api.P_UserStatus, false),
			"user-view-count": api.MkQueryInfo(api.P_Int64, false),
			"user-notes":      api.MkQueryInfo(api.P_True, false),
		},
	}

	modEntry := api.ApiEndPoint{
		Handler: api.ModEntry,
		QueryParams: api.QueryParams{
			"id":           api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"en-title":     api.MkQueryInfo(api.P_NotEmpty, false),
			"native-title": api.MkQueryInfo(api.P_True, false),
			"format":       api.MkQueryInfo(api.P_EntryFormat, false),
			"parent-id":    api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"copy-id":      api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"price":        api.MkQueryInfo(api.P_Float64, false),
			"location":     api.MkQueryInfo(api.P_True, false),
			"tags":         api.MkQueryInfo(api.P_True, false),
			"is-anime":     api.MkQueryInfo(api.P_Bool, false),
		},
	}

	searchApi := api.ApiEndPoint{
		Handler: api.ListEntries,
		QueryParams: api.QueryParams{
			"sort-by": api.MkQueryInfo(api.P_SqlSafe, false),
		},
	}

	// for db management type stuff
	makeEndpoints(apiRoot, EndPointMap{
		"add-entry":        addEntry.Listener,
		"mod-entry":        modEntry.Listener,
		"query":            api.QueryEntries,
		"list-entries":     searchApi.Listener,
		"scan-folder":      api.ScanFolder,
		"stream-entry":     api.Stream,
		"delete-entry":     api.DeleteEntry,
		"list-collections": api.ListCollections,
		"list-copies":      api.GetCopies,
		"list-descendants": api.GetDescendants,
		"total-cost":       api.TotalCostOf,
		"list-tree":        api.GetTree,
	})

	makeEndpoints(apiRoot+"/type", EndPointMap{
		"format": api.ListFormats,
	})

	identify := api.ApiEndPoint{
		Handler: api.IdentifyWithSearch,
		QueryParams: api.QueryParams{
			"title": api.MkQueryInfo(api.P_NotEmpty, true),
		},
	}

	finalizeIdentify := api.ApiEndPoint {
		Handler: api.FinalizeIdentification,
		QueryParams: api.QueryParams {
			"id": api.MkQueryInfo(api.P_NotEmpty, true),
			"provider": api.MkQueryInfo(api.P_IdIdentifier, true),
			"apply-to": api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
	}

	// for metadata stuff
	makeEndpoints(apiRoot+"/metadata", EndPointMap{
		"fetch":        api.FetchMetadataForEntry,
		"retrieve":     api.RetrieveMetadataForEntry,
		"set":          api.SetMetadataForEntry,
		"list-entries": api.ListMetadata,
		"identify":     identify.Listener,
		"finalize-identify": finalizeIdentify.Listener,
	})

	finishEngagement := api.ApiEndPoint{
		Handler: api.FinishMedia,
		QueryParams: api.QueryParams{
			"id":     api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"rating": api.MkQueryInfo(api.P_Float64, true),
		},
	}

	reassociate := api.ApiEndPoint{
		Handler: api.CopyUserViewingEntry,
		QueryParams: api.QueryParams{
			"src-id":  api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"dest-id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	getEvents := api.ApiEndPoint{
		Handler: api.GetEventsOf,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	listEvents := api.ApiEndPoint {
		Handler: api.ListEvents,
		QueryParams: api.QueryParams {},
	}

	// for stuff relating to user viewing info
	// such as user rating, user beginning/ending a media, etc
	// stuff that would normally be managed by strack
	makeEndpoints(apiRoot+"/engagement", EndPointMap{
		"begin-media":  api.BeginMedia,
		"finish-media": finishEngagement.Listener,
		"plan-media":   api.PlanMedia,
		"drop-media":   api.DropMedia,
		"pause-media":  api.PauseMedia,
		"resume-media": api.ResumeMedia,
		"set-note":     api.SetNote,
		"get-entry":    api.GetUserEntry,
		"list-entries": api.UserEntries,
		"copy":         reassociate.Listener,
		"get-events":   getEvents.Listener,
		"list-events": listEvents.Listener,
	})

	http.HandleFunc("/", webservice.Root)

	http.ListenAndServe(":8080", nil)
}
