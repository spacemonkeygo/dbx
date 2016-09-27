package dbx

import (
	"bytes"
	"text/template"
)

func RenderTemplate(tmpl *template.Template, name string, data interface{}) (
	string, error) {

	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, data)
	return buf.String(), err
}
