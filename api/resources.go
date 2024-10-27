package api

import (
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"aiolimas/db"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (self gzipResponseWriter) Write(b []byte) (int, error) {
	return self.Writer.Write(b)
}

func gzipMiddleman(fn func(w http.ResponseWriter, req *http.Request, pp ParsedParams)) func(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	return func(w http.ResponseWriter, r *http.Request, pp ParsedParams) {
		var gz *gzip.Writer

		acceptedEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptedEncoding, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz = gzip.NewWriter(w)
			defer gz.Close()
			w = gzipResponseWriter{Writer: gz, ResponseWriter: w}
		}

		fn(w, r, pp)
	}
}

func thumbnailResource(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	item := pp["id"].(db.InfoEntry)

	aioPath := os.Getenv("AIO_DIR")
	// this should gauranteed exist because we panic if AIO_DIR couldn't be set
	if aioPath == "" {
		panic("$AIO_DIR should not be empty")
	}

	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/item-%d", aioPath, item.ItemId)

	if _, err := os.Stat(itemThumbnailPath); errors.Is(err, os.ErrNotExist) {
		wError(w, 404, "Item does not have a local thumbnail")
		return
	}

	http.ServeFile(w, req, itemThumbnailPath)
}

var ThumbnailResource = gzipMiddleman(thumbnailResource)

func DownloadThumbnail(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	item := pp["id"].(db.MetadataEntry)

	thumb := item.Thumbnail

	if thumb == "" {
		wError(w, 403, "There is no thumbnail for this entry, cannot download it")
		return
	}

	aioPath := os.Getenv("AIO_DIR")

	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/item-%d", aioPath, item.ItemId)

	if strings.HasPrefix(thumb, "data:") {
		_, after, found := strings.Cut(thumb, "base64,")
		if !found {
			wError(w, 403, "Thumbnail is encoded in base64")
			return
		}

		data, err := base64.StdEncoding.DecodeString(after)
		if err != nil {
			wError(w, 500, "Could not decode base64\n%s", err.Error())
			return
		}

		err = os.WriteFile(itemThumbnailPath, data, 0o644)
		if err != nil {
			wError(w, 500, "Could not save thumbnail\n%s", err.Error())
			return
		}

		success(w)
	}

	client := http.Client{}
	resp, err := client.Get(thumb)
	if err != nil {
		wError(w, 500, "Failed to download thumbnail\n%s", err.Error())
		return
	}

	defer resp.Body.Close()

	file, err := os.OpenFile(itemThumbnailPath, os.O_CREATE|os.O_WRONLY, 0o664)
	if err != nil {
		wError(w, 500, "Failed to open thumbnail file location\n%s", err.Error())
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		wError(w, 500, "Failed to write thumbnail to file\n%s", err.Error())
		return
	}
	success(w)
}
