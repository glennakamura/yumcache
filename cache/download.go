package cache

import (
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

type download struct {
	io.Reader
	path   string
	status int
	length int64
	header http.Header
	body   io.ReadCloser
	output io.WriteCloser
	wait   chan bool
	eof    chan bool
}

func (d *download) Close() error {
	mutex.Lock()
	defer mutex.Unlock()
	d.body.Close()
	d.output.Close()
	close(d.eof)
	delete(downloads, d.path)
	if d.valid() {
		mtime, err := http.ParseTime(d.header.Get("Last-Modified"))
		if err == nil {
			err = os.Chtimes(d.path, mtime, mtime)
		}
		return err
	}
	return os.Remove(d.path)
}

func (d *download) valid() bool {
	if d.status == http.StatusOK {
		if d.length < 0 {
			return true
		}
		if info, err := os.Stat(d.path); err == nil {
			if d.length == info.Size() {
				return true
			}
		}
	}
	return false
}

func downloadHandler(path string) http.Handler {
	if output, err := createPath(path); err == nil {
		dl := &download{
			path:   path,
			output: output,
			wait:   make(chan bool),
			eof:    make(chan bool),
		}
		downloads[path] = dl
		return &httputil.ReverseProxy{
			Director:      func(*http.Request) {},
			FlushInterval: 2 * time.Second,
			ModifyResponse: func(resp *http.Response) error {
				dl.Reader = io.TeeReader(resp.Body, output)
				dl.status = resp.StatusCode
				dl.length = resp.ContentLength
				dl.header = resp.Header
				dl.body = resp.Body
				resp.Body = dl
				close(dl.wait)
				return nil
			},
		}
	}
	return proxyHandler
}
