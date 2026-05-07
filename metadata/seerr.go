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
	db_types "aiolimas/types"
)

type SeerrResult struct {
	Id               int
	MediaType        string
	Adult            bool
	GenreIds         []int
	OriginalLanguage string
	OriginalTitle    string
	Overview         string
	Popularity       float64
	ReleaseDate      string
	Title            string
	Video            bool
	VoteAverage      float64
	VoteCount        int
	BackdropPath     string
	PosterPath       string
	MediaInfo        SeerrMediaInfo
}

type SeerrMediaInfo struct {
	DownloadStatus        []any
	DownloadStatus4k      []any
	Id                    int
	MediaType             string
	TmdbId                int
	TvdbId                int
	ImdbId                string
	Status                int
	Status4k              int
	CreatedAt             string
	UpdatedAt             string
	LastSeasonChange      string
	MediaAddedAt          string
	ServiceId             any
	ServiceId4k           any
	ExternalServiceId     any
	ExternalServiceId4k   any
	ExternalServiceSlug   any
	ExternalServiceSlug4k any
	RatingKey             any
	RatingKey4k           any
	JellyfinMediaId       string
	JellyfinMediaId4k     string
	WatchLists            []any
	MediaUrl              string
}

type SeerrResults struct {
	Page         int
	TotalPages   int
	TotalResults int
	Results      []SeerrResult
}

type SeerGenre struct {
	id int
	name string
}

var seer_genres []SeerGenre = []SeerGenre{
	{id:28,name:"Action"},
	{id:12,name:"Adventure"},
	{id:16,name:"Animation"},
	{id:35,name:"Comedy"},
	{id:80,name:"Crime"},
	{id:99,name:"Documentary"},
	{id:18,name:"Drama"},
	{id:10751,name:"Family"},
	{id:14,name:"Fantasy"},
	{id:36,name:"History"},
	{id:27,name:"Horror"},
	{id:10402,name:"Music"},
	{id:9648,name:"Mystery"},
	{id:10749,name:"Romance"},
	{id:878,name:"Science Fiction"},
	{id:10770,name:"TV Movie"},
	{id:53,name:"Thriller"},
	{id:10752,name:"War"},
	{id:37,name:"Western"},
}

func SeerrIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	outMeta := []db_types.MetadataEntry{}
	base := os.Getenv("SEERR_URL")
	if base == "" {
		return outMeta, errors.New("No seerr url configured")
	}

	key := os.Getenv("SEERR_KEY")
	if key == "" {
		return outMeta, errors.New("No seerr key configured")
	}

	url := fmt.Sprintf(
		"%s/api/v1/search?query=%s",
		base,
		url.PathEscape(info.Title),
	)

	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil {
		return outMeta, err
	}

	defer res.Body.Close()
	text, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return outMeta, fmt.Errorf("Request failed: %d, %s\n", res.StatusCode, string(text))
	}

	results := SeerrResults{}
	err = json.Unmarshal(text, &results)
	if err != nil {
		println(err.Error())
		return outMeta, err
	}

	for _, result := range results.Results {
		if result.MediaType != "movie" {
			continue
		}
		meta := db_types.MetadataEntry{}
		meta.ItemId = int64(result.Id)
		meta.Description = result.Overview
		meta.Title = result.Title
		meta.Thumbnail = fmt.Sprintf("https://image.tmdb.org/t/p/w300_and_h450_face%s", result.PosterPath)
		meta.Rating = result.VoteAverage
		meta.RatingMax = 10
		outMeta = append(outMeta, meta)
	}

	return outMeta, nil
}

func SeerrIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	outMeta := db_types.MetadataEntry{}
	base := os.Getenv("SEERR_URL")
	if base == "" {
		return outMeta, errors.New("No seerr url configured")
	}

	key := os.Getenv("SEERR_KEY")
	if key == "" {
		return outMeta, errors.New("No seerr key configured")
	}

	client := http.Client{}
	url := fmt.Sprintf(
		"%s/api/v1/movie/%s",
		base,
		id,
	)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil{
		return outMeta, err
	}

	if res.StatusCode != 200 {
		return outMeta, fmt.Errorf("Could not find id: %s", id)
	}

	result := SeerrResult{}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return outMeta, errors.New("Failed to read body")
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return outMeta, errors.New("Failed to parse body")
	}

	outMeta.ProviderID = strconv.Itoa(result.Id)
	outMeta.Provider = "seerr"
	outMeta.Thumbnail = fmt.Sprintf("https://image.tmdb.org/t/p/w300_and_h450_face%s", result.PosterPath)
	outMeta.Title = result.Title
	outMeta.Description = result.Overview
	outMeta.Rating = result.VoteAverage
	outMeta.RatingMax = 10

	if result.ReleaseDate != "" {
		dates := strings.Split(result.ReleaseDate, "-")
		if len(dates) == 1 {
			goto norelease
		}
		y, err := strconv.ParseInt(dates[0], 10, 64)
		if err != nil {
			goto norelease
		}
		outMeta.ReleaseYear = y
	}
norelease:
	if result.OriginalTitle != ""{
		outMeta.Native_Title = result.OriginalTitle
	}

	genres := []string{}
	for _, id := range result.GenreIds {
		for _, val := range seer_genres {
			if val.id == id {
				genres = append(genres, val.name)
			}
		}
	}
	genresStr, _ := json.Marshal(genres)
	outMeta.Genres = string(genresStr)

	return outMeta, nil
}
