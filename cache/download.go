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
	mtime, err := http.ParseTime(d.header.Get("Last-Modified"))
	if err == nil {
		err = os.Chtimes(d.path, mtime, mtime)
	}
	return err
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
