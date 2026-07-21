package metadata

import (
	"aiolimas/settings"
	db_types "aiolimas/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type WikiSearch struct {
	Batchcomplete string
	Continue any
	Query struct {
		Searchinfo struct {
			Totalhits float64
			Suggestion string
			Suggestionsnippet string
		}
		Search []struct {
			Ns float64
			Title string
			Pageid float64
			Size float64
			Wordcount float64
			Snippet string
			Timestamp string
		}
	}
}

type WikiSummary struct {
	Type string
	Title string
	Displaytitle string
	Namespace any
	Wikibase_item string
	Titles struct {
		Canonical string
		Normalized string
		Display string
	}
	Pageid float64
	Thumbnail struct {
		Source string
		Width float64
		Height float64
	}
	Originalimage struct {
		Source string
		Width float64
		Height float64
	}
	Lang string
	Dir string
	Revision string
	Tid string
	Timestamp string
	Description string
	Description_source string
	Content_urls any
	Extract string
	Extract_html string
}

func WikipediaIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	out := []db_types.MetadataEntry{}

	client := http.Client{}

	req, err := http.NewRequest("GET", "https://en.wikipedia.org/w/api.php", nil)
	if err != nil {
		return out, err
	}

	req.Header.Set("User-Agent", "aio-limas/1.0 (https://github.com/Euro20179/AIO-LIMAS; anon5555@duck.com)")

	req.URL.RawQuery = url.Values {
		"action": {"query"},
		"list": {"search"},
		"srsearch": {info.Title},
		"format": {"json"},
	}.Encode()

	res, err := client.Do(req)
	if err != nil {
		return out, err
	}

	jsonText, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	searchResult := WikiSearch{}
	err = json.Unmarshal(jsonText, &searchResult)
	if err != nil {
		return out, err
	}

	for _, result := range searchResult.Query.Search {
		cur := db_types.MetadataEntry{}
		cur.Description = result.Snippet
		cur.Title = result.Title
		cur.Provider = "wikipedia"
		cur.ItemId = int64(result.Pageid)
		cur.ProviderID = result.Title
		out = append(out, cur)
	}

	return out, nil
}

func WikipediaIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	client := http.Client{}


	eid := url.PathEscape(id)

	req, err := http.NewRequest("GET", fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/summary/%s", eid), nil)
	if err != nil {
		return out, err
	}

	req.Header.Set("User-Agent", "aio-limas/1.0 (https://github.com/Euro20179/AIO-LIMAS; anon5555@duck.com)")

	res, err := client.Do(req)
	if err != nil {
		return out, err
	}

	jsonText, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	searchResult := WikiSummary{}
	err = json.Unmarshal(jsonText, &searchResult)
	if err != nil {
		return out, err
	}

	out.ProviderID = id
	out.Provider = "wikipedia"
	out.Title = searchResult.Title
	out.Thumbnail = searchResult.Thumbnail.Source
	out.Description = searchResult.Extract_html

	return out, nil
}
