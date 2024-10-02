package metadata

import (
	"aiolimas/db"
	"fmt"
	"os"
)

func ImageProvider(entry *db.InfoEntry, metadata *db.MetadataEntry) (db.MetadataEntry, error) {
	var out db.MetadataEntry
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
