package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"aiolimas/settings"
	db_types "aiolimas/types"
)

func GoogleBooksIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	var out []db_types.MetadataEntry
	enc := url.QueryEscape(info.Title)
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s", enc)
	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jdata := map[string]any{}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		return out, err
	}

	items := jdata["items"].([]any)

	for _, i := range items {
		var cur db_types.MetadataEntry
		item := i.(map[string]any)
		volInfo := item["volumeInfo"].(map[string]any)
		identifiers := volInfo["industryIdentifiers"].([]any)
		ident := identifiers[0].(map[string]any)
		isbn := ident["identifier"].(string)
		isbnInt, err := strconv.ParseInt(isbn, 10, 64)
		if err != nil {
			continue
		}
		cur.ItemId = isbnInt
		cur.ProviderID = isbn
		cur.Provider = "googlebooks"
		title := volInfo["title"].(string)
		cur.Title = title
		thumbs := volInfo["imageLinks"].(map[string]any)
		cur.Thumbnail = thumbs["thumbnail"].(string)
		out = append(out, cur)
	}
	return out, nil
}

func OpenLibraryIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", id)

	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jdata := map[string]any{}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		return out, err
	}

	data := jdata[fmt.Sprintf("ISBN:%s", id)].(map[string]any)
	out.ProviderID = id
	out.Provider = "openlibrary"
	out.Title = data["title"].(string)
	md := map[string]string{}
	md["Book-page-count"] = fmt.Sprintf("%0f", data["number_of_pages"].(float64))
	y := data["publish_date"].(string)
	thumbs := data["cover"].(map[string]any)
	out.Thumbnail = thumbs["large"].(string)

	d, _ := json.Marshal(md)
	out.MediaDependant = string(d)

	year, err := strconv.ParseInt(y, 10, 64)
	if err != nil {
		return out, err
	}
	out.ReleaseYear = year
	return out, nil
}

func GoogleBooksProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry

	enc := url.QueryEscape(info.Entry.En_Title)
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s", enc)
	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jdata := map[string]any{}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		return out, err
	}

	items := jdata["items"].([]any)

	item := items[0].(map[string]any)
	volInfo := item["volumeInfo"].(map[string]any)
	identifiers := volInfo["industryIdentifiers"].([]any)
	ident := identifiers[0].(map[string]any)
	isbn := ident["identifier"].(string)
	out.ProviderID = isbn
	out.Provider = "googlebooks"
	out.Description = volInfo["description"].(string)

	pubDate := volInfo["publishedDate"].(string)
	yearSegmentEnd := strings.Index(pubDate, "-")
	yearStr := pubDate[0:yearSegmentEnd]
	year, _ := strconv.ParseInt(yearStr, 10, 64)
	out.ReleaseYear = year

	thumbs := volInfo["imageLinks"].(map[string]any)
	out.Thumbnail = thumbs["thumbnail"].(string)

	categories := volInfo["categories"].([]any)
	genreStr := []string{}
	for _, cat := range categories {
		genreStr = append(genreStr, cat.(string))
	}

	md := map[string]string{}
	md["Book-page-count"] = fmt.Sprintf("%.0f", volInfo["pageCount"].(float64))
	md["Book-genre"] = strings.Join(genreStr, ", ")
	d, _ := json.Marshal(md)
	out.MediaDependant = string(d)

	return out, nil
}
