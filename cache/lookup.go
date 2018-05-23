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
	mutex.Lock()
	defer mutex.Unlock()
	if dl, ok := downloads[path]; ok {
		return concurrentHandler(path, dl)
	}
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return fileHandler(path)
	}
	if path[len(path)-4:] == ".rpm" {
		return downloadHandler(path)
	}
	return proxyHandler
}
