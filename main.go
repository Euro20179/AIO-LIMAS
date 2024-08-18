package main

import (
	"flag"
	"net/http"

	api "aiolimas/api"
	globals "aiolimas/globals"
)

type Format int

const (
	F_VHS       Format = iota // 0
	F_CD        Format = iota // 1
	F_DVD       Format = iota // 2
	F_BLURAY    Format = iota // 3
	F_4KBLURAY  Format = iota // 4
	F_MANGA     Format = iota // 5
	F_BOOK      Format = iota // 6
	F_DIGITAL   Format = iota // 7
	F_VIDEOGAME Format = iota // 8
	F_BOARDGAME Format = iota // 9
)

func makeEndpoints(root string, endPoints map[string]func(http.ResponseWriter, *http.Request)) {
	for name, fn := range endPoints {
		http.HandleFunc(root+"/"+name, fn)
	}
}

func main() {
	dbPathPtr := flag.String("db-path", "./all.db", "Path to the database file")
	flag.Parse()
	globals.InitDb(*dbPathPtr)

	//paths
	//<root> general database stuff
	//

	type EndPointMap map[string]func(http.ResponseWriter, *http.Request)

	apiRoot := "/api/v1"
	makeEndpoints(apiRoot, EndPointMap{
		"add-entry": api.AddEntry,
		"query":     api.QueryEntries,
		"list-entries": api.ListEntries,
	})
	//for stuff relating to user viewing info
	//such as user rating, user beginning/ending a media, etc
	//stuff that would normally be managed by strack
	makeEndpoints(apiRoot + "/engagement", EndPointMap {
		"begin-media": api.BeginMedia,
	})

	http.ListenAndServe(":8080", nil)
}
