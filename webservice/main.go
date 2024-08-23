package webservice

import (
	"net/http"
	"path/filepath"
)

func Root(w http.ResponseWriter, req *http.Request) {
	rootPath := "./webservice/www"
	path := req.URL.Path
	fullPath := filepath.Join(rootPath, path)
	http.ServeFile(w, req, fullPath)
}
