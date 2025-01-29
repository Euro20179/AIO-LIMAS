package metadata

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

/*
	This file is for common functions between the *arr providers
	for now that is sonarr, and radarr
*/

func Lookup(query string, apiPath string, key string) ([]map[string]interface{}, error){
	client := http.Client {}

	fullUrl := apiPath + "?term=" + url.QueryEscape(query)

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

	return all, nil
}
