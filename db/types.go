package db

import "encoding/json"

type MetadataEntry struct {
	ItemId      int64
	Rating      float64
	Description string
	Length      int64
	ReleaseYear int64
}

type InfoEntry struct {
	ItemId        int64
	Title         string
	Format        string
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
	Status string
	ViewCount int64
	StartDate string
	EndDate string
	UserRating float64
}
