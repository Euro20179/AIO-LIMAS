package main

import (
	"flag"
	"net/http"

	globals "aiolimas/globals"
	api "aiolimas/api"
)

const (
	F_VHS       = iota // 0
	F_CD        = iota // 1
	F_DVD       = iota // 2
	F_BLURAY    = iota // 3
	F_4KBLURAY  = iota // 4
	F_MANGA     = iota // 5
	F_BOOK      = iota // 6
	F_DIGITAL   = iota // 7
	F_VIDEOGAME = iota // 8
	F_BOARDGAME = iota // 9
)


func main() {
	dbPathPtr := flag.String("db-path", "./all.db", "Path to the database file")
	flag.Parse()
	globals.InitDb(*dbPathPtr)

	v1 := "/api/v1"
	endPoints := map[string]func(http.ResponseWriter, *http.Request) {
		"add-entry": api.AddEntry,
		"query": api.QueryEntries,
	}
	for name, fn := range endPoints {
		http.HandleFunc(v1 + "/" + name, fn)
	}
	http.ListenAndServe(":8080", nil)
}
