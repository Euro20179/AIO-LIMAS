package metadata

import (
	"aiolimas/db"
)

// entryType is used as a hint for where to get the metadata from
func GetMetadata(entry *db.InfoEntry, metadataEntry *db.MetadataEntry, override string) (db.MetadataEntry, error) {
	if entry.IsAnime {
		return AnilistShow(entry, metadataEntry)
	}
	switch entry.Type {
	case db.TY_MANGA:
		return AnilistManga(entry, metadataEntry)
	case db.TY_SHOW:
		return OMDBProvider(entry, metadataEntry)
	case db.TY_MOVIE:
		return OMDBProvider(entry, metadataEntry)
	}
	var out db.MetadataEntry
	return out, nil
}

func ListMetadataProviders() []string {
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

type ProviderMap map[string]func(*db.InfoEntry, *db.MetadataEntry) (db.MetadataEntry, error)

var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
	"omdb":          OMDBProvider,
}
