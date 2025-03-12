package dynamic

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

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

func handleSearchPath(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query().Get("query")

	if query == "" {
		query = "#"
	}

	results, err := db.Search3(query)
	if err != nil {
		wError(w, 500, "Could not complete search: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)

	header := `
<tr>
	<th>Id</th>
	<th>User Title</th>
	<th>User Native Title</th>
	<th>Format</th>
	<th>Location</th>
	<th>Purchase Price</th>
	<th>Collection</th>
	<th>Parent Id</th>
	<th>Type</th>
	<th>Art Style</th>
	<th>Copy Of</th>
</tr>
`

	template := `
	<tr>
		<td name="ItemId"><a href="/html/by-id/%d" target="aio-item-output">%d</a></td>
		<td name="En_Title"><a href="/html/by-id/%d?fancy" target="aio-item-output">%s</a></td>
		<td name="Native_Title">%s</td>
		<td name="Format" data-format-raw="%d">%s</td>
		<td name="Location">%s</td>
		<td name="PurchasePrice">%.02f</td>
		<td name="Collection">%s</td>
		<td name="ParentId">%d</td>
		<td name="Type">%s</td>
		<td name="ArtStyle" data-art-raw="%d">%s</td>
		<td name="CopyOf">%d</td>
	</tr>
`

	text := "<head><link rel=\"stylesheet\" href=\"/css/general.css\"><link rel=\"stylesheet\" href=\"/lite/css/item-table.css\"></head><body><table id='info-table'>" + header
	for _, item := range results {
		artStr := db_types.ArtStyle2Str(item.ArtStyle)
		fmtStr := db_types.ListFormats()[item.Format]
		text += fmt.Sprintf(template, item.ItemId, item.ItemId, item.ItemId, item.En_Title, item.Native_Title, item.Format, fmtStr, item.Location, item.PurchasePrice, item.Collection, item.ParentId, item.Type, item.ArtStyle, artStr, item.CopyOf)
	}
	text += "</table></body>"
	fmt.Fprint(w, text)
}

func handleById(w http.ResponseWriter, req *http.Request, id string) {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		wError(w, 400, "Item id is not a valid id")
		return
	}

	info, err := db.GetInfoEntryById(i)
	if err != nil {
		wError(w, 404, "Item not found")
		println(err.Error())
		return
	}

	meta, err := db.GetMetadataEntryById(i)
	if err != nil {
		wError(w, 500, "Could not retrieve item metadata")
		println(err.Error())
		return
	}

	view, err := db.GetUserViewEntryById(i)
	if err != nil {
		wError(w, 500, "Could not retrieve item viewing info")
		println(err.Error())
		return
	}

	events, err := db.GetEvents(i)
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

func HtmlEndpoint(w http.ResponseWriter, req *http.Request) {
	// the first item will be "", ignore it
	pathArguments := strings.Split(req.URL.Path, "/")[1:]

	switch pathArguments[1] {
	case "search":
		handleSearchPath(w, req)
	case "by-id":
		if len(pathArguments) < 3 || pathArguments[2] == "" {
			handleSearchPath(w, req)
		} else {
			id := pathArguments[2]
			handleById(w, req, id)
		}
	case "":
		http.ServeFile(w, req, "./webservice/dynamic/help.html")
	}
}
