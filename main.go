package main

import (
	"io"
	"log"
	"net/http"
	"yumcache/cache"
)

func handler(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Method + " " + req.RequestURI)
	if item := cache.Lookup(req); item != nil {
		if item.Header() == nil {
			http.ServeFile(w, req, item.Path())
			return
		}
		io.Copy(io.Writer(w), item)
		item.Close()
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
