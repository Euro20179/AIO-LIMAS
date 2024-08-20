package metadata

import (
	"aiolimas/db"
)

// entryType is used as a hint for where to get the metadata from
func GetMetadata(entry *db.InfoEntry, metadataEntry *db.MetadataEntry, override string) {
	switch entry.Type {
	case db.TY_ANIME:
		AnilistShow(entry, metadataEntry)
		break
	case db.TY_MANGA:
		AnilistManga(entry, metadataEntry)
		break
	}
}

func ListMetadataProviders() []string{
	keys := make([]string, 0, len(Providers))
	for k := range Providers {
		keys = append(keys, k)
	}
	return keys
}

type ProviderMap map[string]func(*db.InfoEntry, *db.MetadataEntry) error

var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
}
