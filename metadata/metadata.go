package metadata

import (
	"fmt"

	"aiolimas/logging"
	"aiolimas/settings"
	"aiolimas/types"
)

type IdIdentifyMetadata struct {
	Id  string
	Uid int64
}

type IdentifyMetadata struct {
	Title  string
	ForUid int64
}

type GetMetadataInfo struct {
	Entry         *db_types.InfoEntry
	MetadataEntry *db_types.MetadataEntry
	Override      string
	Uid           int64
}

// entryType is used as a hint for where to get the metadata from
func GetMetadata(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	entry := info.Entry
	if entry.IsAnime() {
		return AnilistShow(info)
	}

	if entry.Format == db_types.F_STEAM {
		return SteamProvider(info)
	}

	switch entry.Type {
	case db_types.TY_GAME:
		return SteamProvider(info)

	case db_types.TY_MANGA:
		return AnilistManga(info)

	case db_types.TY_BOOK:
		return GoogleBooksProvider(info)

	case db_types.TY_MOVIE_SHORT:
		fallthrough
	case db_types.TY_DOCUMENTARY:
		fallthrough
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
		return []db_types.MetadataEntry{}, "", fmt.Errorf("invalid provider %s", identifier)
	}

	res, err := fn(identifySearch)
	return res, identifier, err
}

func GetMetadataById(id string, foruid int64, provider string) (db_types.MetadataEntry, error) {
	fn, contains := IdIdentifiers[provider]
	if !contains {
		return db_types.MetadataEntry{}, fmt.Errorf("invalid provider: %s", provider)
	}

	us, err := settings.GetUserSettings(foruid)
	if err != nil {
		return db_types.MetadataEntry{}, err
	}
	return fn(id, us)
}

func DetermineBestLocationProvider(info *db_types.InfoEntry, metadata *db_types.MetadataEntry) string {
	if info.Format&db_types.F_STEAM == db_types.F_STEAM {
		return "steam"
	}

	if metadata.Provider == "sonarr" {
		return "sonarr"
	}

	// could not determine
	return ""
}

func GetLocation(providerID string, foruid int64, provider string) (string, error) {
	fn, contains := LocationFinders[provider]
	if !contains {
		return "", fmt.Errorf("invalid provider: %s", provider)
	}

	us, err := settings.GetUserSettings(foruid)
	if err != nil {
		return "", err
	}

	logging.Info(fmt.Sprintf("location lookup using provider: %s", provider))

	return fn(&us, providerID)
}

func ListMetadataProviders() []string {
	keys := make([]string, 0, len(Providers))
	for k := range Providers {
		keys = append(keys, k)
	}
	return keys
}

func IsValidLocationProvider(name string) bool {
	_, contains := LocationFinders[name]
	return contains
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

// parameters: userSettings item_metadata
type LocationFunc func(*settings.SettingsData, string) (string, error)

type LocationMap map[string]LocationFunc

var LocationFinders LocationMap = LocationMap{
	"steam":  SteamLocationFinder,
	"sonarr": SonarrGetLocation,
}

type ProviderFunc func(*GetMetadataInfo) (db_types.MetadataEntry, error)

type ProviderMap map[string]ProviderFunc

// uses an entry's heuristics to find the correct metadata
var Providers ProviderMap = ProviderMap{
	"anilist":       AnlistProvider,
	"anilist-manga": AnilistManga,
	"anilist-show":  AnilistShow,
	"omdb":          OMDBProvider,
	"sonarr":        SonarrProvider,
	"radarr":        RadarrProvider,
	"image":         ImageProvider,
	"steam":         SteamProvider,
	"googlebooks":   GoogleBooksProvider,
}

type IdentifiersMap = map[string]func(info IdentifyMetadata) ([]db_types.MetadataEntry, error)

// uses a search query
var IdentifyProviders IdentifiersMap = IdentifiersMap{
	"anilist": AnilistIdentifier,
	"omdb":    OmdbIdentifier,
	"sonarr":  SonarrIdentifier,
	"radarr":  RadarrIdentifier,
	"steam":   SteamIdentifier,
	"googlebooks": GoogleBooksIdentifier,
}

type (
	IdIdentifier     func(id string, us settings.SettingsData) (db_types.MetadataEntry, error)
	IdIdentifiersMap = map[string]IdIdentifier
)

// does an id lookup
var IdIdentifiers IdIdentifiersMap = IdIdentifiersMap{
	// anilist id
	"anilist": AnilistById,
	// imdb id (without the tt)
	"omdb": OmdbIdIdentifier,
	// sonarr id
	"sonarr": SonarrIdIdentifier,
	// radarr id
	"radarr": RadarrIdIdentifier,
	// steam id
	"steam": SteamIdIdentifier,
	// isbn
	"openlibrary": OpenLibraryIdIdentifier,
	// isbn
	"googlebooks": GoogleBooksIdIdentifier,
}
