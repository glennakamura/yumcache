package cache

import (
	"net/http"
)

type fileHandler string

func (f fileHandler) path() string {
	return string(f)
}

func (f fileHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, f.path())
}
