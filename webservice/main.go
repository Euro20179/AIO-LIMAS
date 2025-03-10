package webservice

import (
	// "aiolimas/webservice/dynamic"
	"net/http"
	"path/filepath"
	// "strings"
)

func Root(w http.ResponseWriter, req *http.Request) {
	rootPath := "./webservice/www"
	path := req.URL.Path
	// if strings.HasPrefix(path, "/html") {
	// 	dynamic.HtmlEndpoint(w, req)
	// 	return
	// }
	fullPath := filepath.Join(rootPath, path)
	http.ServeFile(w, req, fullPath)
}
