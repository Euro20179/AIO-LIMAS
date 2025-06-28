package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"aiolimas/logging"
)

/*
	This file is for common functions between the *arr providers
	for now that is sonarr, and radarr
*/

func _request(fullUrl string, key string) ([]byte, error) {
	client := http.Client {}

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		logging.ELog(err)
		return nil, err
	}

	req.Header.Set("X-Api-Key", key)
	res, err := client.Do(req)
	if err != nil {
		logging.ELog(err)
		return nil, err
	}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		logging.ELog(err)
		return nil, err
	}

	return text, nil
}

func LookupPathById(id float64, apiPath string, key string) (string, error) {
	fullUrl := apiPath + "api/v3/series"

	text, err := _request(fullUrl, key)
	if err != nil {
		logging.ELog(err)
		return "", err
	}

	var all []map[string]interface{}
	err = json.Unmarshal(text, &all)
	if err != nil {
		logging.ELog(err)
		return "", err
	}

	for _, item := range all {
		curId := item["id"].(float64)
		if curId == id {
			return item["path"].(string), nil
		}
	}
	return "", errors.New("could not find item")
}

func Lookup(query string, apiPath string, key string) ([]map[string]interface{}, error){
	fullUrl := apiPath + "?term=" + url.QueryEscape(query)

	var all []map[string]interface{}

	text, err := _request(fullUrl, key)
	if err != nil {
		logging.ELog(err)
		return nil, err
	}

	err = json.Unmarshal(text, &all)
	if err != nil {
		logging.ELog(err)
		return nil, err
	}

	return all, nil
}
