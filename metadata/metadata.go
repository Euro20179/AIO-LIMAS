package metadata

import "aiolimas/db"

type EntryType string

const (
	E_SHOW      EntryType = "Show"
	E_MOVIE     EntryType = "Movie"
	E_ANIME     EntryType = "Anime"
	E_SONG      EntryType = "Song"
	E_VIDEOGAME EntryType = "VideoGame"
	E_PHOTO     EntryType = "Photo" // will use ai to generate a description
	E_MEME      EntryType = "Meme"  // will use ocr to generate a description
)

// entryType is used as a hint for where to get the metadata from
func GetMetadata(entry *db.InfoEntry, metadataEntry *db.MetadataEntry) {
}
