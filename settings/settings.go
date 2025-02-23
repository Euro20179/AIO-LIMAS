package settings

import (
	"encoding/json"
	"io"
	"os"
)

type SettingsData struct {
	SonarrURL string
	SonarrKey string
	RadarrURL string
	RadarrKey string

	WriteIdFile bool

	LocationAliases map[string]string
}

var SettingsChannel chan [2]any

var Settings SettingsData

func InitSettingsManager(aioDir string) {
	SettingsChannel = make(chan [2]any)

	settingsFile := aioDir + "/settings.json"

	if file, err := os.Open(settingsFile); err == nil {
		text, err := io.ReadAll(file)
		if err != nil {
			panic("Could not open settings file")
		}

		err = json.Unmarshal(text, &Settings)
		if err != nil {
			panic("Could not parse settings file")
		}

		return
	}

	file, err := os.OpenFile(settingsFile, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic("Failed to create settings file")
	}

	file.Write([]byte("{}"))
	if err := file.Close(); err != nil {
		panic("Failed to create settings file, writing {}")
	}
}

func ManageSettings() {
	for {
		msg := <-SettingsChannel

		key := msg[0]
		value := msg[1]

		switch key {
		case "SonarrURL":
			Settings.SonarrURL = value.(string)
		case "SonarrKey":
			Settings.SonarrKey = value.(string)
		case "RadarrURL":
			Settings.RadarrURL = value.(string)
		case "RadarrKey":
			Settings.RadarrKey = value.(string)
		case "WriteIdFile":
			Settings.WriteIdFile = value.(bool)
		case "LocationAliases":
			Settings.LocationAliases = value.(map[string]string)
		}
	}
}
