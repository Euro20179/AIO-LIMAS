package dynamic

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"aiolimas/accounts"
	"aiolimas/db"
	db_types "aiolimas/types"
	"aiolimas/util"
	"aiolimas/logging"
)

func handleSearchPath(w http.ResponseWriter, req *http.Request, uid int64) {
	query := req.URL.Query().Get("query")

	if query == "" {
		query = "#"
	}

	results, err := db.Search3(query)
	if err != nil {
		util.WError(w, 500, "Could not complete search: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)

	fnMap := template.FuncMap{
		"ArtStyle2Str": db_types.ArtStyle2Str,
		"Uid": func() int64 { return uid },
	}

	tmpl, err := template.New("base").Funcs(fnMap).ParseFiles(
		"./webservice/dynamic/templates/search-results.html",
		"./webservice/dynamic/templates/search-results-table-row.html",
	)
	if err != nil {
		logging.ELog(err)
	}

	err = tmpl.ExecuteTemplate(w, "search-results", results)
	if err != nil {
		logging.ELog(err)
	}
}

func handleById(w http.ResponseWriter, req *http.Request, id string, uid int64) {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		util.WError(w, 400, "Item id is not a valid id")
		return
	}

	info, err := db.GetInfoEntryById(uid, i)
	if err != nil {
		util.WError(w, 404, "Item not found")
		logging.ELog(err)
		return
	}

	meta, err := db.GetMetadataEntryById(uid, i)
	if err != nil {
		util.WError(w, 500, "Could not retrieve item metadata")
		logging.ELog(err)
		return
	}

	view, err := db.GetUserViewEntryById(uid, i)
	if err != nil {
		util.WError(w, 500, "Could not retrieve item viewing info")
		logging.ELog(err)
		return
	}

	events, err := db.GetEvents(uid, i)
	if err != nil {
		util.WError(w, 500, "Could not retrieve item events")
		logging.ELog(err)
		return
	}

	type AllInfo struct {
		Info   db_types.InfoEntry
		Meta   db_types.MetadataEntry
		View   db_types.UserViewingEntry
		Events []db_types.UserViewingEvent
	}

	allInfo := AllInfo{
		Info:   info,
		View:   view,
		Meta:   meta,
		Events: events,
	}

	if req.URL.Query().Has("fancy") {
		tmpl := template.Must(template.ParseFiles("./webservice/dynamic/templates/by-id.html"))
		tmpl.Execute(w, allInfo)
		return
	}

	text := "<head><link rel=\"stylesheet\" href=\"/css/general.css\"></head><body><dl>"

	t := reflect.TypeOf(info)
	v := reflect.ValueOf(info)
	fields := reflect.VisibleFields(t)

	for _, field := range fields {
		data := v.FieldByName(field.Name)
		text += fmt.Sprintf("<dt>%s</dt><dd name=\"%s\">%v</dd>", field.Name, field.Name, data.Interface())
	}
	text += "</dl>"

	t = reflect.TypeOf(view)
	v = reflect.ValueOf(view)
	fields = reflect.VisibleFields(t)

	for _, field := range fields {
		data := v.FieldByName(field.Name)
		text += fmt.Sprintf("<dt>%s</dt><dd name=\"%s\">%v</dd>", field.Name, field.Name, data.Interface())
	}
	text += "</dl>"

	t = reflect.TypeOf(meta)
	v = reflect.ValueOf(meta)
	fields = reflect.VisibleFields(t)

	for _, field := range fields {
		data := v.FieldByName(field.Name)
		text += fmt.Sprintf("<dt>%s</dt><dd name=\"%s\">%v</dd>", field.Name, field.Name, data.Interface())
	}
	text += "</dl>"

	w.Write([]byte(text))
}

func handleUsersPath(w http.ResponseWriter, req *http.Request) {
	aioPath := os.Getenv("AIO_DIR")
	users, err := accounts.ListUsers(aioPath)
	if err != nil {
		w.Write([]byte("An error occured while fetching users"))
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=\"UTF-8\"")
	w.Write([]byte("<link rel='stylesheet' href='/css/general.css'>"))
	w.Write([]byte("<style> div > * { border-bottom: 1px solid var(--secondary) }</style>"))
	w.Write([]byte("<div style='display: grid; grid-template-columns: 1fr 1fr; font-size: 1.2em; width: fit-content'><span style='padding: 0 4ch'>USERNAME</span> <span>ID</span>"))
	for _, user := range users {
		fmt.Fprintf(w, "<x-name style='padding: 0 4ch'>%s</x-name> <x-id>%d</x-id>\n", user.Username, user.Id)
	}
}

func HtmlEndpoint(w http.ResponseWriter, req *http.Request) {
	// the first item will be "", ignore it
	pathArguments := strings.Split(req.URL.Path, "/")[1:]

	pp := req.URL.Query()

	getuid := func() int64 {
		uid := pp.Get("uid")
		id, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			util.WError(w, 400, "Invalid user id")
			return 0
		}

		return id
	}

	switch pathArguments[1] {
	case "search":
		id := getuid()
		if id == 0 {
			return
		}
		handleSearchPath(w, req, id)
	case "users":
		handleUsersPath(w, req)
	case "by-id":
		uid := getuid()
		if uid == 0 {
			return
		}
		if len(pathArguments) < 3 || pathArguments[2] == "" {
			handleSearchPath(w, req, uid)
		} else {
			id := pathArguments[2]
			handleById(w, req, id, uid)
		}
	case "":
		http.ServeFile(w, req, "./webservice/dynamic/help.html")
	}
}
