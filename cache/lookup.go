package cache

import (
	"net/http"
	"os"
	"sync"
)

var downloads = make(map[string]*download)
var mutex sync.Mutex

func Lookup(req *http.Request) http.Handler {
	path := localPath(req)
	isCompleteFile := req.Header.Get("Range") == ""
	mutex.Lock()
	defer mutex.Unlock()
	if dl, ok := downloads[path]; ok {
		if isCompleteFile {
			return concurrentHandler(path, dl)
		}
		return proxyHandler
	}
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return fileHandler(path)
	}
	if isCompleteFile && path[len(path)-4:] == ".rpm" {
		return downloadHandler(path)
	}
	return proxyHandler
}
