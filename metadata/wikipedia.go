package metadata

import (
	"aiolimas/settings"
	db_types "aiolimas/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
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

	req, err = http.NewRequest("GET", "https://en.wikipedia.org/w/api.php", nil)
	if err != nil {
		return out, err
	}
	req.URL.RawQuery = url.Values {
		"action": {"parse"},
		"page": {id},
		"format": {"json"},
		"section": {"1"}, // This is probably the section the user wants.
	}.Encode()


	req.Header.Set("User-Agent", "aio-limas/1.0 (https://github.com/Euro20179/AIO-LIMAS; anon5555@duck.com)")

	res, err = client.Do(req)
	if err != nil {
		return out, err
	}

	jsonText, err = io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	parsedResp := struct {
		Parse struct {
			Title string
			Pageid float64
			Revid float64
			Text map[string] string
		}
	}{}

	err = json.Unmarshal(jsonText, &parsedResp)
	if err != nil {
		return out, err
	}

	desc, has := parsedResp.Parse.Text["*"]

	if !has {
		return out, nil
	}

	tree, err := html.Parse(strings.NewReader(desc))
	if err != nil {
		return out, err
	}

	getattr := func(n *html.Node, attr string) string {
		for _, a := range n.Attr{
			if a.Key == attr {
				return a.Val
			}
		}
		return ""
	}

	var trimHTML func(n *html.Node)
	trimHTML = func(n *html.Node) {
		for c := n.FirstChild; c != nil; {
			next := c.NextSibling

			if c.Type == html.ElementNode{
				cls := getattr(c, "class")
				if strings.Contains(cls, "mw-editsection") || strings.Contains(cls, "reference") {
					n.RemoveChild(c)
					goto next
				}

				if c.Data == "a"  {
					// Insert each child of <a> before the <a>, keeping order.
					// Take children out of <a> as we insert them.
					insertBefore := next
					for ch := c.FirstChild; ch != nil; {
						chNext := ch.NextSibling

						c.RemoveChild(ch)

						n.InsertBefore(ch, insertBefore)

						ch = chNext
					}

					n.RemoveChild(c)

					goto next
				}
			}

			trimHTML(c)
next:
			c = next
		}
	}

	trimHTML(tree)

	builder := strings.Builder{}
	html.Render(&builder, tree)
	out.Description = builder.String()

	return out, nil
}
