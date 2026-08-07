package api

import (
	"fmt"
	"net/http"
	"text/template"

	"aiolimas/util"
)

func MakeEndPointsFromList(root string, endPoints []ApiEndPoint) {
	// if the user sets this var, make all endpoints behind authorization
	for _, endPoint := range endPoints {
		names := []string{endPoint.EndPoint}
		names = append(names, endPoint.Aliases...)
		for _, name := range names {
			http.HandleFunc(root + "/" + name, func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "*")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization")
				endPoint.Listener(w, r)
			})
		}
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
		Methods: map[string]MethodSpec{
			"POST": {},
		},
		Handler: HookRadarr,
		Description: "Handles radar webhook requests",
		Returns: "InfoEntry",
		EndPoint: "hook-radarr",
	},


	{
		EndPoint: "query-v4",
		Handler: QueryEntries4,
		Methods: map[string]MethodSpec{
			"GET": {
				Params: QueryParams{
					"search": MkQueryInfo(P_NotEmpty, true),
					"order-by": MkQueryInfo(P_SqlSafe, false),
				},
				GuestAllowed: true,
				UserIndependant: true,
			},
		},
		Description: "Search with a plain title search",
		Returns: "InfoEntry[]",
		Deprecated: "Use QUERY /entry/ instead",
	},

	{
		EndPoint: "query-v3",
		Handler:  QueryEntries3,
		Methods: map[string]MethodSpec{
			"GET": {
				Params: QueryParams{
					"search":   MkQueryInfo(P_NotEmpty, true),
					"order-by": MkQueryInfo(P_SqlSafe, false),
				},
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
		Returns:         "InfoEntry[]",
		Description:     "search query similar to how sql where query works",
		Deprecated: "Use QUERY /entry/?v=3 instead",
	},

	{
		EndPoint: "get-all-for-entries",
		Handler:  GetAllForEntries,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"ids": MkQueryInfo(P_TList(",", func(in string) string { return in }), true),
				},
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
		Description:     "Gets the userEntry, metadataEntry, and infoEntry for a , separated list of entries",
		Returns:         "Same as get-all-for-entry, each item is separated by \\n\\n",
	},

	// /entry {{{
	{
		EndPoint: "entry",
		Handler: EntryResource,
		Methods: map[string]MethodSpec {
			"TREE": {
				Description: "Generate a tree representation of all entries",
				Returns: "Record<string, {EntryInfo: InfoEntry, MetaInfo: Metadata, UserInfo: UserEntry, Children: string[], Copies: string[]}>",
				GuestAllowed: true,
				UserIndependant: true,
			},
			"GET": {
				Description: `
					Get various information about all entries <BR>
					Values for ?kind
					<dl>
						<dt> relations
						<dd> lists relations of all entries
						<dt> events
						<dd> lists all events
						<dt> transactions
						<dd> lists all transactions
						<dt> info
						<dd> lists all info items (can use ?sort-by)
						<dt> meta
						<dd> lists all metadata items
						<dt> user
						<dd> lists all user viewing items
					</dl>
				`,
				Params: QueryParams {
					"kind": MkQueryInfo(P_NotEmpty, true),
					"sort-by": MkQueryInfo(P_SqlSafe, false),
				},
				UserIndependant: true,
				GuestAllowed: true,
			},
			"QUERY": {
				Description: "Do a search. ?v specifies the search version, can be 3 or 4.",
				Returns: "JSONL<InfoEntry>",
				Params: QueryParams {
					"q": MkQueryInfo(P_NotEmpty, true),
					"order-by": MkQueryInfo(P_SqlSafe, false),
					"v": MkQueryInfo(P_Int64, false),
				},
				GuestAllowed: true,
				UserIndependant: true,
			},
			"POST": {
				Description: "Create a new entry",
				Returns: "InfoEntry",
				Params: QueryParams{
					"title":             MkQueryInfo(P_NotEmpty, true),
					"type":              MkQueryInfo(P_EntryType, true),
					"format":            MkQueryInfo(P_EntryFormat, true),
					"timezone":          MkQueryInfo(P_NotEmpty, false),
					"price":             MkQueryInfo(P_Float64, false),
					"currency":          MkQueryInfo(P_NotEmpty, false),
					"is-digital":        MkQueryInfo(P_Bool, false),
					"is-anime":          MkQueryInfo(P_Bool, false),
					"format-modifiers":  MkQueryInfo(P_Int64, false),
					"art-style":         MkQueryInfo(P_ArtStyle, false),
					"libraryId":         MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"parentId":          MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"copyOf":            MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"native-title":      MkQueryInfo(P_True, false),
					"tags":              MkQueryInfo(P_True, false),
					"location":          MkQueryInfo(P_True, false),
					"metadata":          MkQueryInfo(P_NotEmpty, false), // user metadata
					"get-metadata":      MkQueryInfo(P_Bool, false), // use heuristics (unless metadata-provider is given) to get metadata
					"metadata-provider": MkQueryInfo(P_MetaProvider, false), // use this provider when using get-metadata
					"user-rating":       MkQueryInfo(P_Float64, false),
					"user-status":       MkQueryInfo(P_UserStatus, false),
					"user-view-count":   MkQueryInfo(P_Int64, false),
					"user-notes":        MkQueryInfo(P_True, false),
					"requires":          MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"recommended-by":    MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
	},

	{
		EndPoint: "entry/{id}",
		Handler: SpecificEntry,
		PathParams: QueryParams {
			"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
		},
		Description: "Various methods for a specific entry",
		Methods: map[string]MethodSpec {
			"GET": {
				Description: `Get various information about an entry
				Values for ?kind
				<dl>
					<dt> relations
					<dd> lists relations of the entry
					<dt> events
					<dd> lists all events
					<dt> transactions
					<dd> lists all transactions
					<dt> info
					<dd> get info
					<dt> meta
					<dd> get metadata
					<dt> user
					<dd> get user viewing info
				</dl>
				`,
				Params: QueryParams {
					"kind": MkQueryInfo(P_NotEmpty, false),
				},
			},
			"TRANSACT": {
				Description: "registers a transaction",
				Params: QueryParams{
					"price": MkQueryInfo(P_Float64, true),
					"currency": MkQueryInfo(P_NotEmpty, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
					"eventId": MkQueryInfo(P_NotEmpty, false),
				},
			},
			"PATCH": {
				Description: "Update the info entry with new information in the form of Partial<InfoEntry> from the request body",
			},
			"DELETE": {
				Description: "Delete the entry {id}",
			},
			"DROP": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
				Description: "Drops an entry",
			},
			"RESUME": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
				Description: "Resumes an entry",
			},
			"PAUSE": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
				Description: "Pauses an entry",
			},
			"PLAN": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
				Description: "Plans an entry",
			},
			"BEGIN": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
				Description: "Begins an entry",
			},
			"WAIT": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
				Description: "Waits an entry",
			},
			"FINISH": {
				Params: QueryParams{
					"timezone": MkQueryInfo(P_NotEmpty, false),
					"rating": MkQueryInfo(P_Float64, true),
				},
				Description: "Finishes an entry",
			},
		},
	},

	{
		EndPoint: "entry/allfor",
		Handler: GetAllForEntry2,
		Methods: map[string]MethodSpec{
			"GET": {
				Params: QueryParams {
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
				GuestAllowed: true,
				UserIndependant: true,
			},
		},
		Description: "Gets all related information for an entry",
	},

	{
		Handler:     AddTags,
		Description: "Adds tag(s) to an entry, tags must be an \\x1F (ascii unit separator) separated list",
		EndPoint:    "add-tags",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"tags": MkQueryInfo(P_TList("\x1F", func(in string) string {
						return in
					}), true),
				},
			},
		},
	},

	{
		Handler:     DeleteTags,
		Description: "Delets tag(s) from an entry, tags must be an \\x1F (ascii unit separator) separated list",
		Aliases:    []string{"delete-tags"},
		EndPoint: "entry/tag/delete",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"tags": MkQueryInfo(P_TList("\x1F", func(in string) string {
						return in
					}), true),
				},
			},
		},
	},

	{
		Handler: AddEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"title":             MkQueryInfo(P_NotEmpty, true),
					"type":              MkQueryInfo(P_EntryType, true),
					"format":            MkQueryInfo(P_EntryFormat, true),
					"timezone":          MkQueryInfo(P_NotEmpty, false),
					"price":             MkQueryInfo(P_Float64, false),
					"currency":          MkQueryInfo(P_NotEmpty, false),
					"is-digital":        MkQueryInfo(P_Bool, false),
					"is-anime":          MkQueryInfo(P_Bool, false),
					"format-modifiers":  MkQueryInfo(P_Int64, false),
					"art-style":         MkQueryInfo(P_ArtStyle, false),
					"libraryId":         MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"parentId":          MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"copyOf":            MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"native-title":      MkQueryInfo(P_True, false),
					"tags":              MkQueryInfo(P_True, false),
					"location":          MkQueryInfo(P_True, false),
					"metadata":          MkQueryInfo(P_NotEmpty, false), // user metadata
					"get-metadata":      MkQueryInfo(P_Bool, false), // use heuristics (unless metadata-provider is given) to get metadata
					"metadata-provider": MkQueryInfo(P_MetaProvider, false), // use this provider when using get-metadata
					"user-rating":       MkQueryInfo(P_Float64, false),
					"user-status":       MkQueryInfo(P_UserStatus, false),
					"user-view-count":   MkQueryInfo(P_Int64, false),
					"user-notes":        MkQueryInfo(P_True, false),
					"requires":          MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					"recommended-by":    MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Deprecated: "Use POST /entry/ instead",
		Description: "Adds a new entry, and registers an Add event",
		Returns:     "InfoEntry",
		Aliases:    []string{"add-entry"},
		EndPoint: "entry/add",
	},

	{
		Aliases: []string{"delete-entry"},
		EndPoint: "entry/delete",
		Handler:  DeleteEntry,
		Deprecated: "use DELETE /entry/{id}",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Deletes an entry",
	},

	{
		Aliases:     []string{"list-tree"},
		EndPoint: "entry/tree",
		Deprecated: "use TREE /entry",
		Handler:      GetTree,
		Methods: map[string]MethodSpec {
			"GET": {
				Params:  QueryParams{},
				GuestAllowed: true,
			},
		},
		Description:  "Gets a tree-like json structure of all entries",
		Returns:      "InfoEntry",
	},

	{
		EndPoint: "entry/mod",
		Aliases: []string{"mod-entry"},
		Handler:  ModEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
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
			},
		},
		Description: "Modifies an individual entry datapoint",
	},

	{
		EndPoint: "entry/set",
		Aliases:    []string{"set-entry"},
		Handler:     SetEntry,
		Deprecated: "use PATCH /entry/{id}",
		Methods: map[string]MethodSpec {
			"POST": {
				Params: QueryParams {
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Set an entry to the json of an entry<br>Post body must be updated entry",
	},

	{
		EndPoint: "entry/list",
		Deprecated: "use GET /entry?kind=info",
		Aliases: []string{"list-entries"},
		Handler:  ListEntries,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"sort-by": MkQueryInfo(P_SqlSafe, false),
				},
				GuestAllowed: true,
			},
		},
		Description:  "List info entries",
		Returns:      "JSONL<InfoEntry>",
	},


	{
		Aliases: []string{"add-child"},
		EndPoint: "entry/child/add",
		Handler: AddChild,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"child": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"parent": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Adds child as a child of parent",
	},

	{
		Aliases: []string{"del-child"},
		EndPoint: "entry/child/delete",
		Handler: DelChild,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"child": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"parent": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Removes child as a child of parent",
	},

	{
		Aliases: []string{"list-descendants"},
		EndPoint: "entry/child/list",
		Handler:  GetDescendants,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
				GuestAllowed: true,
			},
		},
		Description:  "Lists children of an entry",
		Returns:      "JSONL<InfoEntry>",
	},

	{
		Aliases: []string{"add-copy"},
		EndPoint: "entry/copy/add",
		Handler: AddCopy,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"copy": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"copyof": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Makes 2 items copies of each other",
	},

	{
		Aliases: []string{"del-copy"},
		EndPoint: "entry/copy/delete",
		Handler: DelCopy,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"copy": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"copyof": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Makes copy not a copy of copyof anymore, goes both directions",
	},

	{
		Aliases: []string{"list-copies"},
		EndPoint: "entry/copy/list",
		Handler:  GetCopies,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
				GuestAllowed: true,
			},
		},
		Description:  "Lists copies of an entry",
		Returns:      "JSONL<InfoEntry>",
	},

	{
		Aliases: []string{"del-requires"},
		EndPoint: "entry/requirement/delete",
		Handler: DelRequires,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"itemid": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"requires": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Removes requires as a requirement of itemid",
	},


	{
		Aliases: []string{"add-requires"},
		EndPoint: "entry/requirement/add",
		Handler: AddRequires,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"itemid": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"requires": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Makes itemid require requires as a requirement",
	},

	{
		Aliases: []string{"list-relations"},
		EndPoint: "entry/relation/list",
		Deprecated: "use GET /entry?kind=relations",
		Handler: ListRelations,
		Description: "Lists relations of all entries",
		Methods: map[string]MethodSpec{
			"GET": {
				GuestAllowed: true,
			},
		},
	},

	{
		Aliases: []string{"stream-entry"},
		EndPoint: "entry/stream",
		Handler:  Stream,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":      MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"subfile": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Download the file located by the {id}'s location",
		Returns:     "any",
	},

	{
		Aliases:     []string{"list-collections"},
		EndPoint:     "entry/collection/list",
		Handler:      ListCollections,
		Methods: map[string]MethodSpec {
			"GET": {
				Params:  QueryParams{},
				GuestAllowed: true,
			},
		},
		Description:  "Lists en_title of all entries who's type is Collection",
		Returns:      "Sep<string, '\\n'>",
	},

	{
		Aliases:      []string{"list-libraries"},
		EndPoint:     "entry/library/list",
		Handler:      ListLibraries,
		Methods: map[string]MethodSpec {
			"GET": {
				Params:  QueryParams{},
				GuestAllowed: true,
			},
		},
		Description:  "Lists ids of all entries who's type is Library",
		Returns:      "Sep<string, '\\n'>",
	},

	{
		Aliases: []string{"recommenders"},
		EndPoint: "entry/recommender/list",
		Handler: GetRecommenders,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams {},
				GuestAllowed: true,
				UserIndependant: false,
			},
		},
		Description: "Gets a list of all recommenders",
		Returns: "string \\x1F (unit separator) separated",
	},

	{
		EndPoint: "entry/settings",
		Handler: EntrySettings,
		Methods: map[string]MethodSpec {
			"GET": {
				Description: "Gets the settings for an entry",
				Params: QueryParams {
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
			"POST": {
				Description: "Sets the settings for an entry. Body data must be Partial<EntrySettings>, any fields not specified will remain unchanged",
				Params: QueryParams {
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
	},
	// }}}
} // }}}

// `/metadata` endpoints {{{
var metadataEndpointList = []ApiEndPoint{
	{
		EndPoint: "fetch-location",
		Handler:  FetchLocation,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":          MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
					"provider":    MkQueryInfo(P_LocationProvider, false),
					"provider-id": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Fetch the location of an entry based on the metadata and other info",
	},
	{
		Aliases: []string{"identify"},
		EndPoint: "search",
		Handler:  IdentifyWithSearch,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"title":    MkQueryInfo(P_NotEmpty, true),
					"provider": MkQueryInfo(P_Identifier, true),
				},
			},
		},
		Description: `List metadata results based on a search query + provider<br>
The id of the metadata entry will be the id that's supposed to be given to <code>identified-id</code><br>
when using finalize-identify`,
		Returns: "JSONL<MetadataEntry>",
	},

	{
		Aliases: []string{"finalize-identify"},
		EndPoint: "apply",
		Handler:  FinalizeIdentification,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"identified-id": MkQueryInfo(P_NotEmpty, true),
					"provider":      MkQueryInfo(P_IdIdentifier, true),
					"apply-to":      MkQueryInfo(P_VerifyIdAndGetMetaEntry, false),
				},
			},
		},
		Description: "Apply an identified id from /identify, to an entry using a provider",
		Returns:     "none",
	},

	{
		Aliases: []string{"set-entry"},
		EndPoint:    "set",
		Handler:     SetMetadataEntry,
		Methods: map[string]MethodSpec {
			"POST": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
				},
			},
		},
		Description: "Set a metadata entry to the json of an entry<br>post body must be updated metadata entry",
		Returns:     "UserEntry",
	},

	{
		EndPoint: "fetch",
		Handler:  FetchMetadataForEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"provider": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Returns: "MetadataEntry",
		Description: `Fetch the metadata for an entry based on the type<br>
	and using EntryInfo.En_Title as the title search<br>
	if provider is not given, it is automatically chosen based on type`,
	},

	{
		Aliases: []string{"retrieve"},
		EndPoint: "get",
		Deprecated: "use GET entry/{id}?kind=meta",
		Handler:  RetrieveMetadataForEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
				},
				GuestAllowed: true,
			},
		},
		Description:  "Gets the metadata for an entry",
		Returns:      "MetadataEntry",
	},

	{
		Aliases: []string{"mod-entry"},
		EndPoint: "mod",
		Handler:  ModMetadataEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":              MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
					"rating":          MkQueryInfo(P_Float64, false),
					"description":     MkQueryInfo(P_NotEmpty, false),
					"release-year":    MkQueryInfo(P_Int64, false),
					"thumbnail":       MkQueryInfo(P_NotEmpty, false),
					"media-dependant": MkQueryInfo(P_NotEmpty, false),
					"datapoints":      MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Modify metadata by datapoint",
	},

	{
		EndPoint: "set-thumbnail",
		Handler:  SetThumbnail,
		Methods: map[string]MethodSpec {
			"POST": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
				},
			},
		},
		Description: "Set the thumbnail for a metadata entry",
	},

	{
		Aliases:     []string{"list-entries"},
		EndPoint:     "list",
		Deprecated: "use GET /entry?kind=meta",
		Handler:      ListMetadata,
		Methods: map[string]MethodSpec {
			"GET": {
				Params:  QueryParams{},
				GuestAllowed: true,
			},
		},
		Description:  "Lists all metadata entries",
		Returns:      "JSONL<MetadataEntry>",
	},
} // }}}

// `/engagement` endpoints {{{
var engagementEndpointList = []ApiEndPoint{
	{
		EndPoint: "copy",
		Handler:  CopyUserViewingEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"src-id":  MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"dest-id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
			},
		},
		Description: "Moves all user entry data, and events from one entry entry to another",
	},

	{
		Aliases: []string{"get-events"},
		EndPoint: "event/listfor",
		Handler:  GetEventsOf,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
				},
				GuestAllowed: true,
			},
		},
		Description:  "Lists the events of an entry",
		Returns:      "JSONL<EventEntry>",
	},
	{
		EndPoint: "delete-event",
		Handler:  DeleteEvent,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":        MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"timestamp": MkQueryInfo(P_Int64, true),
					"after":     MkQueryInfo(P_Int64, true),
					"before":    MkQueryInfo(P_Int64, true),
				},
			},
		},
		Description: "<b>DEPRECATED, use event/delete</b>",
	},

	{
		Aliases: []string{"delete-event-v2"},
		EndPoint: "event/delete",
		Handler:  DeletEventV2,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_Int64, true),
				},
			},
		},
		Description: "Deletes an event by event id",
	},

	{
		Aliases: []string{"register-event"},
		EndPoint: "event/register",
		Handler:  RegisterEvent,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":        MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
					"name":      MkQueryInfo(P_NotEmpty, true),
					"timestamp": MkQueryInfo(P_Int64, false),
					"after":     MkQueryInfo(P_Int64, false),
					"timezone":  MkQueryInfo(P_NotEmpty, false),
					"before":    MkQueryInfo(P_Int64, false),
				},
			},
		},
		Description: "Registers an event for an entry",
	},

	{
		Aliases: []string{"edit-event"},
		EndPoint: "event/edit",
		Handler: EditEvent,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"eventId":   MkQueryInfo(P_Int64, true),
					"name":      MkQueryInfo(P_NotEmpty, true),
					"timestamp": MkQueryInfo(P_Int64, false),
					"after":     MkQueryInfo(P_Int64, false),
					"timezone":  MkQueryInfo(P_NotEmpty, false),
					"before":    MkQueryInfo(P_Int64, false),
				},
			},
		},
	},

	{
		Aliases:     []string{"list-events"},
		EndPoint: "event/list",
		Handler:      ListEvents,
		Methods: map[string]MethodSpec {
			"GET": {
				Params:  QueryParams{},
				GuestAllowed: true,
			},
		},
		Description:  "Lists all events",
		Returns:      "JSONL<EventEntry>",
	},

	{
		Aliases: []string{"mod-entry"},
		EndPoint: "mod",
		Handler:  ModUserEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":               MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"notes":            MkQueryInfo(P_True, false),
					"rating":           MkQueryInfo(P_Float64, false),
					"view-count":       MkQueryInfo(P_Int64, false),
					"current-position": MkQueryInfo(P_True, false),
					"status":           MkQueryInfo(P_UserStatus, false),
					"minutes":          MkQueryInfo(P_Int64, false),
				},
			},
		},
		Description: "Modifies datapoints of a user entry",
	},

	{
		Aliases:    []string{"set-entry"},
		EndPoint: "set",
		Handler:     SetUserEntry,
		Methods: map[string]MethodSpec {
			"POST": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
				},
			},
		},
		Description: "Updates the user entry with the post body<br>Post body must be updated user entry",
	},

	{
		Aliases:     []string{"list-entries"},
		EndPoint: "list",
		Handler:      UserEntries,
		Deprecated: "use GET /entry?kind=user",
		Description:  "Lists all user entries",
		Returns:      "JSONL<UserEntry>",
		Methods: map[string]MethodSpec {
			"GET": {
				GuestAllowed: true,
			},
		},
	},

	{
		Aliases: []string{"get-entry"},
		EndPoint: "get",
		Deprecated: "use GET /entry/{id}?kind=user",
		Handler:  GetUserEntry,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
				},
				GuestAllowed: true,
			},
		},
		Description:  "Gets a user entry by id",
		Returns:      "UserEntry",
	},

	{
		Aliases: []string{"drop-media"},
		EndPoint: "drop",
		Handler:  DropMedia,
		Deprecated: "use DROP /entry/{id}",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Drops a media, and registers a Drop event",
	},

	{
		Aliases: []string{"resume-media"},
		EndPoint: "resume",
		Deprecated: "use RESUME /entry/{id}",
		Handler:  ResumeMedia,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Resumes a media and registers a ReViewing event",
	},

	{
		Aliases: []string{"pause-media"},
		EndPoint: "pause",
		Deprecated: "use PAUSE /entry/{id}",
		Handler:  PauseMedia,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Pauses a media and registers a Pause event",
	},

	{
		Aliases: []string{"plan-media"},
		EndPoint: "plan",
		Deprecated: "use PLAN /entry/{id}",
		Handler:  PlanMedia,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Plans a media and registers a Plan event",
	},

	{
		Aliases: []string{"begin-media"},
		EndPoint: "begin",
		Deprecated: "use BEGIN /entry/{id}",
		Handler:  BeginMedia,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Begins a media and registers a Viewing event",
	},
	{
		Aliases: []string{"wait-media"},
		EndPoint: "wait",
		Deprecated: "use WAIT /entry/{id}",
		Handler:  WaitMedia,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Sets the status to waiting, and registers a Waiting event",
	},

	{
		Aliases: []string{"finish-media"},
		EndPoint: "finish",
		Deprecated: "use FINISH /entry/{id}",
		Handler:  FinishMedia,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id":       MkQueryInfo(P_VerifyIdAndGetUserEntry, true),
					"rating":   MkQueryInfo(P_Float64, true),
					"timezone": MkQueryInfo(P_NotEmpty, false),
				},
			},
		},
		Description: "Finishes a media, and registers a Finish event",
	},

} //}}}

// `/account` endpoints {{{
var AccountEndPoints = []ApiEndPoint{
	{
		EndPoint:        "create",
		Handler:         CreateAccount,
		Methods:          map[string]MethodSpec{
			"POST": {
				UserIndependant: true,
				GuestAllowed:    true,
			},
		},
		Description:     "Creates an account",
	},

	{
		EndPoint: "access/gen-code",
		Handler: GenSyncCode,
		Description: "Generates a code to let the user give a client access to their account",
	},

	{
		EndPoint: "access/verify-code",
		Handler: VerifySyncCode,
		Description: "Verify a code given from /account/gen-code, returns a random hashstring that can then be used to authenticate as the user<br>Optionally, have a label to describe what the newly generated hashstring is for",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams {
					"code": MkQueryInfo(P_NotEmpty, true),
					"label": MkQueryInfo(P_NotEmpty, true),
				},
				GuestAllowed: true,
				UserIndependant: true,
			},
		},
	},

	{
		EndPoint: "access/list",
		Handler: ListAccesses,
		Description: "List all hashes that are being used as authentication strings",
	},

	{
		EndPoint: "access/delete",
		Handler: DeleteAccessCode,
		Description: "Deletes an access code",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams {
					"label": MkQueryInfo(P_NotEmpty, true),
				},
			},
		},
	},

	{
		EndPoint:    "username2id",
		Handler:     Username2Id,
		Description: "get a user's id from username",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"username": MkQueryInfo(P_NotEmpty, true),
				},
				UserIndependant: true,
				GuestAllowed:    true,
			},
		},
	},

	{
		EndPoint: "login",
		Handler:  Login,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"username": MkQueryInfo(P_NotEmpty, false),
					"password": MkQueryInfo(P_NotEmpty, false),
				},
				UserIndependant: true,
				GuestAllowed:    true,
			},
		},
		Description:     "Login",
	},

	{
		EndPoint:        "authorized",
		Handler:         AuthCk,
		Description:     "Checks if the Authorization header is valid",
		Methods: map[string]MethodSpec {
			"GET": {
				UserIndependant: true,
			},
		},
	},

	{
		EndPoint:        "list",
		Handler:         ListUsers,
		Description:     "List all users",
		Methods: map[string]MethodSpec {
			"GET": {
				UserIndependant: true,
				GuestAllowed:    true,
			},
		},
	},

	{
		EndPoint:    "delete",
		Methods:      map[string]MethodSpec{
			"DELETE": {},
		},
		Description: "Delete an account",
		Handler:     DeleteAccount,
	},

	{
		EndPoint: "rename",
		Description: "change your username",
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"new-username": MkQueryInfo(P_NotEmpty, true),
				},
				UserIndependant: true,
			},
		},
		Handler: RenameAccount,
	},
} // }}}

// `/resource` endpoints {{{
var resourceEndpointList = []ApiEndPoint{
	{
		EndPoint: "get-thumbnail",
		Handler:  ThumbnailResource,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"hash": MkQueryInfo(P_NotEmpty, true),
				},
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
		Description:     "Gets the thumbnail for an id (if it can find the thumbnail in the thumbnails dir)",
	},
	{
		EndPoint: "get-thumbnail-by-id",
		Handler: ThumbnailResourceById,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams {
					"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
				},
				GuestAllowed: true,
				UserIndependant: true,
			},
		},
		Description: "Returns a 303 pointing to the location of where to find the thumbnail for an item id",
	},

	// this is the legacy one, since the url is hardcoded I can't really change it.
	{
		EndPoint: "thumbnail",
		Handler:  ThumbnailResourceLegacy,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_NotEmpty, true),
				},
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
		Description:     "LEGACY, Gets the thumbnail for an id (if it can find the thumbnail in the thumbnails dir)",
	},

	{
		EndPoint: "download-thumbnail",
		Handler:  DownloadThumbnail,
		Methods: map[string]MethodSpec {
			"GET": {
				Params: QueryParams{
					"id": MkQueryInfo(P_VerifyIdAndGetMetaEntry, true),
				},
			},
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
		Methods: map[string]MethodSpec {
			"GET": {
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
	},

	{
		EndPoint:        "type",
		Handler:         ListTypes,
		Description:     "Lists the types for a Type",
		Methods: map[string]MethodSpec {
			"GET": {
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
	},

	{
		EndPoint:        "artstyle",
		Handler:         ListArtStyles,
		Description:     "Lists the types art styles",
		Methods: map[string]MethodSpec {
			"GET": {
				GuestAllowed:    true,
				UserIndependant: true,
			},
		},
	},
} // }}}

// `/docs` endpoints {{{
var MainDocs = ApiEndPoint{
	EndPoint:        "",
	Handler:         DocHTML,
	Description:     "The documentation",
	Methods: map[string]MethodSpec {
		"GET": {
			GuestAllowed:    true,
			UserIndependant: true,
		},
	},
} // }}}

var Endpoints = map[string][]ApiEndPoint{
	"":            mainEndpointList,
	"/metadata":   metadataEndpointList,
	"/engagement": engagementEndpointList,
	"/type":       typeEndpoints,
	"/resource":   resourceEndpointList,
	"/transact": {
		{
			EndPoint: "{id}",
			Handler: TransactResource,
			PathParams: QueryParams {
				"id": MkQueryInfo(P_Int64, true),
			},
			Description: "perform operations on a transaction",
			Methods: map[string]MethodSpec {
				"PATCH": {
					Description: "Modify a transaction",
					Params: QueryParams {
						"price": MkQueryInfo(P_Float64, false),
						"currency": MkQueryInfo(P_NotEmpty, false),
						"eventId": MkQueryInfo(P_Int64, false),
						"itemId": MkQueryInfo(P_Int64, false),
					},
				},
				"DELETE": {
					Description: "Delete a transaction",
				},
			},
		},
		{
			EndPoint: "list",
			Handler: ListTransactions,
			Deprecated: "use GET /entry?kind=transactions",
			Methods: map[string]MethodSpec {
				"GET": {
					Params: QueryParams {
						"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, false),
					},
					GuestAllowed: true,
					UserIndependant: true,
				},
			},
		},
		{
			EndPoint: "do",
			Handler: Transact,
			Deprecated: "use TRANSACT /entry/{id}",
			Methods: map[string]MethodSpec {
				"GET": {
					Params: QueryParams {
						"id": MkQueryInfo(P_VerifyIdAndGetInfoEntry, true),
						"price": MkQueryInfo(P_Float64, true),
						"currency": MkQueryInfo(P_NotEmpty, true),
						"timezone": MkQueryInfo(P_NotEmpty, false),
						"eventId": MkQueryInfo(P_NotEmpty, false),
					},
				},
			},
		},
		{
			EndPoint: "edit",
			Handler: EditTransaction,
			Deprecated: "use PATCH /transact/{id}",
			Methods: map[string]MethodSpec {
				"GET": {
					Params: QueryParams {
						"id": MkQueryInfo(P_Int64, true),
						"price": MkQueryInfo(P_Float64, false),
						"currency": MkQueryInfo(P_NotEmpty, false),
						"eventId": MkQueryInfo(P_Int64, false),
						"itemId": MkQueryInfo(P_Int64, false),
					},
				},
			},
		},
		{
			EndPoint: "delete",
			Handler: DeleteTransaction,
			Deprecated: "use DELETE /transact/{id}",
			Methods: map[string]MethodSpec {
				"GET": {
					Params: QueryParams {
						"id": MkQueryInfo(P_Int64, true),
					},
				},
			},
		},
	},
}

// this way the html at least wont change until a server restart
var htmlCache []byte

func DocHTML(ctx RequestContext) {
	w := ctx.W

	if len(htmlCache) == 0 {
		html := "<style>.required::after { content: \"(required) \"; font-weight: bold; }</style>"
		tableOfContents := "<p>Table of contents</p><ul>"
		docsHTML := ""
		for _, root := range []string {
			"", "/engagement", "/metadata", "/transact", "/resource", "/account", "/type",
		} {
			if root != "" {
				tableOfContents += fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", root, root)
				docsHTML += fmt.Sprintf("<HR><h3 id=\"%s\">%s</h3>", root, root)
			} else {
				tableOfContents += fmt.Sprintf("<li><a href=\"#%s\">%s</a></li>", "/", "/")
				docsHTML += fmt.Sprintf("<HR><h3 id=\"%s\">%s</h3>", "/", "/")
			}
			for _, endP := range Endpoints[root] {
				html := endP.GenerateDocHTML(root)
				if endP.Deprecated != "" {
					docsHTML += fmt.Sprintf(
						"<details><summary><h3 style='display: inline'>deprecated %s/%s</h3></summary>%s</details>",
						root, endP.EndPoint, html,
					)
				} else {
					docsHTML += html
				}
			}
		}
		html += tableOfContents + "</ul>" + docsHTML
		htmlCache = []byte(html)
	}

	// use text template in order to have html not be escaped
	tmpl, err := template.ParseFiles("./docs/docs.html")
	if err != nil {
		util.WError(ctx.W, 500, "Could not render docs %s", err.Error())
		return
	}

	w.WriteHeader(200)

	tmpl.Execute(ctx.W, struct {
		Endpoints string
	}{Endpoints: string(htmlCache)})
}
