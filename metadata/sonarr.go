package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"aiolimas/logging"
	"aiolimas/settings"
	db_types "aiolimas/types"
)

func SonarrGetLocation(us *settings.SettingsData, providerID string) (string, error) {
	url := us.SonarrURL
	key := us.SonarrKey

	if url == "" || key == ""{
		return "", errors.New("sonarr is not setup")
	}

	id := providerID

	if id == "" {
		return "", errors.New("ProviderID is not set")
	}

	fId, _ := strconv.ParseFloat(id, 64)
	path, err := LookupPathById(fId, url, key)

	if err != nil {
		return "", err
	}

	return path, nil
}

func SonarrProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	entry := info.Entry
	var out db_types.MetadataEntry

	us, err := settings.GetUserSettings(info.Uid)
	if err != nil{
		return out, err
	}
	url := us.SonarrURL
	key := us.SonarrKey

	if url == "" || key == ""{
		return out, errors.New("sonarr is not setup")
	}

	fullUrl := url + "api/v3/series/lookup"

	query := entry.En_Title
	if query == "" {
		query = entry.Native_Title
	}
	if query == "" {
		logging.Info("no search possible")
		return out, errors.New("no search possible")
	}

	all, err := Lookup(query, fullUrl, key)

	if err != nil {
		logging.ELog(err)
		return out, err
	}

	data := all[0]

	idPretense, ok := data["id"]
	//this will be nil, if its not in the user's library
	if !ok {
		if entry.IsAnime() {
			return AnilistShow(info)
		}
		return OMDBProvider(info)
	}

	id := uint64(idPretense.(float64))

	out, err = SonarrIdIdentifier(fmt.Sprintf("%d", id), us)
	if err != nil {
		logging.ELog(err)
		return out, err
	}
	out.ItemId = info.Entry.ItemId

	info.Entry.Location = data["path"].(string)
	return out, err
}

func SonarrIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	us, err := settings.GetUserSettings(info.ForUid)
	if err != nil{
		return nil, err
	}
	url := us.SonarrURL
	key := us.SonarrKey

	if url == "" || key == ""{
		return nil, errors.New("sonarr is not setup")
	}

	fullUrl := url + "api/v3/series/lookup"

	query := info.Title

	all, err := Lookup(query, fullUrl, key)

	if err != nil {
		logging.ELog(err)
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

func SonarrIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	key := us.SonarrKey
	url := us.SonarrURL

	if url == "" || key == ""{
		return out, errors.New("sonarr is not setup")
	}

	fullUrl := url + "api/v3/series/" + id

	client := http.Client {}
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		logging.ELog(err)
		return out, err
	}

	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil {
		logging.ELog(err)
		return out, err
	}

	var data map[string]interface{}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		logging.ELog(err)
		return out, err
	}

	err = json.Unmarshal(text, &data)
	if err != nil {
		logging.ELog(err)
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
	out.Provider = "sonarr"
	out.ProviderID = id

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
		"Show-sonarrid": id,
	}

	mdMarshal, err := json.Marshal(mediaDependant)

	if err != nil {
		logging.ELog(err)
		return out, err
	}

	out.MediaDependant = string(mdMarshal)

	return out, nil
}
