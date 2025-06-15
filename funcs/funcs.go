package funcs

import "html/template"

func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"absURL": AbsURL,

		"mod":     Mod,
		"modBool": ModBool,
	}
}
