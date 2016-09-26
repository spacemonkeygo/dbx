package internal

import (
	"bytes"
	"text/template"
)

func RenderToString(tmpl *template.Template, name string, data interface{}) (
	string, error) {

	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, data)
	return buf.String(), err
}
