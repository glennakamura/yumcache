package main

import (
	"log"
	"net/http"
	"yumcache/cache"
)

func handler(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Method + " " + req.RequestURI)
	cache.Lookup(req).ServeHTTP(w, req)
}

func main() {
	http.HandleFunc("/", handler)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
