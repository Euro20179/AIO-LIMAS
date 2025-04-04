package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"aiolimas/accounts"
	api "aiolimas/api"
	lua_api "aiolimas/lua-api"
	"aiolimas/settings"
	"aiolimas/webservice"
	"aiolimas/webservice/dynamic"
)

func makeEndPointsFromList(root string, endPoints []api.ApiEndPoint) {
	// if the user sets this var, make all endpoints behind authorization
	for _, endPoint := range endPoints {
		http.HandleFunc(root+"/"+endPoint.EndPoint, endPoint.Listener)
	}
}

var ( // `/` endpoints {{{
	downloadDB = api.ApiEndPoint{
		Handler:     api.DownloadDB,
		Description: "Creates a copy of the database",
		EndPoint:    "download-db",
	}

	addTags = api.ApiEndPoint {
		Handler: api.AddTags,
		Description: "Adds tag(s) to an entry",
		EndPoint: "add-tags",
		QueryParams: api.QueryParams {
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"tags": api.MkQueryInfo(api.P_TList[string]("\x1F", func(in string) string{
				return in
			}), true),
		},
	}

	delTags = api.ApiEndPoint {
		Handler: api.DeleteTags,
		Description: "Delets tag(s) from an entry",
		EndPoint: "delete-tags",
		QueryParams: api.QueryParams {
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"tags": api.MkQueryInfo(api.P_TList[string]("\x1F", func(in string) string{
				return in
			}), true),
		},
	}

	addEntry = api.ApiEndPoint{
		Handler: api.AddEntry,
		QueryParams: api.QueryParams{
			"title":             api.MkQueryInfo(api.P_NotEmpty, true),
			"type":              api.MkQueryInfo(api.P_EntryType, true),
			"format":            api.MkQueryInfo(api.P_EntryFormat, true),
			"timezone":          api.MkQueryInfo(api.P_NotEmpty, false),
			"price":             api.MkQueryInfo(api.P_Float64, false),
			"is-digital":        api.MkQueryInfo(api.P_Bool, false),
			"is-anime":          api.MkQueryInfo(api.P_Bool, false),
			"art-style":         api.MkQueryInfo(api.P_ArtStyle, false),
			"libraryId":         api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"parentId":          api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"copyOf":            api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, false),
			"native-title":      api.MkQueryInfo(api.P_True, false),
			"tags":              api.MkQueryInfo(api.P_True, false),
			"location":          api.MkQueryInfo(api.P_True, false),
			"get-metadata":      api.MkQueryInfo(api.P_Bool, false),
			"metadata-provider": api.MkQueryInfo(api.P_MetaProvider, false),
			"user-rating":       api.MkQueryInfo(api.P_Float64, false),
			"user-status":       api.MkQueryInfo(api.P_UserStatus, false),
			"user-view-count":   api.MkQueryInfo(api.P_Int64, false),
			"user-notes":        api.MkQueryInfo(api.P_True, false),
		},
		Description: "Adds a new entry, and registers an Add event",
		Returns:     "InfoEntry",
		EndPoint:    "add-entry",
	}

	getTree = api.ApiEndPoint{
		EndPoint:     "list-tree",
		Handler:      api.GetTree,
		QueryParams:  api.QueryParams{},
		Description:  "Gets a tree-like json structure of all entries",
		Returns:      "InfoEntry",
		GuestAllowed: true,
	}

	modEntry = api.ApiEndPoint{
		EndPoint: "mod-entry",
		Handler:  api.ModEntry,
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
			// "is-anime":        api.MkQueryInfo(api.P_Bool, false),
			"art-style": api.MkQueryInfo(api.P_ArtStyle, false),
			"type":      api.MkQueryInfo(api.P_EntryType, false),
		},
		Description: "Modifies an individual entry datapoint",
	}

	setEntry = api.ApiEndPoint{
		EndPoint:    "set-entry",
		Handler:     api.SetEntry,
		QueryParams: api.QueryParams{},
		Method:      api.POST,
		Description: "Set an entry to the json of an entry<br>Post body must be updated entry",
	}

	listApi = api.ApiEndPoint{
		EndPoint: "list-entries",
		Handler:  api.ListEntries,
		QueryParams: api.QueryParams{
			"sort-by": api.MkQueryInfo(api.P_SqlSafe, false),
		},
		Description:  "List info entries",
		Returns:      "JSONL<InfoEntry>",
		GuestAllowed: true,
	}

	stream = api.ApiEndPoint{
		EndPoint: "stream-entry",
		Handler:  api.Stream,
		QueryParams: api.QueryParams{
			"id":      api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"subfile": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Download the file located by the {id}'s location",
		Returns:     "any",
	}

	deleteEntry = api.ApiEndPoint{
		EndPoint: "delete-entry",
		Handler:  api.DeleteEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description: "Deletes an entry",
	}

	listCollections = api.ApiEndPoint{
		EndPoint:     "list-collections",
		Handler:      api.ListCollections,
		QueryParams:  api.QueryParams{},
		Description:  "Lists en_title of all entries who's type is Collection",
		Returns:      "Sep<string, '\\n'>",
		GuestAllowed: true,
	}

	listLibraries = api.ApiEndPoint {
		EndPoint:     "list-libraries",
		Handler:      api.ListLibraries,
		QueryParams:  api.QueryParams{},
		Description:  "Lists ids of all entries who's type is Library",
		Returns:      "Sep<string, '\\n'>",
		GuestAllowed: true,
	}

	listCopies = api.ApiEndPoint{
		EndPoint: "list-copies",
		Handler:  api.GetCopies,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Lists copies of an entry",
		Returns:      "JSONL<InfoEntry>",
		GuestAllowed: true,
	}

	listDescendants = api.ApiEndPoint{
		EndPoint: "list-descendants",
		Handler:  api.GetDescendants,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Lists children of an entry",
		Returns:      "JSONL<InfoEntry>",
		GuestAllowed: true,
	}

	totalCostOf = api.ApiEndPoint{
		EndPoint: "total-cost",
		Handler:  api.TotalCostOf,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Gets the total cost of an entry, summing itself + children",
		Returns:      "float",
		GuestAllowed: true,
	}

	search3Api = api.ApiEndPoint{
		EndPoint: "query-v3",
		Handler:  api.QueryEntries3,
		QueryParams: api.QueryParams{
			"search": api.MkQueryInfo(api.P_NotEmpty, true),
		},
		Returns:      "InfoEntry[]",
		Description:  "search query similar to how sql where query works",
		GuestAllowed: true,
	}

	getAllEntry = api.ApiEndPoint{
		EndPoint: "get-all-for-entry",
		Handler:  api.GetAllForEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Gets the userEntry, metadataEntry, and infoEntry for an entry",
		Returns:      "UserEntry\\nMetadataEntry\\nInfoEntry",
		GuestAllowed: true,
	}
) // }}}

var ( // `/metadata` endpoints {{{
	identify = api.ApiEndPoint{
		EndPoint: "identify",
		Handler:  api.IdentifyWithSearch,
		QueryParams: api.QueryParams{
			"title":    api.MkQueryInfo(api.P_NotEmpty, true),
			"provider": api.MkQueryInfo(api.P_Identifier, true),
		},
		Description: `List metadata results based on a search query + provider<br>
The id of the metadata entry will be the id that's supposed to be given to <code>identified-id</code><br>
when using finalize-identify`,
		Returns: "JSONL<MetadataEntry>",
	}

	finalizeIdentify = api.ApiEndPoint{
		EndPoint: "finalize-identify",
		Handler:  api.FinalizeIdentification,
		QueryParams: api.QueryParams{
			"identified-id": api.MkQueryInfo(api.P_NotEmpty, true),
			"provider":      api.MkQueryInfo(api.P_IdIdentifier, true),
			"apply-to":      api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
		Description: "Apply an identified id from /identify, to an entry using a provider",
		Returns:     "none",
	}

	setMeta = api.ApiEndPoint{
		EndPoint:    "set-entry",
		Handler:     api.SetMetadataEntry,
		Method:      "POST",
		QueryParams: api.QueryParams{},
		Description: "Set a metadata entry to the json of an entry<br>post body must be updated metadata entry",
		Returns:     "UserEntry",
	}

	fetchMetadataForEntry = api.ApiEndPoint{
		EndPoint: "fetch",
		Handler:  api.FetchMetadataForEntry,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"provider": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: `Fetch the metadata for an entry based on the type<br>
	and using EntryInfo.En_Title as the title search<br>
	if provider is not given, it is automatically chosen based on type`,
	}

	retrieveMetadataForEntry = api.ApiEndPoint{
		EndPoint: "retrieve",
		Handler:  api.RetrieveMetadataForEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
		Description:  "Gets the metadata for an entry",
		Returns:      "MetadataEntry",
		GuestAllowed: true,
	}

	modMetaEntry = api.ApiEndPoint{
		EndPoint: "mod-entry",
		Handler:  api.ModMetadataEntry,
		QueryParams: api.QueryParams{
			"id":              api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
			"rating":          api.MkQueryInfo(api.P_Float64, false),
			"description":     api.MkQueryInfo(api.P_NotEmpty, false),
			"release-year":    api.MkQueryInfo(api.P_Int64, false),
			"thumbnail":       api.MkQueryInfo(api.P_NotEmpty, false),
			"media-dependant": api.MkQueryInfo(api.P_NotEmpty, false),
			"datapoints":      api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Modify metadata by datapoint",
	}

	listMetadata = api.ApiEndPoint{
		EndPoint:     "list-entries",
		Handler:      api.ListMetadata,
		QueryParams:  api.QueryParams{},
		Description:  "Lists all metadata entries",
		Returns:      "JSONL<MetadataEntry>",
		GuestAllowed: true,
	}
) // }}}

var ( // `/engagement` endpoints {{{
	finishEngagement = api.ApiEndPoint{
		EndPoint: "finish-media",
		Handler:  api.FinishMedia,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"rating":   api.MkQueryInfo(api.P_Float64, true),
			"timezone": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Finishes a media, and registers a Finish event",
	}

	reassociate = api.ApiEndPoint{
		EndPoint: "copy",
		Handler:  api.CopyUserViewingEntry,
		QueryParams: api.QueryParams{
			"src-id":  api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"dest-id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description: "Moves all user entry data, and events from one entry entry to another",
	}

	getEvents = api.ApiEndPoint{
		EndPoint: "get-events",
		Handler:  api.GetEventsOf,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Lists the events of an entry",
		Returns:      "JSONL<EventEntry>",
		GuestAllowed: true,
	}
	deleteEvent = api.ApiEndPoint{
		EndPoint: "delete-event",
		Handler:  api.DeleteEvent,
		QueryParams: api.QueryParams{
			"id":        api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"timestamp": api.MkQueryInfo(api.P_Int64, true),
			"after":     api.MkQueryInfo(api.P_Int64, true),
		},
		Description: "Deletes an event from an entry",
	}

	registerEvent = api.ApiEndPoint{
		EndPoint: "register-event",
		Handler:  api.RegisterEvent,
		QueryParams: api.QueryParams{
			"id":        api.MkQueryInfo(api.P_VerifyIdAndGetInfoEntry, true),
			"name":      api.MkQueryInfo(api.P_NotEmpty, true),
			"timestamp": api.MkQueryInfo(api.P_Int64, false),
			"after":     api.MkQueryInfo(api.P_Int64, false),
			"timezone":  api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Registers an event for an entry",
	}

	listEvents = api.ApiEndPoint{
		EndPoint:     "list-events",
		Handler:      api.ListEvents,
		QueryParams:  api.QueryParams{},
		Description:  "Lists all events associated with an entry",
		Returns:      "JSONL<EventEntry>",
		GuestAllowed: true,
	}

	modUserEntry = api.ApiEndPoint{
		EndPoint: "mod-entry",
		Handler:  api.ModUserEntry,
		QueryParams: api.QueryParams{
			"id":               api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"notes":            api.MkQueryInfo(api.P_True, false),
			"rating":           api.MkQueryInfo(api.P_Float64, false),
			"view-count":       api.MkQueryInfo(api.P_Int64, false),
			"current-position": api.MkQueryInfo(api.P_True, false),
			"status":           api.MkQueryInfo(api.P_UserStatus, false),
		},
		Description: "Modifies datapoints of a user entry",
	}

	setUserEntry = api.ApiEndPoint{
		EndPoint:    "set-entry",
		Handler:     api.SetUserEntry,
		Method:      "POST",
		QueryParams: api.QueryParams{},
		Description: "Updates the user entry with the post body<br>Post body must be updated user entry",
	}

	userEntries = api.ApiEndPoint{
		EndPoint:     "list-entries",
		Handler:      api.UserEntries,
		Description:  "Lists all user entries",
		Returns:      "JSONL<UserEntry>",
		GuestAllowed: true,
	}

	getUserEntry = api.ApiEndPoint{
		EndPoint: "get-entry",
		Handler:  api.GetUserEntry,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
		},
		Description:  "Gets a user entry by id",
		Returns:      "UserEntry",
		GuestAllowed: true,
	}

	dropMedia = api.ApiEndPoint{
		EndPoint: "drop-media",
		Handler:  api.DropMedia,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"timezone": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Drops a media, and registers a Drop event",
	}

	resumeMedia = api.ApiEndPoint{
		EndPoint: "resume-media",
		Handler:  api.ResumeMedia,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"timezone": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Resumes a media and registers a ReViewing event",
	}

	pauseMedia = api.ApiEndPoint{
		EndPoint: "pause-media",
		Handler:  api.PauseMedia,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"timezone": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Pauses a media and registers a Pause event",
	}

	planMedia = api.ApiEndPoint{
		EndPoint: "plan-media",
		Handler:  api.PlanMedia,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"timezone": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Plans a media and registers a Plan event",
	}

	beginMedia = api.ApiEndPoint{
		EndPoint: "begin-media",
		Handler:  api.BeginMedia,
		QueryParams: api.QueryParams{
			"id":       api.MkQueryInfo(api.P_VerifyIdAndGetUserEntry, true),
			"timezone": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Begins a media and registers a Begin event",
	}
	// }}}

	// `/account` endpoints {{{
	createAccount = api.ApiEndPoint{
		EndPoint: "create",
		Handler:  api.CreateAccount,
		QueryParams: api.QueryParams{
			"username": api.MkQueryInfo(api.P_NotEmpty, true),
			"password": api.MkQueryInfo(api.P_NotEmpty, true),
		},
		Description: "Creates an account",
		UserIndependant: true,
		GuestAllowed: true,
	}

	accountLogin = api.ApiEndPoint {
		EndPoint: "login",
		Handler: api.Login,
		QueryParams: api.QueryParams{
			"username": api.MkQueryInfo(api.P_NotEmpty, false),
			"password": api.MkQueryInfo(api.P_NotEmpty, false),
		},
		Description: "Login",
		UserIndependant: true,
		GuestAllowed: true,
	}

	accountList = api.ApiEndPoint {
		EndPoint: "list",
		Handler: api.ListUsers,
		Description: "List all users",
		UserIndependant: true,
		GuestAllowed:  true,
	}
	//}}}

	// `/resource` endpoints {{{
	thumbResource = api.ApiEndPoint{
		EndPoint: "get-thumbnail",
		Handler:  api.ThumbnailResource,
		QueryParams: api.QueryParams{
			"hash": api.MkQueryInfo(api.P_NotEmpty, true),
		},
		Description:  "Gets the thumbnail for an id (if it can find the thumbnail in the thumbnails dir)",
		GuestAllowed: true,
		UserIndependant: true,
	}

	//this is the legacy one, since the url is hardcoded I can't really change it.
	thumbResourceLegacy = api.ApiEndPoint{
		EndPoint: "thumbnail",
		Handler:  api.ThumbnailResourceLegacy,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_NotEmpty, true),
		},
		Description:  "LEGACY, Gets the thumbnail for an id (if it can find the thumbnail in the thumbnails dir)",
		GuestAllowed: true,
		UserIndependant: true,
	}

	downloadThumb = api.ApiEndPoint{
		EndPoint: "download-thumbnail",
		Handler:  api.DownloadThumbnail,
		QueryParams: api.QueryParams{
			"id": api.MkQueryInfo(api.P_VerifyIdAndGetMetaEntry, true),
		},
		Description: "If the id has a remote thumbnail, download it, does not update metadata",
	}
	//}}}

	// `/type` endpoints {{{
	formatTypesApi = api.ApiEndPoint{
		EndPoint:        "format",
		Handler:         api.ListFormats,
		Description:     "Lists the valid values for a Format",
		GuestAllowed:    true,
		UserIndependant: true,
	}

	typeTypesApi = api.ApiEndPoint{
		EndPoint:        "type",
		Handler:         api.ListTypes,
		Description:     "Lists the types for a Type",
		GuestAllowed:    true,
		UserIndependant: true,
	}

	artStylesApi = api.ApiEndPoint{
		EndPoint:        "artstyle",
		Handler:         api.ListArtStyles,
		Description:     "Lists the types art styles",
		GuestAllowed:    true,
		UserIndependant: true,
	}
	//}}}

	// `/docs` endpoints {{{
	mainDocs = api.ApiEndPoint{
		EndPoint:        "",
		Handler:         DocHTML,
		Description:     "The documentation",
		GuestAllowed:    true,
		UserIndependant: true,
	}
	//}}}
)

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

type EndPointMap map[string]func(http.ResponseWriter, *http.Request)

var (
	mainEndpointList = []api.ApiEndPoint{
		addEntry,
		modEntry,
		setEntry,
		search3Api,
		listApi,
		stream,
		deleteEntry,
		listCollections,
		listLibraries,
		listCopies,
		listDescendants,
		totalCostOf,
		getTree,
		getAllEntry,
		downloadDB,
		addTags,
		delTags,
	}

	metadataEndpointList = []api.ApiEndPoint{
		fetchMetadataForEntry,
		retrieveMetadataForEntry,
		modMetaEntry,
		setMeta,
		listMetadata,
		identify,
		finalizeIdentify,
	}

	// for stuff relating to user viewing info
	// such as user rating, user beginning/ending a media, etc
	// stuff that would normally be managed by strack
	engagementEndpointList = []api.ApiEndPoint{
		beginMedia,
		finishEngagement,
		planMedia,
		dropMedia,
		pauseMedia,
		reassociate,
		getUserEntry,
		userEntries,
		resumeMedia,
		getEvents,
		deleteEvent,
		registerEvent,
		listEvents,
		modUserEntry,
		setUserEntry,
	}

	accountEndPoints = []api.ApiEndPoint{
		createAccount,
		accountLogin,
		accountList,
	}

	resourceEndpointList = []api.ApiEndPoint{
		thumbResource,
		thumbResourceLegacy,
		downloadThumb,
	}

	typeEndpoints = []api.ApiEndPoint{
		formatTypesApi,
		typeTypesApi,
		artStylesApi,
	}

	docsEndpoints = []api.ApiEndPoint{
		mainDocs,
	}

	endPointLists = [][]api.ApiEndPoint{
		mainEndpointList,
		metadataEndpointList,
		engagementEndpointList,
		typeEndpoints,
		resourceEndpointList,
	}
)

func startServer() {
	const apiRoot = "/api/v1"

	// for db management type stuff
	makeEndPointsFromList(apiRoot, mainEndpointList)
	makeEndPointsFromList(apiRoot+"/type", typeEndpoints)
	makeEndPointsFromList(apiRoot+"/metadata", metadataEndpointList)
	makeEndPointsFromList(apiRoot+"/engagement", engagementEndpointList)
	// For resources, such as entry thumbnails
	makeEndPointsFromList(apiRoot+"/resource", resourceEndpointList)

	makeEndPointsFromList("/docs", docsEndpoints)

	makeEndPointsFromList("/account", accountEndPoints)

	// htmlEndpoint := api.ApiEndPoint{
	// 	EndPoint:     "html",
	// 	Handler:      dynamic.HtmlEndpoint,
	// 	Description:  "Dynamic html endpoints",
	// 	GuestAllowed: true,
	// }
	http.HandleFunc("/html/", dynamic.HtmlEndpoint)
	http.HandleFunc("/", webservice.Root)

	port := os.Getenv("AIO_PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func setEnvOrPanic(name string, val string) {
	if err := os.Setenv(name, val); err != nil {
		panic(err.Error())
	}
}

func initConfig(aioPath string) {
	configPath := aioPath + "/config.json"
	setEnvOrPanic("AIO_CONFIG_FILE", configPath)
	if _, err := os.Stat(configPath); err == nil {
		return
	}
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic("Failed to create config file")
	}
	file.Write([]byte("{}"))
	if err := file.Close(); err != nil {
		panic("Failed to create config file, writing {}")
	}
}

func main() {
	aioPath := setupAIODir()
	setEnvOrPanic("AIO_DIR", aioPath)

	initConfig(aioPath)

	accounts.InitAccountsDb(aioPath)

	settings.InitSettingsManager(aioPath)

	flag.Parse()

	inst, err := lua_api.InitGlobalLuaInstance("./lua-extensions/init.lua")
	if err != nil {
		panic("Could not initialize global lua instance")
	}
	lua_api.GlobalLuaInstance = inst

	startServer()
}

func DocHTML(ctx api.RequestContext) {
	w := ctx.W

	html := ""
	for _, list := range endPointLists {
		for _, endP := range list {
			html += endP.GenerateDocHTML()
		}
	}
	w.WriteHeader(200)
	w.Write([]byte(html))
}
