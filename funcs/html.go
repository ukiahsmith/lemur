package funcs

import "html/template"

func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}
