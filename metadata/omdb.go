package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"aiolimas/db"
)

type OMDBResponse struct {
	Title    string
	Year     string
	Rated    string
	Released string
	Runtime  string
	Genre    string
	Director string
	Writer   string
	Actors   string
	Plot     string
	Language string
	Country  string
	Awards   string
	Poster   string
	Ratings  []struct {
		Source string
		Value  string
	}
	Metascore    string
	ImdbRating   string `json:"imdbRating"`
	ImdbVotes    string `json:"imdbVotes"`
	ImdbID       string `json:"imdbID"`
	Type         string
	TotalSeasons string `json:"totalSeasons"`
	Response     string
}

type OMDBSearchItem struct {
	Title  string
	Year   string
	ImdbID string `json:"imdbID"`
	Type   string
	Poster string
}

func titleCase(st string) string {
	return string(strings.ToTitle(st)[0]) + string(st[1:])
}

func omdbResultToMetadata(result OMDBResponse) (db.MetadataEntry, error) {
	out := db.MetadataEntry{}
	mediaDep := make(map[string]string)
	if result.Type == "series" {
		mediaDep["Show-episode-duration"] = strings.Split(result.Runtime, " ")[0]
		mediaDep["Show-imdbid"] = result.ImdbID
	} else {
		mediaDep[fmt.Sprintf("%s-length", titleCase(result.Type))] = strings.Split(result.Runtime, " ")[0]
		mediaDep[fmt.Sprintf("%s-imdbid", titleCase(result.Type))] = result.ImdbID
	}

	if result.ImdbRating != "N/A" {
		res, err := strconv.ParseFloat(result.ImdbRating, 64)
		if err == nil {
			out.Rating = res
		}
	}
	out.RatingMax = 10

	mdStr, err := json.Marshal(mediaDep)
	if err != nil {
		return out, err
	}

	out.MediaDependant = string(mdStr)

	out.Description = result.Plot
	out.Thumbnail = result.Poster

	if n, err := strconv.ParseInt(result.Year, 10, 64); err == nil {
		out.ReleaseYear = n
	}

	return out, nil
}

func OMDBProvider(info *db.InfoEntry, meta *db.MetadataEntry) (db.MetadataEntry, error) {
	var out db.MetadataEntry

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return out, errors.New("No api key")
	}

	search := info.En_Title
	if search == "" {
		search = info.Native_Title
	}
	if search == "" {
		return out, errors.New("No search possible")
	}

	url := fmt.Sprintf(
		"https://www.omdbapi.com/?apikey=%s&t=%s",
		key,
		url.QueryEscape(search),
	)

	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jData := new(OMDBResponse)
	err = json.Unmarshal(body, &jData)
	if err != nil {
		return out, err
	}

	return omdbResultToMetadata(*jData)
}

func OmdbIdentifier(info IdentifyMetadata) ([]db.MetadataEntry, error) {
	outMeta := []db.MetadataEntry{}

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return outMeta, errors.New("No api key")
	}

	searchTitle := info.Title
	url := fmt.Sprintf(
		"https://www.omdbapi.com/?apikey=%s&s=%s",
		key,
		url.QueryEscape(searchTitle),
	)

	res, err := http.Get(url)
	if err != nil {
		return outMeta, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return outMeta, err
	}

	jData := struct {
		Search []OMDBSearchItem
	}{}
	err = json.Unmarshal(body, &jData)
	if err != nil {
		return outMeta, err
	}

	for _, entry := range jData.Search {
		var cur db.MetadataEntry
		imdbId := entry.ImdbID[2:]
		imdbIdInt, err := strconv.ParseInt(imdbId, 10, 64)
		if err != nil {
			println(err.Error())
			continue
		}
		cur.ItemId = imdbIdInt
		cur.Title = entry.Title
		cur.Thumbnail = entry.Poster
		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func OmdbIdIdentifier(id string) (db.MetadataEntry, error) {
	out := db.MetadataEntry{}

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return out, errors.New("No api key")
	}

	url := fmt.Sprintf(
		"https://www.omdbapi.com/?apikey=%s&i=%s",
		key,
		url.QueryEscape("tt" + id),
	)

	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jData := new(OMDBResponse)
	err = json.Unmarshal(body, &jData)
	if err != nil {
		return out, err
	}

	return omdbResultToMetadata(*jData)
}
