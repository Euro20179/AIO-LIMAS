package dynamic

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"aiolimas/api"
	"aiolimas/db"
	db_types "aiolimas/types"
)

// FIXME: make the codepath for this function and the IDENTICAL function in api/api.go  the same function
func wError(w http.ResponseWriter, status int, format string, args ...any) {
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)

	// also write to stderr
	fmt.Fprintf(os.Stderr, format, args...)
}

func handleSearchPath(w http.ResponseWriter, req *http.Request, uid int64) {
	query := req.URL.Query().Get("query")

	if query == "" {
		query = "#"
	}

	results, err := db.Search3(uid, query)
	if err != nil {
		wError(w, 500, "Could not complete search: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)

	fnMap := template.FuncMap{
		"ArtStyle2Str": db_types.ArtStyle2Str,
	}

	tmpl, err := template.New("base").Funcs(fnMap).ParseFiles(
		"./webservice/dynamic/templates/search-results.html",
		"./webservice/dynamic/templates/search-results-table-row.html",
	)
	if err != nil {
		println(err.Error())
	}

	err = tmpl.ExecuteTemplate(w, "search-results", results)
	if err != nil {
		println(err.Error())
	}
}

func handleById(w http.ResponseWriter, req *http.Request, id string, uid int64) {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		wError(w, 400, "Item id is not a valid id")
		return
	}

	info, err := db.GetInfoEntryById(uid, i)
	if err != nil {
		wError(w, 404, "Item not found")
		println(err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(uid, i)
	if err != nil {
		wError(w, 500, "Could not retrieve item metadata")
		println(err.Error())
		return
	}

	view, err := db.GetUserViewEntryById(uid, i)
	if err != nil {
		wError(w, 500, "Could not retrieve item viewing info")
		println(err.Error())
		return
	}

	events, err := db.GetEvents(uid, i)
	if err != nil {
		wError(w, 500, "Could not retrieve item events")
		println(err.Error())
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

func HtmlEndpoint(w http.ResponseWriter, req *http.Request, pp api.ParsedParams) {
	// the first item will be "", ignore it
	pathArguments := strings.Split(req.URL.Path, "/")[1:]

	switch pathArguments[1] {
	case "search":
		handleSearchPath(w, req, pp["uid"].(int64))
	case "by-id":
		if len(pathArguments) < 3 || pathArguments[2] == "" {
			handleSearchPath(w, req, pp["uid"].(int64))
		} else {
			id := pathArguments[2]
			handleById(w, req, id, pp["uid"].(int64))
		}
	case "":
		http.ServeFile(w, req, "./webservice/dynamic/help.html")
	}
}
