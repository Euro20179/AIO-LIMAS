package metadata

import (
	db_types "aiolimas/types"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type SeerrResult struct {
	Id int
	MediaType string
	Adult bool
	GenreIds []int
	OriginalLanguage string
	OriginalTitle string
	Overview string
	Popularity float64
	ReleaseDate string
	Title string
	Video bool
	VoteAverage float64
	VoteCount int
	BackdropPath string
	PosterPath string
	MediaInfo SeerrMediaInfo
}

type SeerrMediaInfo struct {
	DownloadStatus []any
	DownloadStatus4k []any
	Id int
	MediaType string
	TmdbId int
	TvdbId int
	ImdbId string
	Status int
	Status4k int
	CreatedAt string
	UpdatedAt string
	LastSeasonChange string
	MediaAddedAt string
	ServiceId any
	ServiceId4k any
	ExternalServiceId any
	ExternalServiceId4k any
	ExternalServiceSlug any
	ExternalServiceSlug4k any
	RatingKey any
	RatingKey4k any
	JellyfinMediaId string
	JellyfinMediaId4k string
	WatchLists []any
	MediaUrl string
}

type SeerrResults struct {
	Page int
	TotalPages int
	TotalResults int
	Results []SeerrResult
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
		url.QueryEscape(info.Title),
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
