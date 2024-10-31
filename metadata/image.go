package metadata

import (
	"fmt"
	"os"

	"aiolimas/types"
)

func ImageProvider(entry db_types.InfoEntry) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry
	location := entry.Location

	aioPath := os.Getenv("AIO_DIR")
	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/item-%d", aioPath, entry.ItemId)

	err := os.Symlink(location, itemThumbnailPath)
	if err != nil{
		return out, err
	}

	out.Thumbnail = fmt.Sprintf("/api/v1/resource/thumbnail?id=%d", entry.ItemId)

	return out, nil
}
