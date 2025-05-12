package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"aiolimas/logging"
	"aiolimas/settings"
	db_types "aiolimas/types"

	"golang.org/x/net/html"
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

func getSteamSearchTree(search string) (*html.Node, error){
	baseUrl := "https://store.steampowered.com/search/suggest?term=%s&f=games&cc=US&use_search_spellcheck=1"

	fullUrl := fmt.Sprintf(baseUrl, url.QueryEscape(search))

	resp, err := http.Get(fullUrl)
	if err != nil {
		return nil, err
	}

	tree, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func SteamProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry

	title := info.Entry.En_Title
	if title == "" {
		title = info.Entry.Native_Title
	}

	if title == "" {
		logging.Info("no search possible")
		return out, errors.New("no search possible")
	}

	tree, _ := getSteamSearchTree(title)

	us, err := settings.GetUserSettings(info.Uid)
	if err != nil {
		return out, err
	}

	for n := range tree.FirstChild.Descendants() {
		if n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "data-ds-appid" {
					return SteamIdIdentifier(attr.Val, us)
				}
			}
		}
	}
	out.Provider = "steam"

	return out, errors.New("no results")
}

func SteamIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	out := []db_types.MetadataEntry{}

	title := info.Title

	if title == "" {
		logging.Info("no search possible")
		return out, errors.New("no search possible")
	}

	tree, err := getSteamSearchTree(title)
	if err != nil {
		return out, err
	}

	var cur db_types.MetadataEntry

	nextIsName := false
	for n := range tree.FirstChild.Descendants() {
		if n.Data == "a" {
			if cur.ProviderID != ""{
				out = append(out, cur)
			}

			cur = db_types.MetadataEntry{}
			cur.Provider = "steam"

			for _, attr := range n.Attr {
				if attr.Key == "data-ds-appid" {
					i, err := strconv.ParseInt(attr.Val, 10, 64)
					if err != nil {
						break
					}
					cur.ItemId = i
					cur.ProviderID = attr.Val
				}
			}
		} else if n.Data == "img" {
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					cur.Thumbnail = attr.Val
				}
			}
		} else if n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Val == "match_name" {
					nextIsName = true
					break
				}
			}
		} else if nextIsName {
			cur.Title = n.Data
			nextIsName = false
		}
	}

	return out, nil
}

func SteamIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
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

	mainData := respJson[i]

	out.Title = mainData.Data.Name
	out.Description = mainData.Data.DetailedDescription
	out.Provider = "steam"
	out.ProviderID = id
	out.Thumbnail = fmt.Sprintf("http://cdn.origin.steamstatic.com/steam/apps/%s/library_600x900_2x.jpg", url.PathEscape(id))

	if !mainData.Data.ReleaseDate.ComingSoon {
		dateInfo := mainData.Data.ReleaseDate.Date
		dmy := strings.Split(dateInfo, " ")
		year := dmy[len(dmy)-1]
		yearI, err := strconv.ParseInt(year, 10, 64)
		if err != nil {
			return out, err
		}
		out.ReleaseYear = yearI
	}

	reviewsUrl := fmt.Sprintf("https://store.steampowered.com/appreviews/%s?json=1", url.QueryEscape(id))
	res, err = http.Get(reviewsUrl)
	if err != nil {
		return out, err
	}
	text, err = io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	var reviewsRespJson map[string]any
	json.Unmarshal(text, &reviewsRespJson)
	summary := reviewsRespJson["query_summary"].(map[string]any)
	score := summary["review_score"].(float64)
	out.Rating = score
	out.RatingMax = 10

	return out, nil
}

func SteamLocationFinder(_ *settings.SettingsData, providerID string) (string, error) {
	if providerID != "" {
		return fmt.Sprintf("steam://rungameid/%s", providerID), nil
	}

	return "", errors.New("please set the metadata ProviderID before setting the location")
}
