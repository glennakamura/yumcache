package cache

import (
	"io"
	"net/http"
)

type Item interface {
	io.ReadCloser
	Header() http.Header
	Path() string
}

type item struct{}

func (i *item) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (i *item) Close() error {
	return nil
}

func (i *item) Header() http.Header {
	return nil
}

func (i *item) Path() string {
	return ""
}
