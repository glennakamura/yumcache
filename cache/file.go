package cache

type file struct {
	item
	path string
}

func (f *file) Path() string {
	return f.path
}

func newFile(path string) Item {
	return &file{path: path}
}
