package metadata;


type EntryType string
const (
	E_SHOW EntryType = "Show"
	E_MOVIE EntryType = "Movie"
	E_ANIME EntryType = "Anime"
	E_SONG EntryType = "Song"
	E_VIDEOGAME EntryType = "VideoGame"
)

//entryType is used as a hint for where to get the metadata from
func GetMetadata(entryType EntryType) {
}
