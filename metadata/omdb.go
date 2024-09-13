package metadata

import (
	"aiolimas/db"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type OMDBResponse struct {
	Title string
	Year string
	Rated string
	Released string
	Runtime string
	Genre string
	Director string
	Writer string
	Actors string
	Plot string
	Language string
	Country string
	Awards string
	Poster string
	Ratings [] struct {
		Source string
		Value string
	}
	Metascore string
	ImdbRating string `json:"imdbRating"`
	ImdbVotes string `json:"imdbVotes"`
	ImdbID string `json:"imdbID"`
	Type string
	TotalSeasons string `json:"totalSeasons"`
	Response string
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
	if err != nil{
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil{
		return out, err
	}

	jData := new(OMDBResponse)
	err = json.Unmarshal(body, &jData)
	if err != nil{
		return out, err
	}

	mediaDep := make(map[string]string)
	if info.Type == "Show" {
		mediaDep["Show-episode-duration"] = strings.Split(jData.Runtime, " ")[0]
	} else {
		mediaDep[fmt.Sprintf("%s-length", info.Type)] = strings.Split(jData.Runtime, " ")[0]
	}
	mediaDep[fmt.Sprintf("%s-imdbid", info.Type)] = jData.ImdbID

	if jData.ImdbRating != "N/A" {
		res, err := strconv.ParseFloat(jData.ImdbRating, 64)
		if err == nil{
			out.Rating = res
		}
	}
	out.RatingMax = 10

	mdStr, err := json.Marshal(mediaDep)
	if err != nil{
		return out, err
	}

	out.MediaDependant = string(mdStr)

	out.Description = jData.Plot
	out.Thumbnail = jData.Poster
	if n, err := strconv.ParseInt(jData.Year, 10, 64); err == nil {
		out.ReleaseYear = n
	}

	return out, nil
}
