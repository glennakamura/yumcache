package cache

import (
	"net/http"
	"os"
	"sync"
)

var downloads = make(map[string]*download)
var mutex sync.Mutex

func Lookup(req *http.Request) Item {
	path := localPath(req)
	mutex.Lock()
	defer mutex.Unlock()
	if dl, ok := downloads[path]; ok {
		return newClone(path, dl)
	}
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return newFile(path)
	}
	if path[len(path)-4:] == ".rpm" {
		return newDownload(path, req)
	}
	return newProxy(req)
}
