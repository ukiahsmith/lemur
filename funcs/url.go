package funcs

import (
	"net/url"
	"path"
)

func AbsURL(baseURL *url.URL, p string) *url.URL {
	u := *baseURL
	u.Path = path.Join(u.Path, p)
	return &u
}
