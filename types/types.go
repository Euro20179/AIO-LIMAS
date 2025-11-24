package db_types

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"time"
)

type Relations struct {
	Children []int64
	Requires []int64
	Copies   []int64
}

type ArtStyle uint

const (
	AS_ANIME       ArtStyle = 1
	AS_CARTOON     ArtStyle = 2
	AS_HANDRAWN    ArtStyle = 4
	AS_DIGITAL     ArtStyle = 8
	AS_CGI         ArtStyle = 16
	AS_LIVE_ACTION ArtStyle = 32
	AS_2D          ArtStyle = 64
	AS_3D          ArtStyle = 128
)

func ArtStyle2Str(style ArtStyle) string {
	bit2Name := ListArtStyles()
	var styles []string
	for i := AS_ANIME; i <= AS_LIVE_ACTION; i *= 2 {
		if (style & ArtStyle(i)) == ArtStyle(i) {
			styles = append(styles, bit2Name[ArtStyle(i)])
		}
	}

	return strings.Join(styles, " + ")
}

func ListArtStyles() map[ArtStyle]string {
	return map[ArtStyle]string{
		AS_ANIME:       "Anime",
		AS_CARTOON:     "Cartoon",
		AS_HANDRAWN:    "Handdrawn",
		AS_DIGITAL:     "Digital",
		AS_CGI:         "CGI",
		AS_LIVE_ACTION: "Liveaction",
		AS_2D:          "2D",
		AS_3D:          "3D",
	}
}

type Status string

const (
	S_NONE      Status = ""
	S_VIEWING   Status = "Viewing"   // first viewing experience
	S_FINISHED  Status = "Finished"  // when the user has finished viewing/reviewing
	S_DROPPED   Status = "Dropped"   // if the user stops viewing, and does not plan to continue
	S_PAUSED    Status = "Paused"    // if the user stopes viewing, but does plan to continue
	S_PLANNED   Status = "Planned"   // plans to view or review at some point
	S_REVIEWING Status = "ReViewing" // the user has already finished or dropped, but is viewing again
	S_WAITING   Status = "Waiting"   // if the user is waiting for a new season or something
	// or if the user has unpaused
)

func IsValidStatus(status string) bool {
	return slices.Contains(ListStatuses(), Status(status))
}

func ListStatuses() []Status {
	return []Status{
		"",
		"Viewing",
		"Finished",
		"Dropped",
		"Planned",
		"ReViewing",
		"Paused",
	}
}

type Format uint32

// the digital modifier can be applied to any format

// This way the user has 2 options, they can say either that the item is F_DIGITAL
// or they can be more specific and say it's F_VHS that's been digitized with F_MOD_DIGITAL
// another use case is for console versions, eg: F_NIN_SWITCH refers to the cartridge,
// but F_NIN_SWITCH & F_MOD_DIGITAL would be the store version

// F_DIGITAL & F_MOD_DIGITAL has undefined meaning
const (
	F_VHS        Format = iota // 0
	F_CD         Format = iota // 1
	F_DVD        Format = iota // 2
	F_BLURAY     Format = iota // 3
	F_4KBLURAY   Format = iota // 4
	F_MANGA      Format = iota // 5
	F_BOOK       Format = iota // 6
	F_DIGITAL    Format = iota // 7
	F_BOARDGAME  Format = iota // 8
	F_STEAM      Format = iota // 9
	F_NIN_SWITCH Format = iota // 10
	F_XBOXONE    Format = iota // 11
	F_XBOX360    Format = iota // 12
	F_OTHER      Format = iota // 13
	F_VINYL      Format = iota // 14
	F_IMAGE      Format = iota // 15
	F_UNOWNED    Format = iota // 16

	F_MOD_DIGITAL Format = 0x1000
)

func ListFormats() map[Format]string {
	return map[Format]string{
		F_VHS:         "VHS",
		F_CD:          "CD",
		F_DVD:         "DVD",
		F_BLURAY:      "BLURAY",
		F_4KBLURAY:    "4KBLURAY",
		F_MANGA:       "MANGA",
		F_BOOK:        "BOOK",
		F_DIGITAL:     "DIGITAL",
		F_BOARDGAME:   "BOARDGAME",
		F_STEAM:       "STEAM",
		F_NIN_SWITCH:  "NIN_SWITCH",
		F_XBOXONE:     "XBOXONE",
		F_XBOX360:     "XBOX360",
		F_OTHER:       "OTHER",
		F_VINYL:       "VINYL",
		F_IMAGE:       "IMAGE",
		F_UNOWNED:     "UNOWNED",
		F_MOD_DIGITAL: "MOD_DIGITAL",
	}
}

func (self *Format) MkDigital() Format {
	return *self | F_MOD_DIGITAL
}

func (self *Format) MkUnDigital() Format {
	if self.IsDigital() {
		return *self - F_MOD_DIGITAL
	}
	return *self
}

func (self *Format) IsDigital() bool {
	return (*self & F_MOD_DIGITAL) == 1
}

func IsValidFormat(format int64) bool {
	if format&int64(F_MOD_DIGITAL) == int64(F_MOD_DIGITAL) {
		format -= int64(F_MOD_DIGITAL)
	}
	return format >= int64(F_VHS) && format <= int64(F_UNOWNED)
}

type MediaTypes string

const (
	TY_SHOW        MediaTypes = "Show"
	TY_EPISODE     MediaTypes = "Episode"
	TY_DOCUMENTARY MediaTypes = "Documentary"
	TY_MOVIE       MediaTypes = "Movie"
	TY_MOVIE_SHORT MediaTypes = "MovieShort"
	TY_GAME        MediaTypes = "Game"
	TY_BOARDGAME   MediaTypes = "BoardGame"
	TY_SONG        MediaTypes = "Song"
	TY_BOOK        MediaTypes = "Book"
	TY_MANGA       MediaTypes = "Manga"
	TY_COLLECTION  MediaTypes = "Collection"
	TY_PICTURE     MediaTypes = "Picture"
	TY_MEME        MediaTypes = "Meme"
	TY_LIBRARY     MediaTypes = "Library"
	TY_VIDEO       MediaTypes = "Video"
	TY_SHORTSTORY  MediaTypes = "ShortStory"
	TY_ALBUMN      MediaTypes = "Albumn"
	TY_SOUNDTRACK  MediaTypes = "Soundtrack"
)

func ListMediaTypes() []MediaTypes {
	return []MediaTypes{
		TY_SHOW, TY_MOVIE, TY_GAME,
		TY_BOARDGAME, TY_SONG, TY_BOOK, TY_MANGA,
		TY_COLLECTION, TY_MOVIE_SHORT,
		TY_PICTURE, TY_MEME, TY_LIBRARY,
		TY_DOCUMENTARY, TY_EPISODE, TY_VIDEO,
		TY_SHORTSTORY,
		TY_ALBUMN, TY_SOUNDTRACK,
	}
}

func IsValidType(ty string) bool {
	return slices.Contains(ListMediaTypes(), MediaTypes(ty))
}

func StructNamesToDict(entity any) map[string]any {
	items := make(map[string]any)

	val := reflect.ValueOf(entity)

	for i := range val.NumField() {
		field := val.Type().Field(i)

		if field.Tag.Get("runtime") == "true" {
			continue
		}

		name := field.Name

		value := val.FieldByName(name).Interface()

		words := strings.Split(name, "_")
		for i, word := range words {
			lowerLetter := strings.ToLower(string(word[0]))
			words[i] = lowerLetter + word[1:]
		}
		name = strings.Join(words, "_")

		items[name] = value
	}

	return items
}

type TableRepresentation interface {
	Id() int64
	ReadEntryCopy(*sql.Rows) (TableRepresentation, error)
	ToJson() ([]byte, error)
}

// names here MUST match names in the metadta sqlite table
type MetadataEntry struct {
	Uid    int64
	ItemId int64
	Rating float64
	// different sources will do ratings differently,
	// let them set the max rating
	RatingMax      float64
	Description    string
	ReleaseYear    int64
	Thumbnail      string
	MediaDependant string // see docs/types.md
	Datapoints     string // JSON {string: string} as a string
	Title          string // this is different from infoentry in that it's automatically generated
	Native_Title   string // same with this
	Provider       string // the provider that generated the metadata
	ProviderID     string // the id that the provider used
	Genres         string
}

func (self MetadataEntry) Id() int64 {
	return self.ItemId
}

func (self MetadataEntry) ReadEntryCopy(rows *sql.Rows) (TableRepresentation, error) {
	return self, self.ReadEntry(rows)
}

func (self *MetadataEntry) ReadEntry(rows *sql.Rows) error {
	return rows.Scan(
		&self.Uid,
		&self.ItemId,
		&self.Rating,
		&self.Description,
		&self.ReleaseYear,
		&self.Thumbnail,
		&self.MediaDependant,
		&self.Datapoints,
		&self.Title,
		&self.Native_Title,
		&self.RatingMax,
		&self.Provider,
		&self.ProviderID,
		&self.Genres,
	)
}

func (self *MetadataEntry) NormalizedRating() float64 {
	return self.Rating / self.RatingMax * 100
}

func (self MetadataEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

type InfoEntry struct {
	Uid           int64
	ItemId        int64
	En_Title      string // doesn't have to be english, more like, the user's preferred language
	Native_Title  string
	Format        Format
	Location      string
	PurchasePrice float64
	Collection    string
	ParentId      int64
	Type          MediaTypes
	ArtStyle      ArtStyle
	CopyOf        int64
	Library       int64
	Requires      int64
	RecommendedBy string

	// RUNTIME VALUES (not stored in database), see self.ReadEntry
	Tags []string `runtime:"true"`
}

func (self *InfoEntry) IsAnime() bool {
	return self.ArtStyle&AS_ANIME == AS_ANIME
}

func (self InfoEntry) Id() int64 {
	return self.ItemId
}

func (self InfoEntry) ReadEntryCopy(rows *sql.Rows) (TableRepresentation, error) {
	return self, self.ReadEntry(rows)
}

func (self InfoEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

func (self *InfoEntry) ReadEntry(rows *sql.Rows) error {
	err := rows.Scan(
		&self.Uid,
		&self.ItemId,
		&self.En_Title,
		&self.Native_Title,
		&self.Format,
		&self.Location,
		&self.PurchasePrice,
		&self.Collection,
		&self.Type,
		&self.ParentId,
		&self.CopyOf,
		&self.ArtStyle,
		&self.Library,
		&self.Requires,
		&self.RecommendedBy,
	)
	if err != nil {
		return err
	}

	for _, name := range strings.Split(self.Collection, "\x1F") {
		if name == "" {
			continue
		}
		self.Tags = append(self.Tags, name)
	}

	return nil
}

type UserViewingEvent struct {
	Uid       int64
	ItemId    int64
	Event     string
	TimeZone  string
	Timestamp uint64
	Before    uint64
	After     uint64 // this is also a timestamp, for when the exact timestamp is unknown
	// this is to ensure that order can be determined
	EventId int64
}

func (self UserViewingEvent) Id() int64 {
	return self.ItemId
}

func (self UserViewingEvent) ReadEntryCopy(rows *sql.Rows) (TableRepresentation, error) {
	return self, self.ReadEntry(rows)
}

func (self *UserViewingEvent) ReadEntry(rows *sql.Rows) error {
	return rows.Scan(
		&self.Uid,
		&self.ItemId,
		&self.Timestamp,
		&self.After,
		&self.Event,
		&self.TimeZone,
		&self.Before,
		&self.EventId,
	)
}

func (self UserViewingEvent) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

func (self *UserViewingEvent) ToHumanTime() string {
	stamp := self.Timestamp / 1000
	after := false

	if self.After > 0 && self.Timestamp == 0 {
		stamp = self.After / 1000
		after = true
	}

	if stamp == 0 {
		return "unkonown"
	}

	t := time.Unix(int64(stamp), 0)

	if !after {
		return t.Format("01/02/2006 - 15:04:05")
	} else {
		return t.Format("after 01/02/2006 - 15:04:05")
	}
}

type UserViewingEntry struct {
	Uid             int64
	ItemId          int64
	Status          Status
	ViewCount       int64
	UserRating      float64
	Notes           string
	CurrentPosition string
	Extra           string
	Minutes         int64
}

func (self UserViewingEntry) Id() int64 {
	return self.ItemId
}

func (self UserViewingEntry) ReadEntryCopy(rows *sql.Rows) (TableRepresentation, error) {
	return self, self.ReadEntry(rows)
}

func (self *UserViewingEntry) ReadEntry(row *sql.Rows) error {
	return row.Scan(
		&self.Uid,
		&self.ItemId,
		&self.Status,
		&self.ViewCount,
		&self.UserRating,
		&self.Notes,
		&self.CurrentPosition,
		&self.Extra,
		&self.Minutes,
	)
}

func (self UserViewingEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

func (self *UserViewingEntry) IsViewing() bool {
	return self.Status == S_VIEWING || self.Status == S_REVIEWING
}

func (self *UserViewingEntry) CanBegin() bool {
	return self.Status == S_FINISHED || self.Status == S_PLANNED || self.Status == S_DROPPED || self.Status == ""
}

func (self *UserViewingEntry) CanFinish() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) CanPlan() bool {
	return self.Status == S_DROPPED || self.Status == ""
}

func (self *UserViewingEntry) CanDrop() bool {
	return self.IsViewing() || self.Status == S_WAITING
}

func (self *UserViewingEntry) CanPause() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) CanResume() bool {
	return self.Status == S_PAUSED || self.Status == S_WAITING
}

func (self *UserViewingEntry) CanWait() bool {
	return self.Status == S_VIEWING || self.Status == S_REVIEWING
}

type EntryTree struct {
	EntryInfo InfoEntry
	MetaInfo  MetadataEntry
	UserInfo  UserViewingEntry
	Children  []string
	Copies    []string
}

func (self EntryTree) ToJson() ([]byte, error) {
	return json.Marshal(self)
}
