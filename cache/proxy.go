package cache

import (
	"net/http"
	"net/http/httputil"
	"time"
)

var proxyHandler = &httputil.ReverseProxy{
	Director:      func(*http.Request) {},
	FlushInterval: 2 * time.Second,
}
