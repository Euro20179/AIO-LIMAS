package api

import (
	"aiolimas/db"
	"errors"
	"fmt"
	"net/http"
	"os"
)

func ThumbnailResource(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	item := pp["id"].(db.InfoEntry)

	//this should gauranteed exist because we panic if AIO_DIR couldn't be set
	aioPath := os.Getenv("AIO_DIR")

	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/item-%d", aioPath, item.ItemId)

	if _, err := os.Stat(itemThumbnailPath); errors.Is(err, os.ErrNotExist) {
		wError(w, 404, "Item does not have a local thumbnail")
		return
	}

	http.ServeFile(w, req, itemThumbnailPath)
}


//TODO:
//add DownloadThumbnail
//take itemId as input
//read the thumbnail off it
//if it's a remote url, download it in a place ThumbnailResource can find
//if it's a data:image/*;base64,* url
//extract the data and save it to a place ThumbnailResource can find
