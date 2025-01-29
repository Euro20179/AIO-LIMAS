package metadata

import (
	"fmt"

	"aiolimas/types"

	settings "aiolimas/settings"
)

type IdentifyMetadata struct {
	Title string
}

// entryType is used as a hint for where to get the metadata from
func GetMetadata(entry *db_types.InfoEntry, metadataEntry *db_types.MetadataEntry, override string) (db_types.MetadataEntry, error) {

	//anilist is still better for anime
	if settings.Settings.SonarrURL != "" && entry.Type == db_types.TY_SHOW{
		return SonarrProvider(entry)
	}

	if entry.IsAnime(){
		return AnilistShow(entry)
	}
	switch entry.Type {
	case db_types.TY_MANGA:
		return AnilistManga(entry)

	case db_types.TY_SHOW:
		fallthrough
	case db_types.TY_MOVIE:
		return OMDBProvider(entry)

	case db_types.TY_PICTURE:
		fallthrough
	case db_types.TY_MEME:
		return ImageProvider(entry)
	}
	var out db_types.MetadataEntry
	return out, nil
}

func Identify(identifySearch IdentifyMetadata, identifier string) ([]db_types.MetadataEntry, string, error) {
	fn, contains := IdentifyProviders[identifier]
	if !contains {
		return []db_types.MetadataEntry{}, "", fmt.Errorf("Invalid provider %s", identifier)
	}

	res, err := fn(identifySearch)
	return res, identifier, err
}

func GetMetadataById(id string, provider string) (db_types.MetadataEntry, error) {
	fn, contains := IdIdentifiers[provider]
	if !contains {
		return db_types.MetadataEntry{}, fmt.Errorf("Invalid provider: %s", provider)
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

type ProviderFunc func(*db_types.InfoEntry) (db_types.MetadataEntry, error)

type ProviderMap map[string]ProviderFunc

var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
	"omdb":          OMDBProvider,
	"sonarr":        SonarrProvider,
	"image":         ImageProvider,
}

type IdentifiersMap = map[string]func(info IdentifyMetadata) ([]db_types.MetadataEntry, error)

var IdentifyProviders IdentifiersMap = IdentifiersMap{
	"anilist": AnilistIdentifier,
	"omdb":    OmdbIdentifier,
	"sonarr": SonarrIdentifier,
}

type (
	IdIdentifier     func(id string) (db_types.MetadataEntry, error)
	IdIdentifiersMap = map[string]IdIdentifier
)

var IdIdentifiers IdIdentifiersMap = IdIdentifiersMap{
	"anilist": AnilistById,
	"omdb":    OmdbIdIdentifier,
	"sonarr": SonarrIdIdentifier,
}
