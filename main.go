package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	api "aiolimas/api"
	db "aiolimas/db"
	"aiolimas/webservice"
)

func ckAuthorizationHeader(text string) (bool, error) {
	var estring string

	if b64L := strings.SplitN(text, "Basic ", 2); len(b64L) > 0 {
		b64 := b64L[1]
		info, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			estring = "You're bad at encoding base64 😀\n"
			println(err.Error())
			goto unauthorized
		}
		_, password, found := strings.Cut(string(info), ":")
		if !found {
			estring = "Invalid credentials\n"
			goto unauthorized
		}

		accNumber := os.Getenv("ACCOUNT_NUMBER")
		if password == accNumber {
			return true, nil
		}
	} else {
		goto unauthorized
	}

unauthorized:
	return false, errors.New(estring)
}

func authorizationWrapper(fn func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		auth := req.Header.Get("Authorization")

		accNumber := os.Getenv("ACCOUNT_NUMBER")

		if auth == "" && accNumber != "" {
			w.Header().Add("WWW-Authenticate", "Basic realm=\"/\"")
			w.WriteHeader(401)
			return
		}

		authorized := true
		if accNumber != "" {
			var err error
			authorized, err = ckAuthorizationHeader(auth)
			if !authorized {
				w.WriteHeader(401)
				w.Write([]byte(err.Error()))
			}
		}
		if authorized {
			fn(w, req)
			return
		}
	}
}

func makeEndpoints(root string, endPoints map[string]func(http.ResponseWriter, *http.Request)) {
	for name, fn := range endPoints {
		http.HandleFunc(root+"/"+name, authorizationWrapper(fn))
	}
}

var ( // `/` endpoints {{{
	addEntry = api.ApiEndPoint{
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

	getTree = api.ApiEndPoint{
		Handler:     api.GetTree,
		QueryParams: api.QueryParams{},
	}

	modEntry = api.ApiEndPoint{
		Handler: api.ModEntry,
		QueryParams: api.QueryParams{
			"id":              api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"en-title":        api.MkQueryInfo(api.P_NotEmpty, false),
			"native-title":    api.MkQueryInfo(api.P_True, false),
			"format":          api.MkQueryInfo(api.P_EntryFormat, false),
			"parent-id":       api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"become-orphan":   api.MkQueryInfo(api.P_Bool, false),
			"become-original": api.MkQueryInfo(api.P_Bool, false),
			"copy-id":         api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"price":           api.MkQueryInfo(api.P_Float64, false),
			"location":        api.MkQueryInfo(api.P_True, false),
			"tags":            api.MkQueryInfo(api.P_True, false),
			"is-anime":        api.MkQueryInfo(api.P_Bool, false),
			"type":            api.MkQueryInfo(api.P_EntryType, false),
		},
	}

	setEntry = api.ApiEndPoint{
		Handler:     api.SetEntry,
		QueryParams: api.QueryParams{},
		Method:      "POST",
	}

	listApi = api.ApiEndPoint{
		Handler: api.ListEntries,
		QueryParams: api.QueryParams{
			"sort-by": api.MkQueryInfo(api.P_SqlSafe, false),
		},
	}

	stream = api.ApiEndPoint{
		Handler: api.Stream,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	deleteEntry = api.ApiEndPoint{
		Handler: api.DeleteEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	listCollections = api.ApiEndPoint{
		Handler:     api.ListCollections,
		QueryParams: api.QueryParams{},
	}

	listCopies = api.ApiEndPoint{
		Handler: api.GetCopies,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	listDescendants = api.ApiEndPoint{
		Handler: api.GetDescendants,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	totalCostOf = api.ApiEndPoint{
		Handler: api.TotalCostOf,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	searchApi = api.ApiEndPoint{
		Handler: api.QueryEntries,
		QueryParams: api.QueryParams{
			"title":          api.MkQueryInfo(api.P_True, false),
			"native-title":   api.MkQueryInfo(api.P_True, false),
			"location":       api.MkQueryInfo(api.P_True, false),
			"purchase-gt":    api.MkQueryInfo(api.P_Float64, false),
			"purchase-lt":    api.MkQueryInfo(api.P_Float64, false),
			"formats":        api.MkQueryInfo(api.P_True, false),
			"tags":           api.MkQueryInfo(api.P_True, false),
			"types":          api.MkQueryInfo(api.P_True, false),
			"parents":        api.MkQueryInfo(api.P_True, false),
			"is-anime":       api.MkQueryInfo(api.P_Int64, false),
			"copy-ids":       api.MkQueryInfo(api.P_True, false),
			"user-status":    api.MkQueryInfo(
				api.P_TList(
					",",
					func(in string) db.Status {
						return db.Status(in)
					},
				),
				false,
			),
			"user-rating-gt": api.MkQueryInfo(api.P_Float64, false),
			"user-rating-lt": api.MkQueryInfo(api.P_Float64, false),
			"released-ge":    api.MkQueryInfo(api.P_Int64, false),
			"released-le":    api.MkQueryInfo(api.P_Int64, false),
		},
	}

	getAllEntry = api.ApiEndPoint{
		Handler: api.GetAllForEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}
) // }}}

var ( // `/metadata` endpoints {{{
	identify = api.ApiEndPoint{
		Handler: api.IdentifyWithSearch,
		QueryParams: api.QueryParams{
			"title":    api.MkQueryInfo(api.P_NotEmpty, true),
			"provider": api.MkQueryInfo(api.P_Identifier, true),
		},
	}

	finalizeIdentify = api.ApiEndPoint{
		Handler: api.FinalizeIdentification,
		QueryParams: api.QueryParams{
			"identified-id": api.MkQueryInfo(api.P_NotEmpty, true),
			"provider":      api.MkQueryInfo(api.P_IdIdentifier, true),
			"apply-to":      api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
	}

	setMeta = api.ApiEndPoint{
		Handler:     api.SetMetadataEntry,
		Method:      "POST",
		QueryParams: api.QueryParams{},
	}

	fetchMetadataForEntry = api.ApiEndPoint{
		Handler: api.FetchMetadataForEntry,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"provider": api.MkQueryInfo(api.P_NotEmpty, false),
		},
	}

	retrieveMetadataForEntry = api.ApiEndPoint{
		Handler: api.RetrieveMetadataForEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
	}

	modMetaEntry = api.ApiEndPoint{
		Handler: api.ModMetadataEntry,
		QueryParams: api.QueryParams{
			"id":              api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
			"rating":          api.MkQueryInfo(api.P_Float64, false),
			"description":     api.MkQueryInfo(api.P_NotEmpty, false),
			"release-year":    api.MkQueryInfo(api.P_Int64, false),
			"thumbnail":       api.MkQueryInfo(api.P_NotEmpty, false),
			"media-dependant": api.MkQueryInfo(api.P_NotEmpty, false),
			"datapoints":      api.MkQueryInfo(api.P_NotEmpty, false),
		},
	}

	listMetadata = api.ApiEndPoint{
		Handler:     api.ListMetadata,
		QueryParams: api.QueryParams{},
	}
) // }}}

var ( // `/engagement` endpoints {{{
	finishEngagement = api.ApiEndPoint{
		Handler: api.FinishMedia,
		QueryParams: api.QueryParams{
			"id":     api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"rating": api.MkQueryInfo(api.P_Float64, true),
		},
	}

	reassociate = api.ApiEndPoint{
		Handler: api.CopyUserViewingEntry,
		QueryParams: api.QueryParams{
			"src-id":  api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"dest-id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	getEvents = api.ApiEndPoint{
		Handler: api.GetEventsOf,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}
	deleteEvent = api.ApiEndPoint{
		Handler: api.DeleteEvent,
		QueryParams: api.QueryParams{
			"id":        api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"timestamp": api.MkQueryInfo(api.P_Int64, true),
			"after":     api.MkQueryInfo(api.P_Int64, true),
		},
	}

	registerEvent = api.ApiEndPoint{
		Handler: api.RegisterEvent,
		QueryParams: api.QueryParams{
			"id":        api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"timestamp": api.MkQueryInfo(api.P_Int64, false),
			"after":     api.MkQueryInfo(api.P_Int64, false),
		},
	}

	listEvents = api.ApiEndPoint{
		Handler:     api.ListEvents,
		QueryParams: api.QueryParams{},
	}

	modUserEntry = api.ApiEndPoint{
		Handler: api.ModUserEntry,
		QueryParams: api.QueryParams{
			"id":               api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"notes":            api.MkQueryInfo(api.P_True, false),
			"rating":           api.MkQueryInfo(api.P_Float64, false),
			"view-count":       api.MkQueryInfo(api.P_Int64, false),
			"current-position": api.MkQueryInfo(api.P_True, false),
			"status":           api.MkQueryInfo(api.P_UserStatus, false),
		},
	}

	setUserEntry = api.ApiEndPoint{
		Handler:     api.SetUserEntry,
		Method:      "POST",
		QueryParams: api.QueryParams{},
	}

	userEntries = api.ApiEndPoint{
		Handler: api.UserEntries,
	}

	getUserEntry = api.ApiEndPoint{
		Handler: api.GetUserEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
	}

	dropMedia = api.ApiEndPoint{
		Handler: api.DropMedia,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
	}

	resumeMedia = api.ApiEndPoint{
		Handler: api.ResumeMedia,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
	}

	pauseMedia = api.ApiEndPoint{
		Handler: api.PauseMedia,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
	}

	planMedia = api.ApiEndPoint{
		Handler: api.PlanMedia,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
	}

	beginMedia = api.ApiEndPoint{
		Handler: api.BeginMedia,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
	}
	// }}}

	// `/resource` endpoints {{{
	thumbResource = api.ApiEndPoint{
		Handler: api.ThumbnailResource,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
	}

	downloadThumb = api.ApiEndPoint{
		Handler: api.DownloadThumbnail,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
	}
) //}}}

func setupAIODir() string {
	dir, envExists := os.LookupEnv("AIO_DIR")
	if !envExists {
		dataDir, envExists := os.LookupEnv("XDG_DATA_HOME")
		if !envExists {
			home, envEenvExists := os.LookupEnv("HOME")
			if !envEenvExists {
				panic("Could not setup aio directory, $HOME does not exist")
			}
			dataDir = fmt.Sprintf("%s/.local/share", home)
		}
		dir = fmt.Sprintf("%s/aio-limas", dataDir)
	}

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		os.MkdirAll(dir, 0o755)
	} else if err != nil {
		panic(fmt.Sprintf("Could not create directory %s\n%s", dir, err.Error()))
	}
	return dir
}

func main() {
	aioPath := setupAIODir()
	os.Setenv("AIO_DIR", aioPath)

	dbPath := fmt.Sprintf("%s/all.db", aioPath)
	dbPathPtr := flag.String("db-path", dbPath, "Path to the database file")

	flag.Parse()

	db.InitDb(*dbPathPtr)

	type EndPointMap map[string]func(http.ResponseWriter, *http.Request)

	apiRoot := "/api/v1"

	// for db management type stuff
	makeEndpoints(apiRoot, EndPointMap{
		"add-entry":         addEntry.Listener,
		"mod-entry":         modEntry.Listener,
		"set-entry":         setEntry.Listener,
		"query":             searchApi.Listener,
		"list-entries":      listApi.Listener,
		"scan-folder":       api.ScanFolder,
		"stream-entry":      stream.Listener,
		"delete-entry":      deleteEntry.Listener,
		"list-collections":  listCollections.Listener,
		"list-copies":       listCopies.Listener,
		"list-descendants":  listDescendants.Listener,
		"total-cost":        totalCostOf.Listener,
		"list-tree":         getTree.Listener,
		"get-all-for-entry": getAllEntry.Listener,
	})

	makeEndpoints(apiRoot+"/type", EndPointMap{
		"format": api.ListFormats,
		"type":   api.ListTypes,
	})

	// for metadata stuff
	makeEndpoints(apiRoot+"/metadata", EndPointMap{
		"fetch":             fetchMetadataForEntry.Listener,
		"retrieve":          retrieveMetadataForEntry.Listener,
		"mod-entry":         modMetaEntry.Listener,
		"set-entry":         setMeta.Listener,
		"list-entries":      listMetadata.Listener,
		"identify":          identify.Listener,
		"finalize-identify": finalizeIdentify.Listener,
	})

	// for stuff relating to user viewing info
	// such as user rating, user beginning/ending a media, etc
	// stuff that would normally be managed by strack
	makeEndpoints(apiRoot+"/engagement", EndPointMap{
		"begin-media":    beginMedia.Listener,
		"finish-media":   finishEngagement.Listener,
		"plan-media":     planMedia.Listener,
		"drop-media":     dropMedia.Listener,
		"pause-media":    pauseMedia.Listener,
		"resume-media":   reassociate.Listener,
		"get-entry":      getUserEntry.Listener,
		"list-entries":   userEntries.Listener,
		"copy":           reassociate.Listener,
		"get-events":     getEvents.Listener,
		"delete-event":   deleteEvent.Listener,
		"register-event": registerEvent.Listener,
		"list-events":    listEvents.Listener,
		"mod-entry":      modUserEntry.Listener,
		"set-entry":      setUserEntry.Listener,
	})

	// For resources, such as entry thumbnails
	makeEndpoints(apiRoot+"/resource", EndPointMap{
		"thumbnail":          thumbResource.Listener,
		"download-thumbnail": downloadThumb.Listener,
	})

	http.HandleFunc("/", webservice.Root)

	http.ListenAndServe(":8080", nil)
}
