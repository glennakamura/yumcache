package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Method + " " + req.RequestURI)
	if info := cacheLookup(req); info != nil {
		if info.path != "" {
			http.ServeFile(w, req, info.path)
			return
		}
		if info.content != nil {
			out := io.Writer(w)
			if info.output != nil {
				out = io.MultiWriter(out, info.output)
			}
			done := false
			for {
				io.Copy(out, info.content)
				if done || info.signals == nil {
					break
				} else {
					select {
					case <-info.signals:
						done = true
					case <-time.After(5 * time.Second):
					}
				}
			}
			info.content.Close()
			if info.done != nil {
				info.done()
			}
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
