package db

import (
	"encoding/json"
	"slices"
	"time"
)

type Status string
const (
	S_VIEWING Status = "Viewing"
	S_FINISHED Status = "Finished"
	S_DROPPED Status = "Dropped"
	S_PLANNED Status = "Planned"
	S_REVIEWING Status = "ReViewing"
)

func IsValidStatus(status string) bool {
	validStatuses := []string{"Viewing", "Finished", "Dropped", "Planned", "ReViewing"}
	return slices.Contains(validStatuses, status)
}

type Format int

const (
	F_VHS       Format = iota // 0
	F_CD        Format = iota // 1
	F_DVD       Format = iota // 2
	F_BLURAY    Format = iota // 3
	F_4KBLURAY  Format = iota // 4
	F_MANGA     Format = iota // 5
	F_BOOK      Format = iota // 6
	F_DIGITAL   Format = iota // 7
	F_VIDEOGAME Format = iota // 8
	F_BOARDGAME Format = iota // 9
)

func IsValidFormat(format int64) bool{
	return format < 10 && format > -1
}

type MetadataEntry struct {
	ItemId      int64
	Rating      float64
	Description string
	Length      int64
	ReleaseYear int64
	Thumbnail string
	Datapoints string //JSON {string: string} as a string
}

type InfoEntry struct {
	ItemId        int64
	Title         string
	Format        Format
	Location      string
	PurchasePrice float64
	Collection    string
	Parent        int64
}

func (self *InfoEntry) ToJson() ([]byte, error) {
	return json.Marshal(self)
}

type UserViewingEntry struct {
	ItemId int64
	Status Status
	ViewCount int64
	StartDate string
	EndDate string
	UserRating float64
}

func (self *UserViewingEntry) unmarshallTimes() ([]uint64, []uint64, error) {
	var startTimes []uint64
	err := json.Unmarshal([]byte(self.StartDate), &startTimes)
	if err != nil{
		return nil, nil, err
	}
	var endTimes []uint64
	err = json.Unmarshal([]byte(self.EndDate), &endTimes)
	if err != nil {
		return nil, nil, err
	}
	return startTimes, endTimes, nil
}

func (self *UserViewingEntry) marshallTimes(startTimes []uint64, endTimes []uint64) error{
	marshalledStart, err := json.Marshal(startTimes)
	if err != nil{
		return err
	}
	marshalledEnd, err := json.Marshal(endTimes)
	if err != nil{
		return err
	}
	self.StartDate = string(marshalledStart)
	self.EndDate = string(marshalledEnd)
	return nil
}

func (self *UserViewingEntry) CanBegin() bool {
	return self.Status != S_VIEWING && self.Status != S_REVIEWING
}

func (self *UserViewingEntry) Begin() error {
	startTimes, endTimes, err := self.unmarshallTimes()
	if err != nil{
		return err
	}
	startTimes = append(startTimes, uint64(time.Now().UnixMilli()))
	// start times and end times array must be same length
	endTimes = append(endTimes, 0)

	if err := self.marshallTimes(startTimes, endTimes); err != nil {
		return err
	}

	return nil
}

func (self *UserViewingEntry) CanEnd() bool {
	return self.Status == S_VIEWING || self.Status == S_REVIEWING
}

func (self *UserViewingEntry) End() error {
	startTimes, endTimes, err := self.unmarshallTimes()
	if err != nil{
		return err
	}

	//this should be 0, overwrite it to the current time
	endTimes[len(endTimes) - 1] = uint64(time.Now().UnixMilli())

	if err := self.marshallTimes(startTimes, endTimes); err != nil {
		return err
	}

	return nil
}
