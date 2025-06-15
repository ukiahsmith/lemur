package lemur

import (
	"fmt"
	"io"
	"strings"
)

// Srender renders the specified template by name with the given data and returns
// the output as a string.
//
// It is a convenience wrapper around the Render method. Use Srender when you
// need the template output as a string, for example, to pass to another
// function, store in a variable, or when an io.Writer is not readily available.
// If direct writing to an output stream (like an http.ResponseWriter) is
// possible, using Render directly might be more efficient as it avoids the
// intermediate string allocation.
func (wh *Lemur) Srender(tmplName string, data interface{}) (string, error) {
	var buf strings.Builder

	err := wh.Render(&buf, tmplName, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Render executes the specified template by name, writing the output to the
// provided io.Writer.
//
// tmplName specifies which named layout set to use (e.g., "_defaults", "mytemplate").
// If tmplName is an empty string, it defaults to "_defaults".
// The method will then execute the "_index.html.tmpl" template within that layout set.
//
// data is the data to be passed to the template for rendering.
//
// This is the primary method for rendering templates when you have an output
// stream, such as an http.ResponseWriter or a file.
func (wh *Lemur) Render(w io.Writer, tmplName string, data interface{}) error {
	if tmplName == "" {
		tmplName = "_defaults"
	}

	if tmpl, ok := wh.layouts[tmplName]; ok {
		err := tmpl.ExecuteTemplate(w, "_index.html.tmpl", data)
		if err != nil {
			return fmt.Errorf("lemur Render: could not render template: %w", err)
		}
	} else {
		return fmt.Errorf("lemur Render: no template with name %q", tmplName)
	}

	return nil
}
