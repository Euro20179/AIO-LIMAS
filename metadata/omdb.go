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

	"aiolimas/settings"
	"aiolimas/types"
	"aiolimas/logging"
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

func omdbResultToMetadata(result OMDBResponse) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}
	mediaDep := make(map[string]string)
	if result.Type == "series" {
		duration := strings.Split(result.Runtime, " ")[0]

		mediaDep["Show-episode-duration"] = duration
		mediaDep["Show-imdbid"] = result.ImdbID
	} else {
		length := strings.Split(result.Runtime, " ")[0]

		mediaDep[titleCase(result.Type) + "-director"] = result.Director
		mediaDep[titleCase(result.Type) + "-length"] = length
		mediaDep[titleCase(result.Type) + "-imdbid"] = result.ImdbID
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

	out.Provider = "omdb"
	out.ProviderID = result.ImdbID[2:]

	yearSep := "–"
	if strings.Contains(result.Year, yearSep) {
		result.Year = strings.Split(result.Year, yearSep)[0]
	}

	n, err := strconv.ParseInt(result.Year, 10, 64)
	if err == nil {
		out.ReleaseYear = n
	} else {
		logging.ELog(err)
	}

	return out, nil
}

func OMDBProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	entry := info.Entry
	var out db_types.MetadataEntry

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return out, errors.New("No api key")
	}

	search := entry.En_Title
	if search == "" {
		search = entry.Native_Title
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

func OmdbIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	outMeta := []db_types.MetadataEntry{}

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
		var cur db_types.MetadataEntry
		imdbId := entry.ImdbID[2:]
		imdbIdInt, err := strconv.ParseInt(imdbId, 10, 64)
		if err != nil {
			logging.ELog(err)
			continue
		}
		cur.ItemId = imdbIdInt
		cur.Title = entry.Title
		cur.Thumbnail = entry.Poster

		cur.Provider = "omdb"
		cur.ProviderID = imdbId

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func OmdbIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return out, errors.New("No api key")
	}

	for len(id) < 7 {
		id = "0" + id
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
