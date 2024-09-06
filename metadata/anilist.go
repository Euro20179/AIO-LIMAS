package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Query     string       `json:"query"`
	Variables map[string]T `json:"variables"`
}
type AnilistTitles struct {
	English string
	Romaji  string
	Native  string
}
type AnlistMediaEntry struct {
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

	Type   string `json:"type"`
	Id     int64  `json:"id"`
	Format string `json:"format"`
}
type AnilistResponse struct {
	Data struct {
		Media AnlistMediaEntry `json:"media"`
	} `json:"data"`
}
type AnilistIdentifyResponse struct {
	Data struct {
		Page struct {
			Media []AnlistMediaEntry `json:"media"`
		} `json:"Page"`
	} `json:"data"`
}

const ANILIST_MEDIA_QUERY_INFO = `
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
	seasonYear,
	description,
	type,
	id,
	format
`

func AnilistManga(entry *db.InfoEntry, metadataEntry *db.MetadataEntry) (db.MetadataEntry, error) {
	var o db.MetadataEntry
	return o, nil
}

func mkAnilistRequest[T any, TOut any](anilistQuery AnilistQuery[T]) (TOut, error) {
	var out TOut
	bodyBytes, err := json.Marshal(anilistQuery)
	if err != nil {
		return out, err
	}
	bodyReader := bytes.NewReader(bodyBytes)
	res, err := http.Post("https://graphql.anilist.co", "application/json", bodyReader)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jData := new(TOut)
	err = json.Unmarshal(resp, &jData)
	if err != nil {
		return out, err
	}
	return *jData, nil
}

func AnilistShow(entry *db.InfoEntry, metadataEntry *db.MetadataEntry) (db.MetadataEntry, error) {
	var outMeta db.MetadataEntry

	searchTitle := entry.En_Title
	if searchTitle == "" {
		searchTitle = entry.Native_Title
	}
	query := fmt.Sprintf(`
		query ($search: String) {
			Media(search: $search, type: ANIME) {
				%s
			}
		}
	`, ANILIST_MEDIA_QUERY_INFO)
	anilistQuery := AnilistQuery[string]{
		Query: query,
		Variables: map[string]string{
			"search": searchTitle,
		},
	}

	jData, err := mkAnilistRequest[string, AnilistResponse](anilistQuery)
	if err != nil {
		return outMeta, err
	}

	out := jData.Data.Media

	mediaDependant := make(map[string]string)

	mediaDependant[fmt.Sprintf("%s-episodes", entry.Type)] = strconv.Itoa(int(out.Episodes))
	mediaDependant[fmt.Sprintf("%s-episode-duration", entry.Type)] = strconv.Itoa(int(out.Duration))
	mediaDependant[fmt.Sprintf("%s-length", entry.Type)] = strconv.Itoa(int(out.Episodes) * int(out.Duration))
	mediaDependant[fmt.Sprintf("%s-airing-status", entry.Type)] = out.Status

	mdString, _ := json.Marshal(mediaDependant)

	println(out.Title.Native, out.Title.English)

	if out.Title.Native != "" {
		outMeta.Native_Title = out.Title.Native
	}
	if out.Title.English != "" {
		outMeta.Title = out.Title.English
	} else if out.Title.Romaji != "" {
		outMeta.Title = out.Title.Romaji
	}
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

func AnilistIdentifier(info IdentifyMetadata) ([]db.MetadataEntry, error) {
	outMeta := []db.MetadataEntry{}

	searchTitle := info.Title
	query := fmt.Sprintf(`
		query ($search: String) {
			Page(page: 1) {
				media(search: $search) {
					%s
				}
			}
		}
	`, ANILIST_MEDIA_QUERY_INFO)
	anilistQuery := AnilistQuery[string]{
		Query: query,
		Variables: map[string]string{
			"search": searchTitle,
		},
	}
	jData, err := mkAnilistRequest[string, AnilistIdentifyResponse](anilistQuery)
	if err != nil {
		return outMeta, err
	}

	for _, entry := range jData.Data.Page.Media {
		var cur db.MetadataEntry
		cur.Thumbnail = entry.CoverImage.Large
		cur.Description = entry.Description
		cur.Rating = float64(entry.AverageScore)
		cur.ItemId = entry.Id
		if entry.Title.Native != "" {
			cur.Native_Title = entry.Title.Native
		}
		if entry.Title.English != "" {
			cur.Title = entry.Title.English
		} else if entry.Title.Romaji != "" {
			cur.Title = entry.Title.Romaji
		}
		cur.MediaDependant = "{\"provider\": \"anilist\"}"

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func AnilistById(id string) (db.MetadataEntry, error) {
	var outMeta db.MetadataEntry
	query := fmt.Sprintf(`
		query ($id: Int) {
			Media(id: $id) {
				%s
			}
		}
	`, ANILIST_MEDIA_QUERY_INFO)
	anilistQuery := AnilistQuery[string]{
		Query: query,
		Variables: map[string]string{
			"id": id,
		},
	}

	jData, err := mkAnilistRequest[string, AnilistResponse](anilistQuery)
	if err != nil {
		return outMeta, err
	}
	out := jData.Data.Media
	mediaDependant := make(map[string]string)

	var ty string
	switch out.Format {
	case "MOVIE":
		ty = "Movie"
	case "MANGA":
		ty = "Manga"
	case "SHOW":
		ty = "Show"
	}

	mediaDependant[fmt.Sprintf("%s-episodes", ty)] = strconv.Itoa(int(out.Episodes))
	mediaDependant[fmt.Sprintf("%s-episode-duration", ty)] = strconv.Itoa(int(out.Duration))
	mediaDependant[fmt.Sprintf("%s-length", ty)] = strconv.Itoa(int(out.Episodes) * int(out.Duration))
	mediaDependant[fmt.Sprintf("%s-airing-status", ty)] = out.Status

	mdString, _ := json.Marshal(mediaDependant)
	outMeta.Thumbnail = out.CoverImage.Large
	outMeta.Rating = float64(out.AverageScore)
	outMeta.Description = out.Description
	outMeta.MediaDependant = string(mdString)
	outMeta.ReleaseYear = out.SeasonYear
	return outMeta, nil
}
