package cache

import (
	"io"
	"net/http"
)

type proxy struct {
	io.ReadCloser
	header http.Header
}

func (p *proxy) Header() http.Header {
	return p.header
}

func (p *proxy) Path() string {
	return ""
}

func newProxy(req *http.Request) Item {
	if resp, err := http.Get(req.RequestURI); err == nil {
		if resp.StatusCode == http.StatusOK {
			return &proxy{
				ReadCloser: resp.Body,
				header:     resp.Header,
			}
		}
		resp.Body.Close()
	}
	return nil
}
