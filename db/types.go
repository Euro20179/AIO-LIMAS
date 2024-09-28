package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
)

type EntryRepresentor interface {
	ToJson() ([]byte, error)
	ReadEntry(rows *sql.Rows) error
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
	// or if the user has unpaused
)

func IsValidStatus(status string) bool {
	validStatuses := ListStatuses()
	return slices.Contains(validStatuses, Status(status))
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
	return format >= int64(F_VHS) && format <= int64(F_VINYL)
}

type MediaTypes string

const (
	TY_SHOW        MediaTypes = "Show"
	TY_MOVIE       MediaTypes = "Movie"
	TY_MOVIE_SHORT MediaTypes = "MovieShort"
	TY_GAME        MediaTypes = "Game"
	TY_BOARDGAME   MediaTypes = "BoardGame"
	TY_SONG        MediaTypes = "Song"
	TY_BOOK        MediaTypes = "Book"
	TY_MANGA       MediaTypes = "Manga"
	TY_COLLECTION  MediaTypes = "Collection"
)

func ListMediaTypes() []MediaTypes {
	return []MediaTypes{
		TY_SHOW, TY_MOVIE, TY_GAME,
		TY_BOARDGAME, TY_SONG, TY_BOOK, TY_MANGA,
		TY_COLLECTION, TY_MOVIE_SHORT,
	}
}

func IsValidType(ty string) bool {
	return slices.Contains(ListMediaTypes(), MediaTypes(ty))
}

type MetadataEntry struct {
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
}

func (self *MetadataEntry) ReadEntry(rows *sql.Rows) error {
	return rows.Scan(
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
	)
}

func (self *MetadataEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

type InfoEntry struct {
	ItemId        int64
	En_Title      string // doesn't have to be english, more like, the user's preferred language
	Native_Title  string
	Format        Format
	Location      string
	PurchasePrice float64
	Collection    string
	Parent        int64
	Type          MediaTypes
	IsAnime       bool
	CopyOf        int64
}

func (self *InfoEntry) ReadEntry(rows *sql.Rows) error {
	return rows.Scan(
		&self.ItemId,
		&self.En_Title,
		&self.Native_Title,
		&self.Format,
		&self.Location,
		&self.PurchasePrice,
		&self.Collection,
		&self.Type,
		&self.Parent,
		&self.IsAnime,
		&self.CopyOf,
	)
}

func (self *InfoEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

type UserViewingEvent struct {
	ItemId    int64
	Event     string
	Timestamp uint64
	After     uint64 // this is also a timestamp, for when the exact timestamp is unknown
	// this is to ensure that order can be determined
}

func (self *UserViewingEvent) ReadEntry(rows *sql.Rows) error {
	return rows.Scan(
		&self.ItemId,
		&self.Timestamp,
		&self.After,
		&self.Event,
	)
}

func (self *UserViewingEvent) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

type UserViewingEntry struct {
	ItemId          int64
	Status          Status
	ViewCount       int64
	UserRating      float64
	Notes           string
	CurrentPosition string
}

func (self *UserViewingEntry) ReadEntry(row *sql.Rows) error {
	return row.Scan(
		&self.ItemId,
		&self.Status,
		&self.ViewCount,
		&self.UserRating,
		&self.Notes,
		&self.CurrentPosition,
	)
}

func (self *UserViewingEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

func (self *UserViewingEntry) IsViewing() bool {
	return self.Status == S_VIEWING || self.Status == S_REVIEWING
}

func (self *UserViewingEntry) CanBegin() bool {
	return self.Status == S_PLANNED || self.Status == S_FINISHED || self.Status == S_DROPPED
}

func (self *UserViewingEntry) Begin() error {
	err := RegisterBasicUserEvent("Viewing", self.ItemId)
	if err != nil {
		return err
	}

	if self.Status != S_FINISHED {
		self.Status = S_VIEWING
	} else {
		self.Status = S_REVIEWING
	}

	return nil
}

func (self *UserViewingEntry) CanFinish() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) Finish() error {
	err := RegisterBasicUserEvent("Finished", self.ItemId)
	if err != nil {
		return err
	}

	self.Status = S_FINISHED
	self.ViewCount += 1

	return nil
}

func (self *UserViewingEntry) CanPlan() bool {
	return self.Status == S_DROPPED || self.Status == ""
}

func (self *UserViewingEntry) Plan() error {
	err := RegisterBasicUserEvent("Planned", self.ItemId)
	if err != nil {
		return err
	}

	self.Status = S_PLANNED

	return nil
}

func (self *UserViewingEntry) CanDrop() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) Drop() error {
	err := RegisterBasicUserEvent("Dropped", self.ItemId)
	if err != nil {
		return err
	}

	self.Status = S_DROPPED

	return nil
}

func (self *UserViewingEntry) CanPause() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) Pause() error {
	err := RegisterBasicUserEvent("Paused", self.ItemId)
	if err != nil {
		return err
	}

	self.Status = S_PAUSED

	return nil
}

func (self *UserViewingEntry) CanResume() bool {
	return self.Status == S_PAUSED
}

func (self *UserViewingEntry) Resume() error {
	err := RegisterBasicUserEvent("ReViewing", self.ItemId)
	if err != nil {
		return err
	}

	self.Status = S_REVIEWING
	return nil
}

type EntryTree struct {
	EntryInfo InfoEntry
	MetaInfo  MetadataEntry
	UserInfo  UserViewingEntry
	Children  []string
	Copies    []string
}

func (self *EntryTree) ToJson() ([]byte, error) {
	return json.Marshal(*self)
}

func BuildEntryTree() (map[int64]EntryTree, error) {
	out := map[int64]EntryTree{}

	allRows, err := Db.Query(`SELECT * FROM entryInfo`)
	if err != nil {
		return out, err
	}

	for allRows.Next() {
		var cur EntryTree

		err := cur.EntryInfo.ReadEntry(allRows)
		if err != nil {
			println(err.Error())
			continue
		}
		cur.UserInfo, err = GetUserViewEntryById(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		cur.MetaInfo, err = GetMetadataEntryById(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		children, err := GetChildren(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		for _, child := range children {
			cur.Children = append(cur.Children, fmt.Sprintf("%d", child.ItemId))
		}

		copies, err := GetCopiesOf(cur.EntryInfo.ItemId)
		if err != nil {
			println(err.Error())
			continue
		}

		for _, c := range copies {
			cur.Copies = append(cur.Copies, fmt.Sprintf("%d", c.ItemId))
		}

		out[cur.EntryInfo.ItemId] = cur
	}
	//
	// for id, cur := range out {
	// 	children, err := GetChildren(id)
	// 	if err != nil{
	// 		println(err.Error())
	// 		continue
	// 	}
	// 	for _, child := range children {
	// 		cur.Children = append(cur.Children, child.ItemId)
	// 	}
	// }

	return out, nil
}
