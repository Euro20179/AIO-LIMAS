package settings

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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

func ExpandPathWithLocationAliases(aliases map[string]string, path string) string{
	for k, v := range aliases {
		path = strings.Replace(path, "${"+k+"}", v, 1)
	}
	return path
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
