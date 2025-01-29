package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	urlUtil "net/url"

	"aiolimas/settings"
	db_types "aiolimas/types"

)

func SonarrProvider(info *db_types.InfoEntry) (db_types.MetadataEntry, error) {
	//https://api-docs.overseerr.dev/#/
	var out db_types.MetadataEntry

	url := settings.Settings.SonarrURL
	key := settings.Settings.SonarrKey

	fullUrl := url + "api/v3/series/lookup"

	query := info.En_Title
	if query == "" {
		query = info.Native_Title
	}
	if query == "" {
		println("No search possible")
		return out, errors.New("No search possible")
	}

	fullUrl += "?term=" + urlUtil.QueryEscape(query)

	client := http.Client {}
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		println(err.Error())
		return out, err
	}

	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil {
		println(err.Error())
		return out, err
	}

	var all []map[string]interface{}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		println(err.Error())
		return out, err
	}

	err = json.Unmarshal(text, &all)
	if err != nil {
		println(err.Error())
		return out, err
	}

	data := all[0]

	idPretense, ok := data["id"]
	//this will be nil, if its not in the user's library
	if !ok {
		//second best
		return OMDBProvider(info)
	}

	id := uint64(idPretense.(float64))

	out, err = SonarrIdIdentifier(fmt.Sprintf("%d", id))
	if err != nil {
		println(err.Error())
		return out, err
	}
	out.ItemId = info.ItemId
	info.Location = data["path"].(string)
	return out, err
}

func SonarrIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	url := settings.Settings.SonarrURL
	key := settings.Settings.SonarrKey

	fullUrl := url + "api/v3/series/lookup"

	query := info.Title
	fullUrl += "?term=" + urlUtil.QueryEscape(query)

	client := http.Client {}
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	var all []map[string]interface{}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	err = json.Unmarshal(text, &all)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	outMeta := []db_types.MetadataEntry{}
	for _, entry := range all {
		var cur db_types.MetadataEntry
		idPretense, ok := entry["id"]
		if !ok { continue }
		id := int64(idPretense.(float64))
		cur.ItemId = id
		cur.Title = entry["title"].(string)
		images := entry["images"].([]interface{})
		posterImg := images[1].(map[string]interface{})
		cur.Thumbnail = posterImg["remoteUrl"].(string)

		cur.Provider = "sonarr"
		cur.ProviderID = fmt.Sprintf("%d", id)

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func SonarrIdIdentifier(id string) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	key := settings.Settings.SonarrKey
	url := settings.Settings.SonarrURL

	fullUrl := url + "api/v3/series/" + id

	client := http.Client {}
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		println(err.Error())
		return out, err
	}

	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil {
		println(err.Error())
		return out, err
	}

	var data map[string]interface{}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		println(err.Error())
		return out, err
	}

	err = json.Unmarshal(text, &data)
	if err != nil {
		println(err.Error())
		return out, err
	}

	out.Title = data["title"].(string)
	out.Description = data["overview"].(string)
	images := data["images"].([]interface{})
	posterImg := images[1].(map[string]interface{})
	out.Thumbnail = posterImg["remoteUrl"].(string)
	out.ReleaseYear = int64(data["year"].(float64))
	ratings := data["ratings"].(map[string]interface{})
	out.Rating = ratings["value"].(float64)
	out.RatingMax = 10

	seasonsArray := data["seasons"].([]interface{})

	specialsCount := float64(0)

	season0 := seasonsArray[0].(map[string]interface{})
	if season0["seasonNumber"].(float64) == 0 {
		season0Stats := season0["statistics"].(map[string]interface{})
		specialsCount = season0Stats["totalEpisodeCount"].(float64)
	}

	stats := data["statistics"].(map[string]interface{})
	totalEpisodes := stats["totalEpisodeCount"].(float64) - specialsCount
	airingStatus := data["status"].(string)
	episodeDuration := data["runtime"].(float64)

	mediaDependant := map[string]string {
		"Show-airing-status": airingStatus,
		"Show-episode-duration": fmt.Sprintf("%0.2f", episodeDuration),
		"Show-length": fmt.Sprintf("%0.2f", episodeDuration * totalEpisodes),
		"Show-episodes": fmt.Sprintf("%0.2f", totalEpisodes),
	}

	mdMarshal, err := json.Marshal(mediaDependant)

	if err != nil {
		println(err.Error())
		return out, err
	}

	out.MediaDependant = string(mdMarshal)

	return out, nil
}
