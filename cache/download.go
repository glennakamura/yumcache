package cache

import (
	"io"
	"net/http"
)

type download struct {
	io.Reader
	file
	header  http.Header
	content io.ReadCloser
	output  io.WriteCloser
	signal  chan int
}

func (d *download) Close() error {
	mutex.Lock()
	err := d.content.Close()
	d.output.Close()
	close(d.signal)
	delete(downloads, d.path)
	mutex.Unlock()
	return err
}

func (d *download) Header() http.Header {
	return d.header
}

func newDownload(path string, req *http.Request) Item {
	if resp, err := http.Get(req.RequestURI); err == nil {
		if resp.StatusCode == http.StatusOK {
			if w, err := createPath(path); err == nil {
				dl := &download{
					Reader:  io.TeeReader(resp.Body, w),
					file:    file{path: path},
					header:  resp.Header,
					content: resp.Body,
					output:  w,
					signal:  make(chan int),
				}
				downloads[path] = dl
				return dl
			}
		}
		resp.Body.Close()
	}
	return nil
}
