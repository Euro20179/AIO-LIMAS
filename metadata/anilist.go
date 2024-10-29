package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"aiolimas/types"
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
		Medium     string `json:"medium"`
		Large      string `json:"large"`
		ExtraLarge string `json:"extraLarge"`
	} `json:"coverImage"`

	AverageScore uint64 `json:"averageScore"`

	Description string `json:"description"`

	Duration  uint `json:"duration"`
	Episodes  uint `json:"episodes"`
	StartDate struct {
		Year int `json:"year"`
	} `json:"startDate"`

	Status string `json:"status"`

	Type       string `json:"type"`
	Id         int64  `json:"id"`
	Format     string `json:"format"`
	Volumes    int    `json:"volumes"`
	SeasonYear int    `json:"seasonYear"`
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
		large,
		extraLarge
	},
	startDate {
		year
	},
	averageScore,
	duration,
	episodes,
	seasonYear,
	description,
	type,
	id,
	format,
	volumes
`

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

func applyManga(anilistData AnlistMediaEntry) (db_types.MetadataEntry, error) {
	var o db_types.MetadataEntry
	out := anilistData

	mediaDependant := make(map[string]string)
	mediaDependant["Manga-volumes"] = fmt.Sprintf("%d", out.Volumes)
	mdString, _ := json.Marshal(mediaDependant)

	o.MediaDependant = string(mdString)

	if out.Title.English != ""{
		o.Title = out.Title.English
	} else if out.Title.Romaji != "" {
		o.Title = out.Title.Romaji
	}

	o.Native_Title = out.Title.Native
	if out.CoverImage.ExtraLarge != "" {
		o.Thumbnail = out.CoverImage.ExtraLarge
	} else {
		o.Thumbnail = out.CoverImage.Large
	}
	o.ReleaseYear = int64(out.StartDate.Year)
	o.Description = out.Description
	o.ItemId = out.Id
	o.Rating = float64(out.AverageScore)
	o.RatingMax = 100

	o.Provider = "anilist"
	o.ProviderID = fmt.Sprintf("%d", out.Id)

	return o, nil
}

func AnilistManga(entry *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry) (db_types.MetadataEntry, error) {
	var o db_types.MetadataEntry
	searchTitle := entry.En_Title
	if searchTitle == "" {
		searchTitle = entry.Native_Title
	}

	query := fmt.Sprintf(`
		query ($search: String) {
			Media(search: $search, type: MANGA) {
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
		return o, err
	}

	return applyManga(jData.Data.Media)
}

func applyShow(aniInfo AnlistMediaEntry) (db_types.MetadataEntry, error) {
	var outMeta db_types.MetadataEntry
	mediaDependant := make(map[string]string)

	var ty string
	switch aniInfo.Format {
	case "MOVIE":
		ty = "Movie"
	case "MANGA":
		ty = "Manga"
	default:
		ty = "Show"
	}

	mediaDependant[fmt.Sprintf("%s-episodes", ty)] = strconv.Itoa(int(aniInfo.Episodes))
	mediaDependant[fmt.Sprintf("%s-episode-duration", ty)] = strconv.Itoa(int(aniInfo.Duration))
	mediaDependant[fmt.Sprintf("%s-length", ty)] = strconv.Itoa(int(aniInfo.Episodes) * int(aniInfo.Duration))
	mediaDependant[fmt.Sprintf("%s-airing-status", ty)] = aniInfo.Status

	mdString, _ := json.Marshal(mediaDependant)

	if aniInfo.Title.Native != "" {
		outMeta.Native_Title = aniInfo.Title.Native
	}
	if aniInfo.Title.English != "" {
		outMeta.Title = aniInfo.Title.English
	} else if aniInfo.Title.Romaji != "" {
		outMeta.Title = aniInfo.Title.Romaji
	}
	// println(aniInfo.StartDate.Year)
	outMeta.Thumbnail = aniInfo.CoverImage.ExtraLarge
	outMeta.Rating = float64(aniInfo.AverageScore)
	outMeta.RatingMax = 100
	outMeta.Description = aniInfo.Description
	outMeta.MediaDependant = string(mdString)
	outMeta.ReleaseYear = int64(aniInfo.StartDate.Year)
	outMeta.Provider = "anilist"
	outMeta.ProviderID = fmt.Sprintf("%d", aniInfo.Id)

	return outMeta, nil
}

func AnilistShow(entry *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry) (db_types.MetadataEntry, error) {
	var outMeta db_types.MetadataEntry

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
	return applyShow(out)
}

func AnlistProvider(entry *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry) (db_types.MetadataEntry, error) {
	var newMeta db_types.MetadataEntry
	var err error
	if entry.ArtStyle & db_types.AS_ANIME == db_types.AS_ANIME{
		newMeta, err = AnilistShow(entry, metadataEntry)
	} else {
		newMeta, err = AnilistManga(entry, metadataEntry)
	}

	// ensure item ids are consistent
	newMeta.ItemId = entry.ItemId

	return newMeta, err
}

func AnilistIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	outMeta := []db_types.MetadataEntry{}

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
		var cur db_types.MetadataEntry
		if entry.Type == "MANGA" {
			cur, err = applyManga(entry)
			if err != nil {
				println(err.Error())
				continue
			}
		} else {
			cur, err = applyShow(entry)
			if err != nil {
				println(err.Error())
				continue
			}
		}
		cur.ItemId = entry.Id

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func AnilistById(id string) (db_types.MetadataEntry, error) {
	var outMeta db_types.MetadataEntry
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
	// mediaDependant := make(map[string]string)

	var ty string
	switch out.Format {
	case "MOVIE":
		ty = "Movie"
	case "MANGA":
		ty = "Manga"
	case "SHOW":
		ty = "Show"
	}

	if ty == "Manga" {
		outMeta, err = applyManga(out)
		if err != nil {
			println(err.Error())
			return outMeta, err
		}
	} else {
		outMeta, err = applyShow(out)
		if err != nil {
			println(err.Error())
			return outMeta, err
		}
	}

	outMeta.ItemId = out.Id

	return outMeta, nil
}
