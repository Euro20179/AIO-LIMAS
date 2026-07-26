package metadata

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"aiolimas/logging"
	"aiolimas/settings"
	"aiolimas/types"
)

type OMDBError struct {
	Response string
	Error    string
}

type OMDBResponse struct {
	Title    string
	Year     string
	Rated    string
	Released string
	Runtime  string
	Genre    string
	Director string
	Writer   string
	Actors   string
	Plot     string
	Language string
	Country  string
	Awards   string
	Poster   string
	Ratings  []struct {
		Source string
		Value  string
	}
	Metascore    string
	ImdbRating   string `json:"imdbRating"`
	ImdbVotes    string `json:"imdbVotes"`
	ImdbID       string `json:"imdbID"`
	Type         string
	TotalSeasons string `json:"totalSeasons"`
	Response     string
	BoxOffice string
}

type OMDBSearchItem struct {
	Title  string
	Year   string
	ImdbID string `json:"imdbID"`
	Type   string
	Poster string
}

func countryNameToCountryCode(name string) string {
	mapping, has := map[string]string {
		"Japan": "JP",
		"Canada": "CA",
		"United States": "US",
		"USA": "US",
		"United Kingdom": "GB",
		"South Korea": "KR",
		"Netherlands": "NL",
		"Ireland": "IE",
		"France": "FR",
		"Belgium": "BE",
		"Sweden": "SE",
		"Denmark": "DK",
		"Luxembourg": "LU",
		"Brazil": "BR",
		"Italy": "IT",
		"New Zealand": "NZ",
		"Hungary": "HU",
		"Hong Kong": "HK",
		"China": "CN",
		"Ascension Island": "AC",
		"Andorra": "AD",
		"United Arab Emirates": "AE",
		"Afghanistan": "AF",
		"Antigua & Barbuda": "AG",
		"Anguilla": "AI",
		"Albania": "AL",
		"Armenia": "AM",
		"Angola": "AO",
		"Antarctica": "AQ",
		"Argentina": "AR",
		"American Samoa": "AS",
		"Austria": "AT",
		"Australia": "AU",
		"Aruba": "AW",
		"Åland Islands": "AX",
		"Azerbaijan": "AZ",
		"Bosnia & Herzegovina": "BA",
		"Barbados": "BB",
		"Bangladesh": "BD",
		"Burkina Faso": "BF",
		"Bulgaria": "BG",
		"Bahrain": "BH",
		"Burundi": "BI",
		"Benin": "BJ",
		"St. Barthélemy": "BL",
		"Bermuda": "BM",
		"Brunei": "BN",
		"Bolivia": "BO",
		"Caribbean Netherlands": "BQ",
		"Bahamas": "BS",
		"Bhutan": "BT",
		"Bouvet Island": "BV",
		"Botswana": "BW",
		"Belarus": "BY",
		"Belize": "BZ",
		"Cocos (Keeling) Islands": "CC",
		"Congo - Kinshasa": "CD",
		"Central African Republic": "CF",
		"Congo - Brazzaville": "CG",
		"Switzerland": "CH",
		"Côte d’Ivoire": "CI",
		"Cook Islands": "CK",
		"Chile": "CL",
		"Cameroon": "CM",
		"Colombia": "CO",
		"Clipperton Island": "CP",
		"Sark": "CQ",
		"Costa Rica": "CR",
		"Cuba": "CU",
		"Cape Verde": "CV",
		"Curaçao": "CW",
		"Christmas Island": "CX",
		"Cyprus": "CY",
		"Czechia": "CZ",
		"Germany": "DE",
		"Diego Garcia": "DG",
		"Djibouti": "DJ",
		"Dominica": "DM",
		"Dominican Republic": "DO",
		"Algeria": "DZ",
		"Ceuta & Melilla": "EA",
		"Ecuador": "EC",
		"Estonia": "EE",
		"Egypt": "EG",
		"Western Sahara": "EH",
		"Eritrea": "ER",
		"Spain": "ES",
		"Ethiopia": "ET",
		"European Union": "EU",
		"Finland": "FI",
		"Fiji": "FJ",
		"Falkland Islands": "FK",
		"Micronesia": "FM",
		"Faroe Islands": "FO",
		"Gabon": "GA",
		"Grenada": "GD",
		"Georgia": "GE",
		"French Guiana": "GF",
		"Guernsey": "GG",
		"Ghana": "GH",
		"Gibraltar": "GI",
		"Greenland": "GL",
		"Gambia": "GM",
		"Guinea": "GN",
		"Guadeloupe": "GP",
		"Equatorial Guinea": "GQ",
		"Greece": "GR",
		"South Georgia & South Sandwich Islands": "GS",
		"Guatemala": "GT",
		"Guam": "GU",
		"Guinea-Bissau": "GW",
		"Guyana": "GY",
		"Hong Kong SAR China": "HK",
		"Heard & McDonald Islands": "HM",
		"Honduras": "HN",
		"Croatia": "HR",
		"Haiti": "HT",
		"Canary Islands": "IC",
		"Indonesia": "ID",
		"Israel": "IL",
		"Isle of Man": "IM",
		"India": "IN",
		"British Indian Ocean Territory": "IO",
		"Iraq": "IQ",
		"Iran": "IR",
		"Iceland": "IS",
		"Jersey": "JE",
		"Jamaica": "JM",
		"Jordan": "JO",
		"Kenya": "KE",
		"Kyrgyzstan": "KG",
		"Cambodia": "KH",
		"Kiribati": "KI",
		"Comoros": "KM",
		"St. Kitts & Nevis": "KN",
		"North Korea": "KP",
		"Kuwait": "KW",
		"Cayman Islands": "KY",
		"Kazakhstan": "KZ",
		"Laos": "LA",
		"Lebanon": "LB",
		"St. Lucia": "LC",
		"Liechtenstein": "LI",
		"Sri Lanka": "LK",
		"Liberia": "LR",
		"Lesotho": "LS",
		"Lithuania": "LT",
		"Latvia": "LV",
		"Libya": "LY",
		"Morocco": "MA",
		"Monaco": "MC",
		"Moldova": "MD",
		"Montenegro": "ME",
		"St. Martin": "MF",
		"Madagascar": "MG",
		"Marshall Islands": "MH",
		"North Macedonia": "MK",
		"Mali": "ML",
		"Myanmar (Burma)": "MM",
		"Mongolia": "MN",
		"Macao SAR China": "MO",
		"Northern Mariana Islands": "MP",
		"Martinique": "MQ",
		"Mauritania": "MR",
		"Montserrat": "MS",
		"Malta": "MT",
		"Mauritius": "MU",
		"Maldives": "MV",
		"Malawi": "MW",
		"Mexico": "MX",
		"Malaysia": "MY",
		"Mozambique": "MZ",
		"Namibia": "NA",
		"New Caledonia": "NC",
		"Niger": "NE",
		"Norfolk Island": "NF",
		"Nigeria": "NG",
		"Nicaragua": "NI",
		"Norway": "NO",
		"Nepal": "NP",
		"Nauru": "NR",
		"Niue": "NU",
		"Oman": "OM",
		"Panama": "PA",
		"Peru": "PE",
		"French Polynesia": "PF",
		"Papua New Guinea": "PG",
		"Philippines": "PH",
		"Pakistan": "PK",
		"Poland": "PL",
		"St. Pierre & Miquelon": "PM",
		"Pitcairn Islands": "PN",
		"Puerto Rico": "PR",
		"Palestinian Territories": "PS",
		"Portugal": "PT",
		"Palau": "PW",
		"Paraguay": "PY",
		"Qatar": "QA",
		"Réunion": "RE",
		"Romania": "RO",
		"Serbia": "RS",
		"Russia": "RU",
		"Rwanda": "RW",
		"Saudi Arabia": "SA",
		"Solomon Islands": "SB",
		"Seychelles": "SC",
		"Sudan": "SD",
		"Singapore": "SG",
		"St. Helena": "SH",
		"Slovenia": "SI",
		"Svalbard & Jan Mayen": "SJ",
		"Slovakia": "SK",
		"Sierra Leone": "SL",
		"San Marino": "SM",
		"Senegal": "SN",
		"Somalia": "SO",
		"Suriname": "SR",
		"South Sudan": "SS",
		"São Tomé & Príncipe": "ST",
		"El Salvador": "SV",
		"Sint Maarten": "SX",
		"Syria": "SY",
		"Eswatini": "SZ",
		"Tristan da Cunha": "TA",
		"Turks & Caicos Islands": "TC",
		"Chad": "TD",
		"French Southern Territories": "TF",
		"Togo": "TG",
		"Thailand": "TH",
		"Tajikistan": "TJ",
		"Tokelau": "TK",
		"Timor-Leste": "TL",
		"Turkmenistan": "TM",
		"Tunisia": "TN",
		"Tonga": "TO",
		"Türkiye": "TR",
		"Trinidad & Tobago": "TT",
		"Tuvalu": "TV",
		"Taiwan": "TW",
		"Tanzania": "TZ",
		"Ukraine": "UA",
		"Uganda": "UG",
		"U.S. Outlying Islands": "UM",
		"United Nations": "UN",
		"Uruguay": "UY",
		"Uzbekistan": "UZ",
		"Vatican City": "VA",
		"St. Vincent & Grenadines": "VC",
		"Venezuela": "VE",
		"British Virgin Islands": "VG",
		"U.S. Virgin Islands": "VI",
		"Vietnam": "VN",
		"Vanuatu": "VU",
		"Wallis & Futuna": "WF",
		"Samoa": "WS",
		"Kosovo": "XK",
		"Yemen": "YE",
		"Mayotte": "YT",
		"South Africa": "ZA",
		"Zambia": "ZM",
		"Zimbabwe": "ZW",
	}[name]

	if has {
		return mapping
	}
	return name
}

func titleCase(st string) string {
	if st == "" {
		return st
	}
	return string(strings.ToTitle(st)[0]) + string(st[1:])
}

func omdbResultToMetadata(result OMDBResponse) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}
	mediaDep := make(map[string]string)
	if result.Type == "series" {
		duration := strings.Split(result.Runtime, " ")[0]

		mediaDep["Show-episode-duration"] = duration
		mediaDep["Show-imdbid"] = result.ImdbID
	} else {
		length := strings.Split(result.Runtime, " ")[0]

		mediaDep[titleCase(result.Type)+"-director"] = result.Director
		mediaDep[titleCase(result.Type)+"-length"] = length
		mediaDep[titleCase(result.Type)+"-imdbid"] = result.ImdbID
		if result.Rated != "" && result.Rated != "N/A" {
			mediaDep[titleCase(result.Type)+"-rating"] = result.Rated
		}
		if result.BoxOffice != "N/A" && result.BoxOffice != "" {
			mediaDep[titleCase(result.Type)+"-revenue"] = strings.Replace(strings.ReplaceAll(result.BoxOffice, ",", ""), "$", "", 1)
		}
	}

	if result.ImdbRating != "N/A" {
		res, err := strconv.ParseFloat(result.ImdbRating, 64)
		if err == nil {
			out.Rating = res
		}
	}
	out.RatingMax = 10

	mdStr, err := json.Marshal(mediaDep)
	if err != nil {
		return out, err
	}

	out.MediaDependant = string(mdStr)

	out.Description = result.Plot
	out.Thumbnail = result.Poster

	out.Provider = "omdb"
	out.ProviderID = result.ImdbID[2:]

	countries := strings.Split(result.Country, ",")
	correctCountries := []string{}
	for _, country := range countries {
		correctCountries = append(correctCountries, countryNameToCountryCode(country))
	}
	out.Country = strings.Join(correctCountries, ",")

	out.Title = result.Title

	genres := strings.Split(result.Genre, ", ")
	genreList, err := json.Marshal(genres)
	if err == nil {
		out.Genres = string(genreList)
	} else {
		logging.ELog(err)
		out.Genres = ""
	}

	yearSep := "–"
	if strings.Contains(result.Year, yearSep) {
		result.Year = strings.Split(result.Year, yearSep)[0]
	}

	n, err := strconv.ParseInt(result.Year, 10, 64)
	if err == nil {
		out.ReleaseYear = n
	} else {
		logging.ELog(err)
	}

	return out, nil
}

func OMDBProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	entry := info.Entry
	var out db_types.MetadataEntry

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return out, errors.New("no api key")
	}

	search := entry.En_Title
	if search == "" {
		search = entry.Native_Title
	}
	if search == "" {
		return out, errors.New("no search possible")
	}

	url := fmt.Sprintf(
		"https://www.omdbapi.com/?apikey=%s&t=%s",
		key,
		url.QueryEscape(search),
	)

	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	if bytes.Contains(body, []byte("\"Response\":\"False\"")) {
		var omdbErr OMDBError
		json.Unmarshal(body, &omdbErr)
		return out, errors.New(omdbErr.Error)
	}

	jData := new(OMDBResponse)
	err = json.Unmarshal(body, &jData)
	if err != nil {
		return out, err
	}

	return omdbResultToMetadata(*jData)
}

func OmdbIdentifier(info IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	outMeta := []db_types.MetadataEntry{}

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return outMeta, errors.New("No api key")
	}

	searchTitle := info.Title
	url := fmt.Sprintf(
		"https://www.omdbapi.com/?apikey=%s&s=%s",
		key,
		url.QueryEscape(searchTitle),
	)

	res, err := http.Get(url)
	if err != nil {
		return outMeta, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return outMeta, err
	}

	jData := struct {
		Search []OMDBSearchItem
	}{}

	err = json.Unmarshal(body, &jData)
	if err != nil {
		return outMeta, err
	}

	for _, entry := range jData.Search {
		var cur db_types.MetadataEntry
		imdbId := entry.ImdbID[2:]
		imdbIdInt, err := strconv.ParseInt(imdbId, 10, 64)
		if err != nil {
			logging.ELog(err)
			continue
		}
		cur.ItemId = imdbIdInt
		cur.Title = entry.Title
		cur.Thumbnail = entry.Poster

		cur.Provider = "omdb"
		cur.ProviderID = imdbId

		outMeta = append(outMeta, cur)
	}

	return outMeta, nil
}

func OmdbIdIdentifier(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}

	key := os.Getenv("OMDB_KEY")
	if key == "" {
		return out, errors.New("No api key")
	}

	for len(id) < 7 {
		id = "0" + id
	}
	url := fmt.Sprintf(
		"https://www.omdbapi.com/?apikey=%s&i=%s",
		key,
		url.QueryEscape("tt"+id),
	)

	res, err := http.Get(url)
	if err != nil {
		return out, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return out, err
	}

	jData := new(OMDBResponse)
	err = json.Unmarshal(body, &jData)
	if err != nil {
		return out, err
	}

	return omdbResultToMetadata(*jData)
}
