package settings

import (
	"encoding/json"
	"fmt"
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

	DefaultTimeZone string
}

func GetUserSettigns(uid int64) (SettingsData, error) {
	settingsFile := os.Getenv("AIO_DIR") + fmt.Sprintf("/users/%d/settings.json", uid)

	file, err := os.Open(settingsFile)
	if err != nil {
		return SettingsData{}, nil
	}
	text, err := io.ReadAll(file)
	if err != nil {
		return SettingsData{}, err
	}

	var settings SettingsData

	err = json.Unmarshal(text, &settings)
	if err != nil {
		return SettingsData{}, err
	}

	return settings, nil
}

func InitUserSettings(uid int64) error {
	settingsFile := os.Getenv("AIO_DIR") + fmt.Sprintf("/users/%d/settings.json", uid)

	file, err := os.OpenFile(settingsFile, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	file.Write([]byte("{}"))
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
