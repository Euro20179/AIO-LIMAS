package db

import "encoding/json"

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
	Status string
	ViewCount int64
	StartDate string
	EndDate string
	UserRating float64
}
