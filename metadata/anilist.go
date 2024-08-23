package metadata

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"aiolimas/db"
)

type AnilistStatus string

const (
	SS_FINISHED         = "FINISHED"
	SS_RELEASING        = "RELEASING"
	SS_NOT_YET_RELEASED = "NOT_YET_RELEASED"
	SS_CANCELLED        = "CANCELLED"
	SS_HIATUS           = "HIATUS"
)

type AnilistQuery[T any] struct {
	Query     string   `json:"query"`
	Variables map[string]T `json:"variables"`
}
type AnilistTitles struct {
	English string
	Romaji  string
	Native  string
}
type AnilistResponse struct {
	Data struct {
		Media struct {
			Title AnilistTitles `json:"title"`

			CoverImage struct {
				Medium string `json:"medium"`
				Large  string `json:"large"`
			} `json:"coverImage"`

			AverageScore uint64 `json:"averageScore"`

			Description string `json:"description"`

			Duration   uint  `json:"duration"`
			Episodes   uint  `json:"episodes"`
			SeasonYear int64 `json:"seasonYear"`

			Status string `json:"status"`
		} `json:"media"`
	} `json:"data"`
}

func AnilistManga(entry *db.InfoEntry, metadataEntry *db.MetadataEntry) (db.MetadataEntry, error) {
	var o db.MetadataEntry
	return o, nil
}

func AnilistShow(entry *db.InfoEntry, metadataEntry *db.MetadataEntry) (db.MetadataEntry, error) {
	var outMeta db.MetadataEntry

	searchTitle := entry.En_Title
	if searchTitle == "" {
		searchTitle = entry.Native_Title
	}
	query := `
		query ($search: String) {
			Media(search: $search, type: ANIME) {
				title {
					english
					romaji
					native
				},
				coverImage {
					large
				},
				averageScore,
				duration,
				episodes,
				seasonYear
			}
		}
	`
	anilistQuery := AnilistQuery[string]{
		Query:     query,
		Variables: map[string]string {
			"search": searchTitle,
		},
	}
	bodyBytes, err := json.Marshal(anilistQuery)
	if err != nil {
		return outMeta, err
	}
	bodyReader := bytes.NewReader(bodyBytes)
	res, err := http.Post("https://graphql.anilist.co", "application/json", bodyReader)
	if err != nil {
		return outMeta, err
	}

	defer res.Body.Close()

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return outMeta, err
	}

	jData := new(AnilistResponse)
	err = json.Unmarshal(resp, &jData)
	if err != nil {
		return outMeta, err
	}

	out := jData.Data.Media

	mediaDependant := make(map[string]string)

	if out.Title.Native != "" {
		entry.Native_Title = out.Title.Native
	}
	if out.Title.English != "" {
		entry.En_Title = out.Title.English
	} else if out.Title.Romaji != "" {
		entry.En_Title = out.Title.Romaji
	}

	mediaDependant["Show-episodes"] = strconv.Itoa(int(out.Episodes))
	mediaDependant["Show-episode-duration"] = strconv.Itoa(int(out.Episodes))
	mediaDependant["Show-length"] = strconv.Itoa(int(out.Episodes))
	mediaDependant["Show-airing-status"] = out.Status

	mdString, _ := json.Marshal(mediaDependant)

	outMeta.Thumbnail = out.CoverImage.Large
	outMeta.Rating = float64(out.AverageScore)
	outMeta.Description = out.Description
	outMeta.MediaDependant = string(mdString)
	outMeta.ReleaseYear = out.SeasonYear

	return outMeta, nil
}

func AnlistProvider(entry *db.InfoEntry, metadataEntry *db.MetadataEntry) (db.MetadataEntry, error) {
	var newMeta db.MetadataEntry
	var err error
	if entry.IsAnime {
		newMeta, err = AnilistShow(entry, metadataEntry)
	} else {
		newMeta, err = AnilistManga(entry, metadataEntry)
	}

	// ensure item ids are consistent
	newMeta.ItemId = entry.ItemId

	return newMeta, err
}
