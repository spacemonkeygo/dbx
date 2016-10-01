package dbx

import (
	"bytes"
	"io"
	"text/template"
)

func RenderTemplateString(tmpl *template.Template, name string,
	data interface{}) (string, error) {

	var buf bytes.Buffer
	if err := RenderTemplate(tmpl, &buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderTemplate(tmpl *template.Template, w io.Writer, name string,
	data interface{}) error {

	if name == "" {
		return Error.Wrap(tmpl.Execute(w, data))
	}
	return Error.Wrap(tmpl.ExecuteTemplate(w, name, data))
}
