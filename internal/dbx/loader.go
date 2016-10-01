package dbx

import (
	"io/ioutil"
	"path/filepath"
	"text/template"

	"bitbucket.org/pkg/inflect"
)

type Loader interface {
	Load(name string) (*template.Template, error)
}

type LoaderFunc func(name string) (*template.Template, error)

func (fn LoaderFunc) Load(name string) (*template.Template, error) {
	return fn(name)
}

type DirLoader string

func (d DirLoader) Load(name string) (*template.Template, error) {
	data, err := ioutil.ReadFile(filepath.Join(string(d), name))
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return loadTemplate(name, data)
}

type BinLoader func(name string) ([]byte, error)

func (b BinLoader) Load(name string) (*template.Template, error) {
	data, err := b(name)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	return loadTemplate(name, data)
}

func loadTemplate(name string, data []byte) (*template.Template, error) {
	globalFuncs := template.FuncMap{
		"pluralize":   inflect.Pluralize,
		"singularize": inflect.Singularize,
		"camelize":    inflect.Camelize,
		"underscore":  inflect.Underscore,
	}

	tmpl, err := template.New(name).Funcs(globalFuncs).Parse(string(data))
	return tmpl, Error.Wrap(err)
}
