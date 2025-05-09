package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"aiolimas/settings"
	db_types "aiolimas/types"
	"aiolimas/logging"

)

func RadarrProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry

	us, err := settings.GetUserSettings(info.Uid)
	if err != nil{
		return out, err
	}

	url := us.RadarrURL
	key := us.RadarrKey

	if url == "" || key == ""{
		return out, errors.New("radarr is not setup")
	}

	fullUrl := url + "api/v3/movie/lookup"

	query := info.Entry.En_Title
	if query == "" {
		query = info.Entry.Native_Title
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
		//second best
		return OMDBProvider(info)
	}

	id := uint64(idPretense.(float64))

	out, err = RadarrIdIdentifier(fmt.Sprintf("%d", id), us)
	if err != nil {
		logging.ELog(err)
		return out, err
	}
	out.ItemId = info.Entry.ItemId
	info.Entry.Location = data["path"].(string)
	return out, err
}

func RadarrIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	us, err := settings.GetUserSettings(info.ForUid)
	if err != nil{
		return nil, err
	}

	url := us.RadarrURL
	key := us.RadarrKey

	if url == "" || key == ""{
		return nil, errors.New("radarr is not setup")
	}

	fullUrl := url + "api/v3/movie/lookup"

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
		posterImg := images[0].(map[string]interface{})
		cur.Thumbnail = posterImg["remoteUrl"].(string)

		cur.Provider = "sonarr"
		cur.ProviderID = fmt.Sprintf("%d", id)

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func RadarrIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	url := us.RadarrURL
	key := us.RadarrKey

	if url == "" || key == ""{
		return out, errors.New("radarr is not setup")
	}

	fullUrl := url + "api/v3/movie/" + id

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
	out.Native_Title = data["originalTitle"].(string)
	out.Description = data["overview"].(string)
	images := data["images"].([]interface{})
	posterImg := images[0].(map[string]interface{})
	out.Thumbnail = posterImg["remoteUrl"].(string)
	out.ReleaseYear = int64(data["year"].(float64))
	ratings := data["ratings"].(map[string]interface{})
	imdbRatings := ratings["imdb"].(map[string]interface{})
	out.Rating = imdbRatings["value"].(float64)
	out.RatingMax = 10

	mediaDependant := map[string]string {
		"Movie-length": fmt.Sprintf("%0.2f", data["runtime"].(float64)),
		"Movie-radarrid": fmt.Sprintf("%d", id),
	}

	mdMarshal, err := json.Marshal(mediaDependant)

	if err != nil {
		logging.ELog(err)
		return out, err
	}

	out.MediaDependant = string(mdMarshal)

	return out, nil
}
