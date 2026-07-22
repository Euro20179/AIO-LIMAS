package metadata

import (
	"aiolimas/logging"
	"aiolimas/settings"
	db_types "aiolimas/types"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

var gtdbMEM, err = sql.Open("sqlite3", "file:gtdb?mode=memory&cache=shared")

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

type GTDBESRBRating struct {
	XMLName xml.Name `xml:"rating_ESRB"`
	Rating string `xml:"value,attr"`
	Descriptor []string `xml:"descriptor_ESRB"`
}

type GTDBPEGIRating struct {
	XMLName xml.Name `xml:"rating_PEGI"`
	Rating string `xml:"value,attr"`
	Descriptor []string `xml:"descriptor_PEGI"`
}

type GTDBRating struct {
	XMLName xml.Name `xml:"rating"`
	Agency string `xml:"type,attr"`
	Rating string `xml:"value,attr"`
	Descriptors string
}

type GTDBGame struct {
	XMLName xml.Name `xml:"game"`
	Name string `xml:"name,attr"`
	Id string `xml:"id"`
	Type string `xml:"type"`
	Region string `xml:"region"`
	Languages string `xml:"languages"`
	Locales []GTDBLocale `xml:"locale"`
	Developer string `xml:"developer"`
	Publisher string `xml:"publisher"`
	Date GTDBDate `xml:"date"`
	Genre string `xml:"genre"`
	RatingESRB GTDBESRBRating `xml:"rating_ESRB"`
	RatingPEGI GTDBPEGIRating `xml:"rating_PEGI"`
	Rating GTDBRating `xml:"rating"`
	Input any `xml:"input"`
	Rom any `xml:"rom"`
	NativeName string
}

//map of format -> (map of id -> title)
var id2name = map[string]map[string]string{}

var formatsLoaded = []string{}

//map of format -> (map of id -> game)
var formatDB = map[string]map[string]GTDBGame{}

/*
format can be "wii" | "switch"
*/
func gtdbLoad(format string) error {
	if slices.Contains(formatsLoaded, format){
		return nil
	}

	_, err = gtdbMEM.Exec(`CREATE TABLE IF NOT EXISTS games (
		native_name TEXT,
		id TEXT,
		name TEXT,
		region TEXT DEFAULT '',
		languages TEXT DEFAULT '',
		developer TEXT DEFAULT '',
		publisher TEXT DEFAULT '',
		year TEXT DEFAULT '',
		month TEXT DEFAULT '',
		day TEXT DEFAULT '',
		genres TEXT DEFAULT '',
		format TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS locales (
		gameID TEXT DEFAULT '',
		lang TEXT DEFAULT '',
		title TEXT DEFAULT '',
		synopsis TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS ratings (
		gameID TEXT DEFAULT '',
		agency TEXT DEFAULT '',
		rating TEXT DEFAULT '',
		descriptors TEXT DEFAULT ''
	);
	`)

	if err != nil {
		logging.ELog(err)
		return err
	}

	gtdbFolder := os.Getenv("GTDB_ROOT")
	if _, err := os.Stat(gtdbFolder); err != nil {
		logging.ELog(err)
		return err
	}

	fullPath := fmt.Sprintf("%s/%s", gtdbFolder, format)

	if _, err := os.Stat(fullPath + ".txt"); err != nil {
		logging.ELog(err)
		return err
	}

	ftxt, err := os.Open(fullPath + ".txt")
	if err != nil {
		logging.ELog(err)
		return err
	}

	txtData, err := io.ReadAll(ftxt)
	if err != nil {
		logging.ELog(err)
		return err
	}

	tx, err := gtdbMEM.Begin()
	if err != nil {
		logging.ELog(err)
		return err
	}

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
		if _, err = tx.Exec(`INSERT INTO games (native_name, id) VALUES (?, ?);`, nativeName, id); err != nil {
			tx.Rollback()
			logging.ELog(err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		logging.ELog(err)
		return err
	}

	fxml, err := os.Open(fullPath + ".xml")
	if err != nil {
		logging.ELog(err)
		return err
	}

	xmlData, err := io.ReadAll(fxml)
	if err != nil {
		logging.ELog(err)
		return err
	}

	df := Datafile{}

	err = xml.Unmarshal(xmlData, &df)
	if err != nil {
		logging.ELog(err)
		return err
	}

	tx, err = gtdbMEM.Begin()
	if err != nil {
		logging.ELog(err)
		return err
	}

	formatDB[format] = map[string]GTDBGame{}
	for _, game := range df.Games {
		if game.Type == "" {
			game.Type = "Wii"
		}

		langs := strings.Split(game.Languages, ",")
		genres := strings.Split(game.Genre, ",")

		genresJSON, _ := json.Marshal(genres)
		langsJSON, _ := json.Marshal(langs)

		tx.Exec(
			`UPDATE games SET
				name = ?,
				region = ?,
				languages = ?,
				developer = ?,
				publisher = ?,
				year = ?,
				month = ?,
				day = ?,
				genres = ?,
				format = ?
			WHERE id = ?`,
			game.Name,
			game.Region,
			langsJSON,
			game.Developer,
			game.Publisher,
			game.Date.Year,
			game.Date.Month,
			game.Date.Day,
			genresJSON,
			format,
			game.Id,
		)

		for _, locale := range game.Locales {
			tx.Exec(`INSERT INTO locales (gameID, lang, title, synopsis) VALUES (?, ?, ?, ?);`,
				game.Id,
				locale.Lang,
				locale.Title,
				locale.Synopsis,
			)
		}

		if game.Rating.Rating != "" {
			tx.Exec(`INSERT INTO ratings (gameId, agency, rating) VALUES (?, ?, ?)`,
					game.Id,
					game.Rating.Agency,
					game.Rating.Rating)
		}

		if game.RatingESRB.Rating != "" {
			descriptors := strings.Join(game.RatingESRB.Descriptor, ", ")
			tx.Exec(`INSERT INTO ratings (gameId, agency, rating, descriptors) VALUES (?, ?, ?, ?)`,
					game.Id,
					"ESRB",
					game.RatingESRB.Rating,
					descriptors)
		}

		if game.RatingPEGI.Rating != "" {
			descriptors := strings.Join(game.RatingPEGI.Descriptor, ", ")
			tx.Exec(`INSERT INTO ratings (gameId, agency, rating, descriptors) VALUES (?, ?, ?, ?)`,
					game.Id,
					"PEGI",
					game.RatingESRB.Rating,
					descriptors)
		}
	}

	err = tx.Commit()
	if err != nil {
		logging.ELog(err)
		return err
	}

	_, err = gtdbMEM.Exec(fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS %sSearch USING fts4(rowid, name, native_name);

		INSERT INTO %sSearch(rowid, name, native_name)
		SELECT id, name, native_name FROM games;
	`, format, format))

	if err != nil {
		logging.ELog(err)
		return err
	}

	formatsLoaded = append(formatsLoaded, format)

	return nil
}

func gtdbApply(game GTDBGame, out *db_types.MetadataEntry, format string) {
	out.Title = game.Name
	langs := []string{}
	err = json.Unmarshal([]byte(game.Languages), &langs)
	if err == nil {
		out.Title = strings.ReplaceAll(out.Title, " (" + strings.Join(langs, ",") + ")", "")
	}

	if len(game.Locales) > 0 {
		out.Description = game.Locales[0].Synopsis
	}

	if game.NativeName != "" && game.Name != game.NativeName {
		out.Native_Title = game.NativeName
	}

	md := map[string]string{}

	if game.Developer != "" {
		md["Game-developers"] = game.Developer
	}

	if game.Publisher != "" {
		md["Game-publishers"] = game.Publisher
	}
	if game.Languages != "" {
		md["Game-languages"] = strings.Join(langs, ", ")
	}

	if game.Type == "" {
		md["Game-console"] = "Wii"
	} else {
		md["Game-console"] = game.Type
	}

	if game.Rating.Rating != "" {
		md["Game-rating"] = game.Rating.Rating
	}

	if game.Rating.Agency != "" {
		md["Game-rating-agency"] = game.Rating.Agency
	}

	if game.Rating.Descriptors != "" {
		md["Game-rating-descriptors"] = game.Rating.Descriptors
	}

	mdMarshalled, err := json.Marshal(md)
	if err == nil {
		out.MediaDependant = string(mdMarshalled)
	}

	y, _ := strconv.ParseInt(game.Date.Year, 10, 64)
	out.Genres = game.Genre
	out.ReleaseYear = y
	out.ProviderID = game.Id
	ext := "jpg"
	if format == "wii" {
		ext = "png"
	}
	out.Thumbnail = fmt.Sprintf("https://art.gametdb.com/%s/cover/US/%s.%s", format, game.Id, ext)
}

func gtdbIdIdentify(format string, id string) (db_types.MetadataEntry, error) {
	out := db_types.MetadataEntry{}
	if err := gtdbLoad(format); err != nil {
		return out, err
	}

	rows, err := gtdbMEM.Query(`SELECT games.*, lang, synopsis, title, rating, agency FROM games JOIN locales ON games.id = locales.gameID LEFT JOIN ratings ON games.id = ratings.gameID WHERE id = ?`, id)
	if err != nil {
		return out, err
	}

	defer rows.Close()

	game := GTDBGame{}
	locale := GTDBLocale{}
	if !rows.Next() {
		return out, fmt.Errorf("id %s not found", id)
	}
	if err = rows.Scan(
		&game.NativeName,
		&game.Id,
		&game.Name,
		&game.Region,
		&game.Languages,
		&game.Developer,
		&game.Publisher,
		&game.Date.Year,
		&game.Date.Month,
		&game.Date.Day,
		&game.Genre,
		&game.Type,
		&locale.Lang,
		&locale.Synopsis,
		&locale.Title,
		&game.Rating.Rating,
		&game.Rating.Agency,
	); err != nil {
		return out, err
	}

	game.Locales = append(game.Locales, locale)

	gtdbApply(game, &out, format)
	out.Provider = fmt.Sprintf("gtdb%s", format)
	return out, nil
}

func gtdbIdentify(format string, iinfo IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	out := []db_types.MetadataEntry{}
	if err := gtdbLoad(format); err != nil {
		return out, err
	}

	rows, err := gtdbMEM.Query(fmt.Sprintf(
		`SELECT g.*, l.lang, l.synopsis, l.title FROM games as g
		JOIN %sSearch
			ON g.id = %sSearch.rowid
		LEFT JOIN locales as l
			ON g.id = l.gameID
		WHERE %sSearch MATCH ?
			AND g.format LIKE ?`,
		format,
		format,
		format,
	), iinfo.Title, format)

	if err != nil {
		logging.ELog(err)
		return out, err
	}

	defer rows.Close()

	count := 1
	for rows.Next() {
		game := GTDBGame{}
		locale := GTDBLocale{}
		if err = rows.Scan(
			&game.NativeName,
			&game.Id,
			&game.Name,
			&game.Region,
			&game.Languages,
			&game.Developer,
			&game.Publisher,
			&game.Date.Year,
			&game.Date.Month,
			&game.Date.Day,
			&game.Genre,
			&game.Type,
			&locale.Lang,
			&locale.Synopsis,
			&locale.Title,
		); err != nil {
			return out, err
		}

		game.Locales = append(game.Locales, locale)
		cur := db_types.MetadataEntry{}
		gtdbApply(game, &cur, format)
		cur.Provider = fmt.Sprintf("gtdb%s", format)
		cur.ItemId = int64(count)
		out = append(out, cur)
		count++
	}

	return out, nil
}

func GTDBWiiIdIdentify(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	return gtdbIdIdentify("wii", id)
}

func GTDBWiiIdentify(iinfo IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	return gtdbIdentify("wii", iinfo)
}


func GTDBSwitchIdIdentify(id string, us settings.SettingsData) (db_types.MetadataEntry, error) {
	return gtdbIdIdentify("switch", id)
}

func GTDBSwitchIdentify(iinfo IdentifyMetadata) ([]db_types.MetadataEntry, error) {
	return gtdbIdentify("switch", iinfo)
}
