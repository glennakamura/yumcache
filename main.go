package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"yumcache/cache"
	"yumcache/cert"
	"yumcache/config"
)

type handler struct {
	cert.CA
	fs http.Handler
}

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println(req.Method + " " + req.Host + " " + req.URL.Path)
	if req.Method == http.MethodConnect {
		h.serveConnect(w, req)
		return
	}
	if req.URL.Host == "" {
		if hostPort(req.Host) == "8080" {
			h.fs.ServeHTTP(w, req)
			return
		}
		req.URL.Scheme = "https"
		req.URL.Host = req.Host
	}
	cache.Lookup(req).ServeHTTP(w, req)
}

func (h handler) serveConnect(w http.ResponseWriter, req *http.Request) {
	clientConn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacker not implemented",
			http.StatusNotImplemented)
		return
	}
	hijackConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	_, err = hijackConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	if err != nil {
		hijackConn.Close()
		return
	}
	tlsConfig := &tls.Config{
		GetCertificate: func(
			info *tls.ClientHelloInfo,
		) (*tls.Certificate, error) {
			serverName := hostName(req.Host)
			if info.ServerName != "" {
				serverName = info.ServerName
			}
			cert, err := h.GenerateCert(serverName)
			return cert.TLS(), err
		},
	}
	tlsConn := tls.Server(hijackConn, tlsConfig)
	if err = tlsConn.Handshake(); err == nil {
		go io.Copy(clientConn, tlsConn)
		io.Copy(tlsConn, clientConn)
	}
	tlsConn.Close()
}

func hostName(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return ""
}

func hostPort(addr string) string {
	if _, port, err := net.SplitHostPort(addr); err == nil {
		return port
	}
	return ""
}

func newHandler() handler {
	err := os.MkdirAll(config.DocumentRoot, 0755)
	if err == nil {
		err = os.MkdirAll(config.Private, 0700)
		if err == nil {
			certFile := path.Join(config.DocumentRoot, "ca.pem")
			keyFile := path.Join(config.Private, "ca.key")
			name := "YUM Cache CA"
			ca, err := cert.LoadCA(certFile, keyFile, name)
			if err == nil {
				fs := http.FileServer(
					http.Dir(config.DocumentRoot),
				)
				return handler{CA: ca, fs: fs}
			}
		}
	}
	panic(err)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Fatal(http.ListenAndServe(":8080", newHandler()))
}
