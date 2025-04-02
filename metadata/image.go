package metadata

import (
	"fmt"
	"os"

	"aiolimas/types"
)

func ImageProvider(info *GetMetadataInfo) (db_types.MetadataEntry, error) {
	var out db_types.MetadataEntry
	location := info.Entry.Location

	aioPath := os.Getenv("AIO_DIR")
	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/item-%d", aioPath, info.Entry.ItemId)

	err := os.Symlink(location, itemThumbnailPath)
	if err != nil{
		return out, err
	}

	out.Thumbnail = fmt.Sprintf("/api/v1/resource/thumbnail?id=%d", info.Entry.ItemId)

	return out, nil
}
