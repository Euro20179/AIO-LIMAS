package metadata

import (
	"fmt"

	"aiolimas/db"
)

type IdentifyMetadata struct {
	Title string
}

// entryType is used as a hint for where to get the metadata from
func GetMetadata(entry *db.InfoEntry, metadataEntry *db.MetadataEntry, override string) (db.MetadataEntry, error) {
	if entry.IsAnime {
		return AnilistShow(entry, metadataEntry)
	}
	switch entry.Type {
	case db.TY_MANGA:
		return AnilistManga(entry, metadataEntry)

	case db.TY_SHOW:
		fallthrough
	case db.TY_MOVIE:
		return OMDBProvider(entry, metadataEntry)

	case db.TY_PICTURE:
		fallthrough
	case db.TY_MEME:
		return ImageProvider(entry, metadataEntry)
	}
	var out db.MetadataEntry
	return out, nil
}

func Identify(identifySearch IdentifyMetadata, identifier string) ([]db.MetadataEntry, string, error) {
	fn, contains := IdentifyProviders[identifier]
	if !contains {
		return []db.MetadataEntry{}, "", fmt.Errorf("Invalid provider %s", identifier)
	}

	res, err := fn(identifySearch)
	return res, identifier, err
}

func GetMetadataById(id string, provider string) (db.MetadataEntry, error) {
	fn, contains := IdIdentifiers[provider]
	if !contains {
		return db.MetadataEntry{}, fmt.Errorf("Invalid provider: %s", provider)
	}
	return fn(id)
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

func IsValidIdentifier(name string) bool {
	_, contains := IdentifyProviders[name]
	return contains
}

func IsValidIdIdentifier(name string) bool {
	_, contains := IdIdentifiers[name]
	return contains
}

type ProviderMap map[string]func(*db.InfoEntry, *db.MetadataEntry) (db.MetadataEntry, error)

var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
	"omdb":          OMDBProvider,
	"image":         ImageProvider,
}

type IdentifiersMap = map[string]func(info IdentifyMetadata) ([]db.MetadataEntry, error)

var IdentifyProviders IdentifiersMap = IdentifiersMap{
	"anilist": AnilistIdentifier,
	"omdb":    OmdbIdentifier,
}

type (
	IdIdentifier     func(id string) (db.MetadataEntry, error)
	IdIdentifiersMap = map[string]IdIdentifier
)

var IdIdentifiers IdIdentifiersMap = IdIdentifiersMap{
	"anilist": AnilistById,
	"omdb":    OmdbIdIdentifier,
}
