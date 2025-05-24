package api

import (
	"fmt"
	"net/http"
)

func MakeEndPointsFromList(root string, endPoints []ApiEndPoint) {
	// if the user sets this var, make all endpoints behind authorization
	for _, endPoint := range endPoints {
		http.HandleFunc(root+"/"+endPoint.EndPoint, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			mthd := string(endPoint.Method)
			if mthd == "" {
				mthd = "GET"
			}
			w.Header().Set("Access-Control-Allow-Methods", mthd)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization")
			endPoint.Listener(w, r)
		})
	}
}

// `/` endpoints {{{
var mainEndpointList = []ApiEndPoint{
	{
		Handler:     DownloadDB,
		Description: "Creates a copy of the database",
		EndPoint:    "download-db",
	},

	{
		Handler:     AddTags,
		Description: "Adds tag(s) to an entry, tags must be an \\x1F (ascii unit separator) separated list",
		EndPoint:    "add-tags",
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"tags": MkQueryInfo(P_TList("\x1F", func(in string) string {
				return in
			}), true),
		},
	},

	{
		Handler:     DeleteTags,
		Description: "Delets tag(s) from an entry, tags must be an \\x1F (ascii unit separator) separated list",
		EndPoint:    "delete-tags",
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"tags": MkQueryInfo(P_TList("\x1F", func(in string) string {
				return in
			}), true),
		},
	},

	{
		Handler: AddEntry,
		QueryParams: QueryParams{
			"title":             MkQueryInfo(P_NotEmpty, true),
			"type":              MkQueryInfo(P_EntryType, true),
			"format":            MkQueryInfo(P_EntryFormat, true),
			"timezone":          MkQueryInfo(P_NotEmpty, false),
			"price":             MkQueryInfo(P_Float64, false),
			"is-digital":        MkQueryInfo(P_Bool, false),
			"is-anime":          MkQueryInfo(P_Bool, false),
			"art-style":         MkQueryInfo(P_ArtStyle, false),
			"libraryId":         MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
			"parentId":          MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
			"copyOf":            MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
			"native-title":      MkQueryInfo(P_True, false),
			"tags":              MkQueryInfo(P_True, false),
			"location":          MkQueryInfo(P_True, false),
			"get-metadata":      MkQueryInfo(P_Bool, false),
			"metadata-provider": MkQueryInfo(P_MetaProvider, false),
			"user-rating":       MkQueryInfo(P_Float64, false),
			"user-status":       MkQueryInfo(P_UserStatus, false),
			"user-view-count":   MkQueryInfo(P_Int64, false),
			"user-notes":        MkQueryInfo(P_True, false),
		},
		Description: "Adds a new entry, and registers an Add event",
		Returns:     "InfoEntry",
		EndPoint:    "add-entry",
	},

	{
		EndPoint:     "list-tree",
		Handler:      GetTree,
		QueryParams:  QueryParams{},
		Description:  "Gets a tree-like json structure of all entries",
		Returns:      "InfoEntry",
		GuestAllowed: true,
	},

	{
		EndPoint: "mod-entry",
		Handler:  ModEntry,
		QueryParams: QueryParams{
			"id":              MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"en-title":        MkQueryInfo(P_NotEmpty, false),
			"native-title":    MkQueryInfo(P_True, false),
			"format":          MkQueryInfo(P_EntryFormat, false),
			"parent-id":       MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
			"become-orphan":   MkQueryInfo(P_Bool, false),
			"become-original": MkQueryInfo(P_Bool, false),
			"copy-id":         MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
			"price":           MkQueryInfo(P_Float64, false),
			"location":        MkQueryInfo(P_True, false),
			"tags":            MkQueryInfo(P_True, false),
			// "is-anime":        MkQueryInfo(P_Bool, false),
			"art-style": MkQueryInfo(P_ArtStyle, false),
			"type":      MkQueryInfo(P_EntryType, false),
		},
		Description: "Modifies an individual entry datapoint",
	},

	{
		EndPoint:    "set-entry",
		Handler:     SetEntry,
		QueryParams: QueryParams{},
		Method:      POST,
		Description: "Set an entry to the json of an entry<br>Post body must be updated entry",
	},

	{
		EndPoint: "list-entries",
		Handler:  ListEntries,
		QueryParams: QueryParams{
			"sort-by": MkQueryInfo(P_SqlSafe, false),
		},
		Description:  "List info entries",
		Returns:      "JSONL<InfoEntry>",
		GuestAllowed: true,
	},

	{
		EndPoint: "stream-entry",
		Handler:  Stream,
		QueryParams: QueryParams{
			"id":      MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"subfile": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Download the file located by the {id}'s location",
		Returns:     "any",
	},

	{
		EndPoint: "delete-entry",
		Handler:  DeleteEntry,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description: "Deletes an entry",
	},

	{
		EndPoint:     "list-collections",
		Handler:      ListCollections,
		QueryParams:  QueryParams{},
		Description:  "Lists en_title of all entries who's type is Collection",
		Returns:      "Sep<string, '\\n'>",
		GuestAllowed: true,
	},

	{
		EndPoint:     "list-libraries",
		Handler:      ListLibraries,
		QueryParams:  QueryParams{},
		Description:  "Lists ids of all entries who's type is Library",
		Returns:      "Sep<string, '\\n'>",
		GuestAllowed: true,
	},

	{
		EndPoint: "list-copies",
		Handler:  GetCopies,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Lists copies of an entry",
		Returns:      "JSONL<InfoEntry>",
		GuestAllowed: true,
	},

	{
		EndPoint: "list-descendants",
		Handler:  GetDescendants,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Lists children of an entry",
		Returns:      "JSONL<InfoEntry>",
		GuestAllowed: true,
	},

	{
		EndPoint: "total-cost",
		Handler:  TotalCostOf,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Gets the total cost of an entry, summing itself + children",
		Returns:      "float",
		GuestAllowed: true,
	},

	{
		EndPoint: "query-v3",
		Handler:  QueryEntries3,
		QueryParams: QueryParams{
			"search": MkQueryInfo(P_NotEmpty, true),
		},
		Returns:      "InfoEntry[]",
		Description:  "search query similar to how sql where query works",
		GuestAllowed: true,
	},

	{
		EndPoint: "get-all-for-entry",
		Handler:  GetAllForEntry,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Gets the userEntry, metadataEntry, and infoEntry for an entry",
		Returns:      "UserEntry\\nMetadataEntry\\nInfoEntry",
		GuestAllowed: true,
	},
} // }}}

// `/metadata` endpoints {{{
var metadataEndpointList = []ApiEndPoint{
	{
		EndPoint: "fetch-location",
		Handler: FetchLocation,
		QueryParams: QueryParams {
			"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
			"provider": MkQueryInfo(P_LocationProvider, false),
			"provider-id": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Fetch the location of an entry based on the metadata and other info",
	},
	{
		EndPoint: "identify",
		Handler:  IdentifyWithSearch,
		QueryParams: QueryParams{
			"title":    MkQueryInfo(P_NotEmpty, true),
			"provider": MkQueryInfo(P_Identifier, true),
		},
		Description: `List metadata results based on a search query + provider<br>
The id of the metadata entry will be the id that's supposed to be given to <code>identified-id</code><br>
when using finalize-identify`,
		Returns: "JSONL<MetadataEntry>",
	},

	{
		EndPoint: "finalize-identify",
		Handler:  FinalizeIdentification,
		QueryParams: QueryParams{
			"identified-id": MkQueryInfo(P_NotEmpty, true),
			"provider":      MkQueryInfo(P_IdIdentifier, true),
			"apply-to":      MkQueryInfo(P_VerifyIdAndGetMetaEntry, false),
		},
		Description: "Apply an identified id from /identify, to an entry using a provider",
		Returns:     "none",
	},

	{
		EndPoint:    "set-entry",
		Handler:     SetMetadataEntry,
		Method:      "POST",
		QueryParams: QueryParams{},
		Description: "Set a metadata entry to the json of an entry<br>post body must be updated metadata entry",
		Returns:     "UserEntry",
	},

	{
		EndPoint: "fetch",
		Handler:  FetchMetadataForEntry,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"provider": MkQueryInfo(P_NotEmpty, false),
		},
		Returns: "MetadataEntry",
		Description: `Fetch the metadata for an entry based on the type<br>
	and using EntryInfo.En_Title as the title search<br>
	if provider is not given, it is automatically chosen based on type`,
	},

	{
		EndPoint: "retrieve",
		Handler:  RetrieveMetadataForEntry,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
		},
		Description:  "Gets the metadata for an entry",
		Returns:      "MetadataEntry",
		GuestAllowed: true,
	},

	{
		EndPoint: "mod-entry",
		Handler:  ModMetadataEntry,
		QueryParams: QueryParams{
			"id":              MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
			"rating":          MkQueryInfo(P_Float64, false),
			"description":     MkQueryInfo(P_NotEmpty, false),
			"release-year":    MkQueryInfo(P_Int64, false),
			"thumbnail":       MkQueryInfo(P_NotEmpty, false),
			"media-dependant": MkQueryInfo(P_NotEmpty, false),
			"datapoints":      MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Modify metadata by datapoint",
	},

	{
		EndPoint: "set-thumbnail",
		Handler: SetThumbnail,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
		},
		Description: "Set the thumbnail for a metadata entry",
		Method: POST,
	},

	{
		EndPoint:     "list-entries",
		Handler:      ListMetadata,
		QueryParams:  QueryParams{},
		Description:  "Lists all metadata entries",
		Returns:      "JSONL<MetadataEntry>",
		GuestAllowed: true,
	},
} // }}}

// `/engagement` endpoints {{{
var engagementEndpointList = []ApiEndPoint{
	{
		EndPoint: "finish-media",
		Handler:  FinishMedia,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"rating":   MkQueryInfo(P_Float64, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Finishes a media, and registers a Finish event",
	},

	{
		EndPoint: "copy",
		Handler:  CopyUserViewingEntry,
		QueryParams: QueryParams{
			"src-id":  MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"dest-id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description: "Moves all user entry data, and events from one entry entry to another",
	},

	{
		EndPoint: "get-events",
		Handler:  GetEventsOf,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description:  "Lists the events of an entry",
		Returns:      "JSONL<EventEntry>",
		GuestAllowed: true,
	},
	{
		EndPoint: "delete-event",
		Handler:  DeleteEvent,
		QueryParams: QueryParams{
			"id":        MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"timestamp": MkQueryInfo(P_Int64, true),
			"after":     MkQueryInfo(P_Int64, true),
			"before": MkQueryInfo(P_Int64, true),
		},
		Description: "Deletes an event from an entry",
	},

	{
		EndPoint: "register-event",
		Handler:  RegisterEvent,
		QueryParams: QueryParams{
			"id":        MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
			"name":      MkQueryInfo(P_NotEmpty, true),
			"timestamp": MkQueryInfo(P_Int64, false),
			"after":     MkQueryInfo(P_Int64, false),
			"timezone":  MkQueryInfo(P_NotEmpty, false),
			"before":    MkQueryInfo(P_Int64, false),
		},
		Description: "Registers an event for an entry",
	},

	{
		EndPoint:     "list-events",
		Handler:      ListEvents,
		QueryParams:  QueryParams{},
		Description:  "Lists all events associated with an entry",
		Returns:      "JSONL<EventEntry>",
		GuestAllowed: true,
	},

	{
		EndPoint: "mod-entry",
		Handler:  ModUserEntry,
		QueryParams: QueryParams{
			"id":               MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"notes":            MkQueryInfo(P_True, false),
			"rating":           MkQueryInfo(P_Float64, false),
			"view-count":       MkQueryInfo(P_Int64, false),
			"current-position": MkQueryInfo(P_True, false),
			"status":           MkQueryInfo(P_UserStatus, false),
		},
		Description: "Modifies datapoints of a user entry",
	},

	{
		EndPoint:    "set-entry",
		Handler:     SetUserEntry,
		Method:      "POST",
		QueryParams: QueryParams{},
		Description: "Updates the user entry with the post body<br>Post body must be updated user entry",
	},

	{
		EndPoint:     "list-entries",
		Handler:      UserEntries,
		Description:  "Lists all user entries",
		Returns:      "JSONL<UserEntry>",
		GuestAllowed: true,
	},

	{
		EndPoint: "get-entry",
		Handler:  GetUserEntry,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
		},
		Description:  "Gets a user entry by id",
		Returns:      "UserEntry",
		GuestAllowed: true,
	},

	{
		EndPoint: "drop-media",
		Handler:  DropMedia,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Drops a media, and registers a Drop event",
	},

	{
		EndPoint: "resume-media",
		Handler:  ResumeMedia,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Resumes a media and registers a ReViewing event",
	},

	{
		EndPoint: "pause-media",
		Handler:  PauseMedia,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Pauses a media and registers a Pause event",
	},

	{
		EndPoint: "plan-media",
		Handler:  PlanMedia,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Plans a media and registers a Plan event",
	},

	{
		EndPoint: "begin-media",
		Handler:  BeginMedia,
		QueryParams: QueryParams{
			"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Begins a media and registers a Viewing event",
	},
	{
		EndPoint: "wait-media",
		Handler: WaitMedia,
		QueryParams: QueryParams {
			"id": MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
			"timezone": MkQueryInfo(P_NotEmpty, false),
		},
		Description: "Sets the status to waiting, and registers a Waiting event",
	},
} //}}}

// `/account` endpoints {{{
var AccountEndPoints = []ApiEndPoint{
	{
		EndPoint:        "create",
		Handler:         CreateAccount,
		Method:          POST,
		Description:     "Creates an account",
		UserIndependant: true,
		GuestAllowed:    true,
	},

	{
		EndPoint: "username2id",
		Handler: Username2Id,
		Description: "get a user's id from username",
		QueryParams: QueryParams {
			"username": MkQueryInfo(P_NotEmpty, true),
		},
		UserIndependant: true,
		GuestAllowed: true,
	},

	{
		EndPoint: "login",
		Handler:  Login,
		QueryParams: QueryParams{
			"username": MkQueryInfo(P_NotEmpty, false),
			"password": MkQueryInfo(P_NotEmpty, false),
		},
		Description:     "Login",
		UserIndependant: true,
		GuestAllowed:    true,
	},

	{
		EndPoint:        "list",
		Handler:         ListUsers,
		Description:     "List all users",
		UserIndependant: true,
		GuestAllowed:    true,
	},

	{
		EndPoint:    "delete",
		Method:      DELETE,
		Description: "Delete an account",
		Handler:     DeleteAccount,
	},
} // }}}

// `/resource` endpoints {{{
var resourceEndpointList = []ApiEndPoint{
	{
		EndPoint: "get-thumbnail",
		Handler:  ThumbnailResource,
		QueryParams: QueryParams{
			"hash": MkQueryInfo(P_NotEmpty, true),
		},
		Description:     "Gets the thumbnail for an id (if it can find the thumbnail in the thumbnails dir)",
		GuestAllowed:    true,
		UserIndependant: true,
	},

	// this is the legacy one, since the url is hardcoded I can't really change it.
	{
		EndPoint: "thumbnail",
		Handler:  ThumbnailResourceLegacy,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_NotEmpty, true),
		},
		Description:     "LEGACY, Gets the thumbnail for an id (if it can find the thumbnail in the thumbnails dir)",
		GuestAllowed:    true,
		UserIndependant: true,
	},

	{
		EndPoint: "download-thumbnail",
		Handler:  DownloadThumbnail,
		QueryParams: QueryParams{
			"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
		},
		Description: "If the id has a remote thumbnail, download it, does not update metadata",
	},
} // }}}

// `/type` endpoints {{{
var typeEndpoints = []ApiEndPoint{
	{
		EndPoint:        "format",
		Handler:         ListFormats,
		Description:     "Lists the valid values for a Format",
		GuestAllowed:    true,
		UserIndependant: true,
	},

	{
		EndPoint:        "type",
		Handler:         ListTypes,
		Description:     "Lists the types for a Type",
		GuestAllowed:    true,
		UserIndependant: true,
	},

	{
		EndPoint:        "artstyle",
		Handler:         ListArtStyles,
		Description:     "Lists the types art styles",
		GuestAllowed:    true,
		UserIndependant: true,
	},
} // }}}

// `/docs` endpoints {{{
var MainDocs = ApiEndPoint{
	EndPoint:        "",
	Handler:         DocHTML,
	Description:     "The documentation",
	GuestAllowed:    true,
	UserIndependant: true,
} // }}}

var Endpoints = map[string][]ApiEndPoint{
	"":            mainEndpointList,
	"/metadata":   metadataEndpointList,
	"/engagement": engagementEndpointList,
	"/type":       typeEndpoints,
	"/resource":   resourceEndpointList,
}

// this way the html at least wont change until a server restart
var htmlCache []byte

func DocHTML(ctx RequestContext) {
	w := ctx.W

	if len(htmlCache) == 0 {
		html := "<style>.required::after { content: \"(required) \"; font-weight: bold; }</style>"
		tableOfContents := "<p>Table of contents</p><ul>"
		docsHTML := ""
		for root, list := range Endpoints {
			if root != "" {
				tableOfContents += fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", root, root)
				docsHTML += fmt.Sprintf("<HR><h1 id=\"%s\">%s</h1>", root, root)
			} else {
				tableOfContents += fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", "/", "/")
				docsHTML += fmt.Sprintf("<HR><h1 id=\"%s\">%s</h1>", "/", "/")
			}
			for _, endP := range list {
				docsHTML += endP.GenerateDocHTML(root)
			}
		}
		html += tableOfContents + "</ul>" + docsHTML
		htmlCache = []byte(html)
	}
	w.WriteHeader(200)
	w.Write(htmlCache)
}
