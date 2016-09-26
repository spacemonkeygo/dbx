package internal

import (
	"text/template"

	"bitbucket.org/pkg/inflect"
)

var (
	globalFuncs = template.FuncMap{
		"pluralize":   inflect.Pluralize,
		"singularize": inflect.Singularize,
		"camelize":    inflect.Camelize,
		"underscore":  inflect.Underscore,
	}
)

type Loader interface {
	Load(name string) (data []byte, err error)
}

type LoaderFunc func(name string) (data []byte, err error)

func (fn LoaderFunc) Load(name string) (data []byte, err error) {
	return fn(name)
}

type Language struct {
	tmpl *template.Template
	sql  *SQL
}

func NewLanguage(loader Loader, name string, sql *SQL) (*Language, error) {
	data, err := loader.Load(name)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("").Funcs(globalFuncs).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return &Language{
		tmpl: tmpl,
		sql:  sql,
	}, nil
}

func (lang *Language) RenderHeader(schema *Schema) (string, error) {
	return RenderTemplate(lang.tmpl, "header",
		map[string]interface{}{
			"Package": "db",
			"Dialect": "postgres",
			"Schema":  schema,
		})
}

func (lang *Language) RenderSelect(params *SelectParams) (
	string, error) {

	sql, err := lang.sql.RenderSelect(params)
	if err != nil {
		return "", err
	}

	args := []*Column{}
	if params.Where != nil {
		for _, cond := range params.Where.Conditions {
			if cond.Right == nil {
				args = append(args, cond.Left)
			}
		}
	}

	var tmpl string
	var paging_on *Column
	if params.Many {
		if params.Where != nil && params.Where.PagingOn != nil {
			tmpl = "select-paged"
			paging_on = params.Where.PagingOn
		} else {
			tmpl = "select-all"
		}
	} else {
		tmpl = "select"
	}

	return RenderTemplate(lang.tmpl, tmpl,
		map[string]interface{}{
			"SQL":      sql,
			"Table":    params.Table,
			"Args":     args,
			"PagingOn": paging_on,
		})
}

func (lang *Language) RenderDelete(params *DeleteParams) (
	string, error) {

	sql, err := lang.sql.RenderDelete(params)
	if err != nil {
		return "", err
	}

	args := []*Column{}
	if params.Where != nil {
		for _, cond := range params.Where.Conditions {
			if cond.Right == nil {
				args = append(args, cond.Left)
			}
		}
	}

	var tmpl string
	switch {
	case params.Many:
		tmpl = "delete-all"
	default:
		tmpl = "delete"
	}

	return RenderTemplate(lang.tmpl, tmpl,
		map[string]interface{}{
			"SQL":   sql,
			"Table": params.Table,
			"Args":  args,
		})
}
