package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"aiolimas/logging"
	"aiolimas/settings"
	db_types "aiolimas/types"
)

func googlebooksfillmeta(meta *db_types.MetadataEntry, item map[string]any) {
	mediaDependant := map[string]string{}

	volInfo := item["volumeInfo"].(map[string]any)
	industryIdents, ok := volInfo["industryIdentifiers"]
	if !ok {
		return
	}
	identifiers := industryIdents.([]any)
	ident := identifiers[0].(map[string]any)
	isbn := ident["identifier"].(string)
	isbnInt, err := strconv.ParseInt(isbn, 10, 64)
	if err != nil {
		return
	}
	meta.ItemId = isbnInt
	meta.ProviderID = isbn
	meta.Provider = "googlebooks"

	if title, ok := volInfo["title"]; ok {
		meta.Title = title.(string)
	}

	if images, ok := volInfo["imageLinks"]; ok {
		thumbs := images.(map[string]any)
		meta.Thumbnail = thumbs["thumbnail"].(string)
	}

	if desc, ok := volInfo["description"]; ok {
		meta.Description = desc.(string)
	}

	if categories, ok := volInfo["categories"]; ok {
		genresList := []string{}
		for _, catList := range categories.([]any) {
			cats := strings.Split(catList.(string), " ")
			for _, cat := range cats {
				genresList = append(genresList, cat)
			}
		}
		genres, err := json.Marshal(genresList)
		if err == nil {
			meta.Genres = string(genres)
		} else {
			logging.ELog(err)
			meta.Genres = ""
		}
	}

	pubDate := volInfo["publishedDate"].(string)
	yearSegmentEnd := strings.Index(pubDate, "-")
	if yearSegmentEnd != -1 {
		yearStr := pubDate[0:yearSegmentEnd]
		year, _ := strconv.ParseInt(yearStr, 10, 64)
		meta.ReleaseYear = year
	}

	if thumbs, ok := volInfo["imageLinks"]; ok {
		meta.Thumbnail = thumbs.(map[string] any)["thumbnail"].(string)
	}

	if pi, ok := volInfo["pageCount"]; ok {
		mediaDependant["Book-page-count"] = fmt.Sprintf("%.0f", pi.(float64))
	}
	d, _ := json.Marshal(mediaDependant)
	meta.MediaDependant = string(d)
}

func GoogleBooksIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	var out []db_types.MetadataEntry
	enc := url.PathEscape(info.Title)
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&langRestrict=en", enc)
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

	items, ok := jdata["items"]
	if !ok {
		return out, errors.New("no results")
	}

	hasErr, ok := jdata["error"]
	if ok {
		return out, errors.New(fmt.Sprintf("error from googlebooks: %s", hasErr.(map[string]any)["message"].(string)))
	}

	for _, i := range items.([]any) {
		var cur db_types.MetadataEntry

		googlebooksfillmeta(&cur, i.(map[string] any))

		out = append(out, cur)
	}
	return out, nil
}

func OpenLibraryIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry
	id = strings.ReplaceAll(id, "-", "")
	url := fmt.Sprintf("https://openlibrary.org/works/%s.json", url.QueryEscape(id))

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

	if err, ok := jdata["error"]; ok {
		return out, errors.New(err.(string))
	}

	out.ProviderID = id
	out.Provider = "openlibrary"

	descriptionβ, ok := jdata["description"]
	if ok {
		out.Description = descriptionβ.(map[string]any)["value"].(string)
	}

	titleβ, ok := jdata["title"]
	if ok {
		out.Title = titleβ.(string)
	}

	coversβ, ok := jdata["covers"]
	if ok {
		coverID := coversβ.([]any)[0]
		out.Thumbnail = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d.jpg", (int64)(coverID.(float64)))
	}

	subjectsβ, ok := jdata["subjects"]
	if ok {
		subjectsListβ := (subjectsβ.([]any))
		arr, err := json.Marshal(subjectsListβ)
		if err != nil {
			return out, err
		}

		out.Genres = string(arr)
	}

	return out, nil
}

func GoogleBooksProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry

	enc := url.PathEscape(info.Entry.En_Title)
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&langRestrict=en", enc)
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

	itemsCHK, ok := jdata["items"]
	if !ok {
		return out, errors.New("no results")
	}
	items := itemsCHK.([]any)

	item := items[0].(map[string]any)
	googlebooksfillmeta(&out, item)
	return out, nil
}

func GoogleBooksIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	id = strings.ReplaceAll(id, "-", "")
	q := fmt.Sprintf("isbn:%s", id)
	i := GetMetadataInfo{
		Entry: &db_types.InfoEntry{En_Title: q},
	}
	return GoogleBooksProvider(&i)
}
