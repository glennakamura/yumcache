package config

import (
	"path"
)

var (
	AppHome      string
	DocumentRoot string
	Private      string
)

func init() {
	setAppHome("/var/cache/yumcache")
}

func setAppHome(dirpath string) {
	AppHome = dirpath
	DocumentRoot = path.Join(dirpath, "www")
	Private = path.Join(dirpath, ".private")
}
