package metadata

import (
	"fmt"

	"aiolimas/types"
)

type IdIdentifyMetadata struct {
	Id string;
	Uid int64;
}

type IdentifyMetadata struct {
	Title string;
	ForUid int64
}

type GetMetadataInfo struct {
	Entry *db_types.InfoEntry;
	MetadataEntry *db_types.MetadataEntry;
	Override string;
	Uid int64
}

// entryType is used as a hint for where to get the metadata from
func GetMetadata(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	entry := info.Entry
	if entry.IsAnime(){
		return AnilistShow(info)
	}
	switch entry.Type {
	case db_types.TY_MANGA:
		return AnilistManga(info)

	case db_types.TY_SHOW:
		fallthrough
	case db_types.TY_MOVIE:
		return OMDBProvider(info)

	case db_types.TY_PICTURE:
		fallthrough
	case db_types.TY_MEME:
		return ImageProvider(info)
	}
	return db_types.MetadataEntry{}, nil
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

type ProviderFunc func(*GetMetadataInfo) (db_types.MetadataEntry, error)

type ProviderMap map[string]ProviderFunc

var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
	"omdb":          OMDBProvider,
	"sonarr":        SonarrProvider,
	"radarr":        RadarrProvider,
	"image":         ImageProvider,
}

type IdentifiersMap = map[string]func(info IdentifyMetadata) ([]db_types.MetadataEntry, error)

var IdentifyProviders IdentifiersMap = IdentifiersMap{
	"anilist": AnilistIdentifier,
	"omdb":    OmdbIdentifier,
	"sonarr": SonarrIdentifier,
	"radarr": RadarrIdentifier,
}

type (
	IdIdentifier     func(id string) (db_types.MetadataEntry, error)
	IdIdentifiersMap = map[string]IdIdentifier
)

var IdIdentifiers IdIdentifiersMap = IdIdentifiersMap{
	"anilist": AnilistById,
	"omdb":    OmdbIdIdentifier,
	"sonarr": SonarrIdIdentifier,
	"radarr": RadarrIdIdentifier,
	"steam": SteamIdIdentifier,
}
