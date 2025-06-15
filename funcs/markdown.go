package funcs

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func Markdown(s string) string {
	var buf bytes.Buffer

	if err := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // required to add `id` attributes to headers
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	).Convert([]byte(s), &buf); err != nil {
		fmt.Printf("error converting to markdown: %s", err)
	}

	return buf.String()
}
