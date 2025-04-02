package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"aiolimas/settings"
	db_types "aiolimas/types"

)

func RadarrProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry

	url := settings.Settings.RadarrURL
	key := settings.Settings.RadarrKey

	fullUrl := url + "api/v3/movie/lookup"

	query := info.Entry.En_Title
	if query == "" {
		query = info.Entry.Native_Title
	}
	if query == "" {
		println("No search possible")
		return out, errors.New("no search possible")
	}

	all, err := Lookup(query, fullUrl, key)

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

	out, err = RadarrIdIdentifier(fmt.Sprintf("%d", id))
	if err != nil {
		println(err.Error())
		return out, err
	}
	out.ItemId = info.Entry.ItemId
	info.Entry.Location = data["path"].(string)
	return out, err
}

func RadarrIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	url := settings.Settings.RadarrURL
	key := settings.Settings.RadarrKey

	fullUrl := url + "api/v3/movie/lookup"

	query := info.Title

	all, err := Lookup(query, fullUrl, key)

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
		posterImg := images[0].(map[string]interface{})
		cur.Thumbnail = posterImg["remoteUrl"].(string)

		cur.Provider = "sonarr"
		cur.ProviderID = fmt.Sprintf("%d", id)

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func RadarrIdIdentifier(id string) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	url := settings.Settings.RadarrURL
	key := settings.Settings.RadarrKey

	fullUrl := url + "api/v3/movie/" + id

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
		println(err.Error())
		return out, err
	}

	out.MediaDependant = string(mdMarshal)

	return out, nil
}
