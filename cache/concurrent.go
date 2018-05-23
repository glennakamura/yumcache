package cache

import (
	"io"
	"net/http"
	"os"
	"time"
)

type concurrent struct {
	content io.ReadCloser
	dl      *download
	wait    chan bool
	eof     chan bool
}

func (c *concurrent) Read(p []byte) (int, error) {
	eof := false
	for {
		n, err := c.content.Read(p)
		if err != nil && err != io.EOF {
			return n, err
		}
		if n > 0 {
			return n, nil
		}
		if eof {
			return 0, io.EOF
		}
		select {
		case <-c.eof:
			eof = true
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *concurrent) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	<-c.wait
	copyHeader(w.Header(), c.dl.header)
	w.WriteHeader(c.dl.status)
	io.Copy(io.Writer(w), io.Reader(c))
	c.content.Close()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func concurrentHandler(path string, dl *download) http.Handler {
	if content, err := os.Open(path); err == nil {
		return &concurrent{
			content: content,
			dl:      dl,
			wait:    dl.wait,
			eof:     dl.eof,
		}
	}
	return proxyHandler
}
