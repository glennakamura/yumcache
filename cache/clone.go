package cache

import (
	"io"
	"net/http"
	"os"
	"time"
)

type clone struct {
	file
	signal  chan int
	header  http.Header
	content io.ReadCloser
}

func (c *clone) Read(p []byte) (int, error) {
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
		case <-c.signal:
			eof = true
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *clone) Close() error {
	return c.content.Close()
}

func (c *clone) Header() http.Header {
	return c.header
}

func newClone(path string, dl *download) Item {
	if content, err := os.Open(path); err == nil {
		return &clone{
			file:    file{path: path},
			signal:  dl.signal,
			header:  dl.header,
			content: content,
		}
	}
	return nil
}
