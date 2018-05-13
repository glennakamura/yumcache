package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"
)

const centos = `^.*(?i:centos).*(/[1-9](?:\.?[0-9])*/[a-z]+/x86_64/.*$)`

var mirrors = map[string]*regexp.Regexp{
	"centos": regexp.MustCompile(centos),
}

type cache_download struct {
	signals chan int
	header  *http.Header
}

type cache_info struct {
	signals chan int
	header  *http.Header
	content io.ReadCloser
	output  io.Writer
	path    string
	done    func()
}

var cache_dirname = "/var/cache/yumcache"
var cache_downloads = make(map[string]*cache_download)
var cache_mutex sync.Mutex

var errNotFound = errors.New("not found in cache")

func cachePath(req *http.Request) string {
	for dist, re := range mirrors {
		uri_path := re.FindStringSubmatch(req.RequestURI)
		if uri_path != nil {
			return path.Join(cache_dirname, dist, uri_path[1])
		}
	}
	return path.Join(cache_dirname, req.Host, path.Clean(req.URL.Path))
}

func cacheLookup(req *http.Request) *cache_info {
	cache_path := cachePath(req)
	cache_mutex.Lock()
	defer cache_mutex.Unlock()
	if download, ok := cache_downloads[cache_path]; ok {
		return cacheDownload(cache_path, download)
	}
	if stat, err := os.Stat(cache_path); err == nil && !stat.IsDir() {
		return cacheFile(cache_path)
	}
	if cache_path[len(cache_path)-4:] == ".rpm" {
		return cacheNewFile(cache_path, req)
	}
	return cachePassthrough(req)
}

func cacheFile(cache_path string) *cache_info {
	log.Println("cacheFile: " + cache_path)
	return &cache_info{path: cache_path}
}

func cacheNewFile(cache_path string, req *http.Request) *cache_info {
	log.Println("cacheNewFile: " + cache_path)
	if resp, err := http.Get(req.RequestURI); err == nil {
		if resp.StatusCode == 200 {
			os.MkdirAll(path.Dir(cache_path), 0755)
			if file, err := os.Create(cache_path); err == nil {
				download := &cache_download{
					signals: make(chan int),
					header:  &resp.Header,
				}
				cache_downloads[cache_path] = download
				return &cache_info{
					header:  &resp.Header,
					content: resp.Body,
					output:  file,
					done: func() {
						cache_mutex.Lock()
						file.Close()
						close(download.signals)
						delete(cache_downloads, cache_path)
						cache_mutex.Unlock()
					},
				}
			}
		}
		resp.Body.Close()
	}
	return nil
}

func cacheDownload(cache_path string, download *cache_download) *cache_info {
	log.Println("cacheDownload: " + cache_path)
	if file, err := os.Open(cache_path); err == nil {
		return &cache_info{
			signals: download.signals,
			header:  download.header,
			content: file,
			done: func() {
				file.Close()
			},
		}
	}
	return nil
}

func cachePassthrough(req *http.Request) *cache_info {
	log.Println("cachePassthrough")
	if resp, err := http.Get(req.RequestURI); err == nil {
		if resp.StatusCode == 200 {
			return &cache_info{
				header:  &resp.Header,
				content: resp.Body,
			}
		}
		resp.Body.Close()
	}
	return nil
}
