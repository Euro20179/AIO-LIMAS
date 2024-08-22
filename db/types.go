package db

import (
	"database/sql"
	"encoding/json"
	"slices"
	"time"
)

type Status string

const (
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
	F_VHS        Format = iota // 1
	F_CD         Format = iota // 2
	F_DVD        Format = iota // 3
	F_BLURAY     Format = iota // 4
	F_4KBLURAY   Format = iota // 5
	F_MANGA      Format = iota // 6
	F_BOOK       Format = iota // 7
	F_DIGITAL    Format = iota // 8
	F_BOARDGAME  Format = iota // 9
	F_STEAM      Format = iota // 10
	F_NIN_SWITCH Format = iota // 11
	F_XBOXONE    Format = iota // 12
	F_XBOX360    Format = iota // 13
	F_OTHER      Format = iota // 14

	F_MOD_DIGITAL Format = 0x1000
)

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
	if format&int64(F_MOD_DIGITAL) == 1 {
		format -= int64(F_MOD_DIGITAL)
	}
	return format >= int64(F_VHS) && format <= int64(F_OTHER)
}

type MediaTypes string

const (
	TY_SHOW      MediaTypes = "Show"
	TY_ANIME     MediaTypes = "Anime"
	TY_MOVIE     MediaTypes = "Movie"
	TY_GAME      MediaTypes = "Game"
	TY_BOARDGAME MediaTypes = "BoardGame"
	TY_SONG      MediaTypes = "Song"
	TY_BOOK      MediaTypes = "Book"
	TY_MANGA     MediaTypes = "Manga"
)

func ListMediaTypes() []MediaTypes {
	return []MediaTypes{
		TY_SHOW, TY_ANIME, TY_MOVIE, TY_GAME,
		TY_BOARDGAME, TY_SONG, TY_BOOK, TY_MANGA,
	}
}

func IsValidType(ty string) bool {
	return slices.Contains(ListMediaTypes(), MediaTypes(ty))
}

type MetadataEntry struct {
	ItemId         int64
	Rating         float64
	Description    string
	ReleaseYear    int64
	Thumbnail      string
	MediaDependant string // see docs/types.md
	Datapoints     string // JSON {string: string} as a string
}

func (self *MetadataEntry) ReadEntry(rows *sql.Rows) error{
	return rows.Scan(
		&self.ItemId,
		&self.Rating,
		&self.Description,
		&self.ReleaseYear,
		&self.Thumbnail,
		&self.MediaDependant,
		&self.Datapoints,
	)
}

type InfoEntry struct {
	ItemId        int64
	En_Title      string //doesn't have to be english, more like, the user's preferred language
	Native_Title  string
	Format        Format
	Location      string
	PurchasePrice float64
	Collection    string
	Parent        int64
	Type          MediaTypes
}

func (self *InfoEntry) ReadEntry(rows *sql.Rows) error {
	return rows.Scan(&self.ItemId, &self.En_Title, &self.Native_Title, &self.Format, &self.Location, &self.PurchasePrice, &self.Collection, &self.Type, &self.Parent)
}

func (self *InfoEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

type UserViewingEntry struct {
	ItemId     int64
	Status     Status
	ViewCount  int64
	StartDate  string
	EndDate    string
	UserRating float64
	Notes      string
}

func (self *UserViewingEntry) ReadEntry(row *sql.Rows) error{
	return row.Scan(
		&self.ItemId,
		&self.Status,
		&self.ViewCount,
		&self.StartDate,
		&self.EndDate,
		&self.UserRating,
		&self.Notes,
	)
}

func (self *UserViewingEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

func (self *UserViewingEntry) unmarshallTimes() ([]uint64, []uint64, error) {
	var startTimes []uint64
	err := json.Unmarshal([]byte(self.StartDate), &startTimes)
	if err != nil {
		return nil, nil, err
	}
	var endTimes []uint64
	err = json.Unmarshal([]byte(self.EndDate), &endTimes)
	if err != nil {
		return nil, nil, err
	}
	return startTimes, endTimes, nil
}

func (self *UserViewingEntry) marshallTimes(startTimes []uint64, endTimes []uint64) error {
	marshalledStart, err := json.Marshal(startTimes)
	if err != nil {
		return err
	}
	marshalledEnd, err := json.Marshal(endTimes)
	if err != nil {
		return err
	}
	self.StartDate = string(marshalledStart)
	self.EndDate = string(marshalledEnd)
	return nil
}

func (self *UserViewingEntry) IsViewing() bool {
	return self.Status == S_VIEWING || self.Status == S_REVIEWING
}

func (self *UserViewingEntry) CanBegin() bool {
	return self.Status == S_PLANNED || self.Status == S_FINISHED || self.Status == S_DROPPED
}

func (self *UserViewingEntry) Begin() error {
	startTimes, endTimes, err := self.unmarshallTimes()
	if err != nil {
		return err
	}
	startTimes = append(startTimes, uint64(time.Now().UnixMilli()))
	// start times and end times array must be same length
	endTimes = append(endTimes, 0)

	if err := self.marshallTimes(startTimes, endTimes); err != nil {
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
	startTimes, endTimes, err := self.unmarshallTimes()
	if err != nil {
		return err
	}

	// this should be 0, overwrite it to the current time
	endTimes[len(endTimes)-1] = uint64(time.Now().UnixMilli())

	if err := self.marshallTimes(startTimes, endTimes); err != nil {
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
	self.Status = S_PLANNED

	return nil
}

func (self *UserViewingEntry) CanDrop() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) Drop() error {
	self.Status = S_DROPPED

	return nil
}

func (self *UserViewingEntry) CanPause() bool {
	return self.IsViewing()
}

func (self *UserViewingEntry) Pause() error {
	self.Status = S_PAUSED

	return nil
}

func (self *UserViewingEntry) CanResume() bool {
	return self.Status == S_PAUSED
}

func (self *UserViewingEntry) Resume() error {
	self.Status = S_REVIEWING
	return nil
}
