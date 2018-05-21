package cache

import (
	"net/http"
	"os"
	"path"
	"regexp"
)

const centos = `^.*(?i:centos).*(/[1-9](?:\.?[0-9])*/[a-z]+/x86_64/.*$)`
const epel = `^.*(?:epel).*(/[1-9](?:\.?[0-9])*/x86_64/.*$)`

var mirrors = map[string]*regexp.Regexp{
	"centos": regexp.MustCompile(centos),
	"epel":   regexp.MustCompile(epel),
}

var basedir = "/var/cache/yumcache"

func localPath(req *http.Request) string {
	url := path.Join(req.Host, path.Clean(req.URL.Path))
	for dist, re := range mirrors {
		if match := re.FindStringSubmatch(url); match != nil {
			return path.Join(basedir, dist, match[1])
		}
	}
	return path.Join(basedir, url)
}

func createPath(name string) (*os.File, error) {
	os.MkdirAll(path.Dir(name), 0755)
	return os.Create(name)
}
