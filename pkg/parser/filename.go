package parser

import (
	"net/url"
	"path"
	"strings"
)

// URLFilename returns the file name from a given url
func URLFilename(filename string) string {
	return path.Base(filename)
}

// PathFilename returns the file name from a given path
func PathFilename(givenPath string) string {
	return path.Base(givenPath)
}

// URLToLocalFilename returns a collision-safe filename derived from the URL
// path by flattening directory separators into underscores and dropping query
// strings and fragments.  Two URLs that share a basename but differ in path
// (e.g. /css/main.css vs /admin/css/main.css) produce distinct filenames.
func URLToLocalFilename(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return path.Base(rawURL)
	}
	p := strings.TrimPrefix(u.Path, "/")
	p = strings.ReplaceAll(p, "/", "_")
	if p == "" {
		return "index"
	}
	return p
}
