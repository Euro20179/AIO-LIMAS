package api

import (
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"aiolimas/types"
	"aiolimas/util"
	"aiolimas/logging"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (self gzipResponseWriter) Write(b []byte) (int, error) {
	return self.Writer.Write(b)
}

func gzipMiddleman(fn func(w http.ResponseWriter, req *http.Request, pp ParsedParams)) func(ctx RequestContext) {
	return func(ctx RequestContext) {
		r := ctx.Req
		w := ctx.W
		pp := ctx.PP

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

func serveThumbnail(w http.ResponseWriter, req *http.Request, path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		util.WError(w, 404, "Thumbnail does not exist\n")
		return
	}

	http.ServeFile(w, req, path)
}

func thumbnailResource(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	hash := pp["hash"].(string)

	aioPath := os.Getenv("AIO_DIR")
	// this should gauranteed exist because we panic if AIO_DIR couldn't be set
	if aioPath == "" {
		panic("$AIO_DIR should not be empty")
	}

	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/%c/%s", aioPath, hash[0], hash)

	serveThumbnail(w, req, itemThumbnailPath)
}

func thumbnailResourceLegacy(w http.ResponseWriter, req *http.Request, pp ParsedParams) {
	id := pp["id"].(string)

	aioPath := os.Getenv("AIO_DIR")
	// this should gauranteed exist because we panic if AIO_DIR couldn't be set
	if aioPath == "" {
		panic("$AIO_DIR should not be empty")
	}

	itemThumbnailPath := fmt.Sprintf("%s/thumbnails/item-%s", aioPath, id)

	serveThumbnail(w, req, itemThumbnailPath)
}

var (
	ThumbnailResource       = gzipMiddleman(thumbnailResource)
	ThumbnailResourceLegacy = gzipMiddleman(thumbnailResourceLegacy)
)

func DownloadThumbnail(ctx RequestContext) {
	pp := ctx.PP
	w := ctx.W

	item := pp["id"].(db_types.MetadataEntry)

	thumb := item.Thumbnail

	if thumb == "" {
		util.WError(w, 403, "There is no thumbnail for this entry, cannot download it")
		return
	}

	aioPath := os.Getenv("AIO_DIR")

	thumbnailPath := fmt.Sprintf("%s/thumbnails", aioPath)

	if strings.HasPrefix(thumb, "data:") {
		_, after, found := strings.Cut(thumb, "base64,")
		if !found {
			util.WError(w, 403, "Thumbnail is encoded in base64")
			return
		}

		data, err := base64.StdEncoding.DecodeString(after)
		if err != nil {
			util.WError(w, 500, "Could not decode base64\n%s", err.Error())
			return
		}

		h := sha1.New()
		h.Sum(data)
		shaSum := h.Sum(nil)

		sumHex := hex.EncodeToString(shaSum)

		itemThumbnailPath := fmt.Sprintf("%s/%c/%s", thumbnailPath, sumHex[0], sumHex)

		// path alr exists, no need to write it again
		if _, err := os.Stat(itemThumbnailPath); err == nil {
			goto done
		}

		err = os.WriteFile(itemThumbnailPath, data, 0o644)
		if err != nil {
			util.WError(w, 500, "Could not save thumbnail\n%s", err.Error())
			return
		}

	done:
		w.Write([]byte(sumHex))
		return
	}

	client := http.Client{}
	resp, err := client.Get(thumb)
	if err != nil {
		util.WError(w, 500, "Failed to download thumbnail\n%s", err.Error())
		return
	}

	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		util.WError(w, 500, "Failed to download thumbnail from url\n%s", err.Error())
		return
	}

	shaSum := sha1.Sum(out)
	sumHex := hex.EncodeToString(shaSum[:])

	thumbnailPath = fmt.Sprintf("%s/%c", thumbnailPath, sumHex[0])

	if err := os.MkdirAll(thumbnailPath, 0o700); err != nil {
		util.WError(w, 500, "Failed to create thumbnail dir")
		logging.ELog(err)
		return
	}

	itemThumbnailPath := fmt.Sprintf("%s/%s", thumbnailPath, sumHex)

	// path alr exists, no need to write it again
	if _, err := os.Stat(itemThumbnailPath); err == nil {
		w.Write([]byte(sumHex))
		return
	}

	file, err := os.OpenFile(itemThumbnailPath, os.O_CREATE|os.O_WRONLY, 0o664)
	if err != nil {
		util.WError(w, 500, "Failed to open thumbnail file location\n%s", err.Error())
		return
	}

	_, err = file.Write(out)
	if err != nil {
		util.WError(w, 500, "Failed to save thumbnail\n%s", err.Error())
		return
	}

	w.Write([]byte(sumHex))
}
