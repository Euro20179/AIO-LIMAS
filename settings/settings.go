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
}

var SettingsChannel chan [2]string

var Settings SettingsData

func InitSettingsManager(aioDir string) {
	SettingsChannel = make(chan [2]string)

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
			Settings.SonarrURL = value
		case "SonarrKey":
			Settings.SonarrKey = value
		case "RadarrURL":
			Settings.RadarrURL = value
		case "RadarrKey":
			Settings.RadarrKey = value
		}
	}
}
