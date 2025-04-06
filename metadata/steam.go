package metadata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"io"

	db_types "aiolimas/types"
)

type IdAppData struct {
	Success bool
	Data    struct {
		Type                 string
		Name                 string
		SteamAppId           int64 `json:"steam_appid"`
		RequiredAge          int64 `json:"required_age"`
		IsFree               bool  `json:"is_free"`
		Dlc                  any
		DetailedDescription  string `json:"detailed_description"`
		AboutTheGame         string `json:"about_the_game"`
		ShortDescription     string `json:"short_description"`
		SupportedLanguages   string `json:"supported_languages"`
		HeaderImage          string `json:"header_image"`
		CapsuleImage         string `json:"capsule_image"`
		CapsuleImageV5       string `json:"capsule_imagev5"`
		Website              string
		PcRequirements       any    `json:"pc_requirements"`
		MacRequirements      any    `json:"mac_requirements"`
		LegalNotice          string `json:"legal_notice"`
		ExtUserAccountNotice string `json:"ext_user_account_notice"`
		Developers           any
		Publishers           any
		PackageGroups        any `json:"package_groups"`
		Platforms            any
		Categories           any
		Genres               any
		Screenshots          any
		Movies               any
		ReleaseDate          struct {
			ComingSoon bool `json:"coming_soon"`
			Date       string
		} `json:"release_date"`
		SupportInfo        any `json:"support_info"`
		Background         string
		BackgroundRaw      string `json:"background_raw"`
		ContentDescriptors any    `json:"content_descriptors"`
		Ratins             any
	}
}

func SteamIdIdentifier(id string) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return out, err
	}

	fullUrl := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%s", url.QueryEscape(id))

	res, err := http.Get(fullUrl)
	if err != nil {
		return out, err
	}

	text, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	var respJson map[int64]IdAppData
	json.Unmarshal(text, &respJson)

	data := respJson[i]

	out.Title = data.Data.Name
	out.Description = data.Data.DetailedDescription
	out.Provider = "steam"
	out.ProviderID = id
	out.Thumbnail = fmt.Sprintf("http://cdn.origin.steamstatic.com/steam/apps/%s/library_600x900_2x.jpg", url.PathEscape(id))

	if !data.Data.ReleaseDate.ComingSoon {
		dateInfo := data.Data.ReleaseDate.Date
		dmy := strings.Split(dateInfo, " ")
		year := dmy[len(dmy) - 1]
		yearI, err := strconv.ParseInt(year, 10, 64)
		if err != nil{
			return out, err
		}
		out.ReleaseYear = yearI
	}

	return out, nil
}
