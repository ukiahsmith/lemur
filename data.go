package lemur

import "net/url"

type Data struct {
	Site Site
	Page Page
}

type Site struct {
	BaseURL *url.URL

	Title     string
	Copyright string
}

type Page struct {
	Title string
	Data  map[string]interface{}
	Form  map[string]interface{}
}
