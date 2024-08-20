package metadata

import (
	"aiolimas/db"
)

// entryType is used as a hint for where to get the metadata from
func GetMetadata(entry *db.InfoEntry, metadataEntry *db.MetadataEntry, override string) error{
	switch entry.Type {
	case db.TY_ANIME:
		return AnilistShow(entry, metadataEntry)
	case db.TY_MANGA:
		return AnilistManga(entry, metadataEntry)
	}
	return nil
}

func ListMetadataProviders() []string{
	keys := make([]string, 0, len(Providers))
	for k := range Providers {
		keys = append(keys, k)
	}
	return keys
}

func IsValidProvider(name string) bool {
	 _, contains := Providers[name]
	return contains
}

type ProviderMap map[string]func(*db.InfoEntry, *db.MetadataEntry) error

var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
}
