package metadata

import (
	"aiolimas/settings"
	db_types "aiolimas/types"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Datafile struct {
	XMLName xml.Name `xml:"datafile"`
	WiiTDB any  `xml:"WiiTDB"`
	Copmanies any `xml:"companies"`
	Genres any `xml:"genres"`
	Descriptors any `xml:"descriptors"`
	Games []GTDBGame `xml:"game"`
}

type GTDBLocale struct {
	XMLName xml.Name `xml:"locale"`
	Lang string `xml:"lang,attr"`
	Title string `xml:"title"`
	Synopsis string `xml:"synopsis"`
}

type GTDBDate struct {
	XMLName xml.Name `xml:"date"`
	Year string `xml:"year,attr"`
	Month string `xml:"month,attr"`
	Day string `xml:"day,attr"`
}

type GTDBGame struct {
	XMLName xml.Name `xml:"game"`
	Name string `xml:"name,attr"`
	Id string `xml:"id"`
	Region string `xml:"region"`
	Languages string `xml:"languages"`
	Locales []GTDBLocale `xml:"locale"`
	Developer string `xml:"developer"`
	Publisher string `xml:"publisher"`
	Date GTDBDate `xml:"date"`
	Genre string `xml:"genre"`
	Rating any `xml:"rating"`
	Input any `xml:"input"`
	Rom any `xml:"rom"`
}

//map of format -> (map of id -> title)
var id2name = map[string]map[string]string{}

//map of format -> (map of id -> game)
var formatDB = map[string]map[string]GTDBGame{}

/*
format can be "wii" | "switch"
*/
func gtdbLoad(format string) error {

	if _, has := id2name[format]; has {
		return nil
	}

	gtdbFolder := os.Getenv("GTDB_ROOT")
	if _, err := os.Stat(gtdbFolder); err != nil {
		return err
	}

	fullPath := fmt.Sprintf("%s/%s", gtdbFolder, format)

	if _, err := os.Stat(fullPath + ".txt"); err != nil {
		return err
	}

	ftxt, err := os.Open(fullPath + ".txt")
	if err != nil {
		return err
	}

	txtData, err := io.ReadAll(ftxt)
	if err != nil {
		return err
	}

	id2names := map[string]string{}

	for _, line := range strings.Split(string(txtData), "\n") {
		if strings.HasPrefix(line, "TITLES") {
			continue
		}

		info := strings.SplitN(line, " = ", 2)
		if len(info) == 1 {
			continue
		}
		id := info[0]
		nativeName := info[1]
		id2names[id] = nativeName
	}

	id2name[format] = id2names

	fxml, err := os.Open(fullPath + ".xml")
	if err != nil {
		return err
	}

	xmlData, err := io.ReadAll(fxml)
	if err != nil {
		return err
	}

	df := Datafile{}

	err = xml.Unmarshal(xmlData, &df)
	if err != nil {
		return err
	}

	formatDB[format] = map[string]GTDBGame{}
	for _, game := range df.Games {
		formatDB[format][game.Id] = game
	}

	return nil
}

func GTDBWiiIdIdentify(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}
	if err := gtdbLoad("wii"); err != nil {
		return out, err
	}

	game, has := formatDB["wii"][id]
	if !has {
		return out, fmt.Errorf("%s not found", id)
	}

	out.Title = game.Name
	lang := game.Languages
	out.Description = game.Locales[0].Synopsis
	out.Title = game.Locales[0].Title
	for _, l := range game.Locales {
		if l.Lang != lang {
			continue
		}
		out.Native_Title = l.Title
	}
	y, _ := strconv.ParseInt(game.Date.Year, 10, 64)
	out.ReleaseYear = y
	out.Provider = "GTDBWii"
	out.ProviderID = game.Id
	out.Thumbnail = fmt.Sprintf("https://art.gametdb.com/wii/cover/US/%s.png", game.Id)

	return out, nil
}

func GTDBWiiIdentify(iinfo IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	out := []db_types.MetadataEntry{}
	if err := gtdbLoad("wii"); err != nil {
		return out, err
	}

	valid := []string{}
	search := strings.ToLower(iinfo.Title)
	for id, v := range formatDB["wii"] {
		if strings.Contains(strings.ToLower(v.Name), search) {
			valid = append(valid, id)
		}
	}

	for count, id := range valid {
		cur := db_types.MetadataEntry{}
		game, has := formatDB["wii"][id]
		if !has {
			return out, fmt.Errorf("%s not found", id)
		}

		cur.Title = game.Name
		lang := game.Languages
		cur.Description = game.Locales[0].Synopsis
		cur.Title = game.Locales[0].Title
		for _, l := range game.Locales {
			if l.Lang != lang {
				continue
			}
			cur.Native_Title = l.Title
		}
		y, _ := strconv.ParseInt(game.Date.Year, 10, 64)
		cur.ReleaseYear = y
		cur.Provider = "gtdbwii"
		cur.ProviderID = game.Id
		cur.ItemId = int64(count)
		cur.Thumbnail = fmt.Sprintf("https://art.gametdb.com/wii/cover/US/%s.png", game.Id)
		out = append(out, cur)
	}

	return out, nil
}
