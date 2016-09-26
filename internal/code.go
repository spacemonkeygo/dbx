package internal

import (
	"fmt"
	"io"
	"text/template"
)

func RenderCode(schema *Schema, lang *Language, w io.Writer) (err error) {

	renderer := newCodeRenderer(w, lang)
	return renderer.RenderCode(schema)
}

type codeRenderer struct {
	w    io.Writer
	lang *Language
	err  error
}

func newCodeRenderer(w io.Writer, lang *Language) *codeRenderer {
	return &codeRenderer{
		w:    w,
		lang: lang,
	}
}

func (r *codeRenderer) RenderCode(schema *Schema) (err error) {
	r.err = nil
	defer func() {
		err = r.err
	}()

	r.render(r.lang.RenderHeader(schema))

	for _, table := range schema.Tables {
		r.renderTable(table)
	}

	return
}

func (r *codeRenderer) renderTable(table *Table) {
	unique := table.Unique()

	r.render(r.lang.RenderDelete(&DeleteParams{
		Table: table,
		Many:  true,
	}))
	r.render(r.lang.RenderSelect(&SelectParams{
		Table: table,
		Many:  true,
	}))

	for _, by := range unique {
		r.render(r.lang.RenderDelete(&DeleteParams{
			Table: table,
			Many:  false,
			Where: Where(EqualsQ(by)),
		}))
		r.render(r.lang.RenderSelect(&SelectParams{
			Table: table,
			Many:  false,
			Where: Where(EqualsQ(by)),
		}))
		r.render(r.lang.RenderSelect(&SelectParams{
			Table: table,
			Many:  true,
			Where: &WhereParams{PagingOn: by},
		}))
	}

	for _, chain := range table.RelationChains() {
		var joins []*LeftJoinParams
		for _, relation := range chain {
			joins = append(joins, LeftJoin(relation))
		}

		for _, foreign_by := range chain[len(chain)-1].ForeignColumn.Table.Unique() {
			r.render(r.lang.RenderDelete(&DeleteParams{
				Table:     table,
				Many:      true,
				LeftJoins: joins,
				Where:     Where(EqualsQ(foreign_by)),
			}))
			r.render(r.lang.RenderSelect(&SelectParams{
				Table:     table,
				Many:      true,
				LeftJoins: joins,
				Where:     Where(EqualsQ(foreign_by)),
			}))
			for _, by := range unique {
				r.render(r.lang.RenderDelete(&DeleteParams{
					Table:     table,
					Many:      false,
					LeftJoins: joins,
					Where:     Where(EqualsQ(foreign_by), EqualsQ(by)),
				}))
				r.render(r.lang.RenderSelect(&SelectParams{
					Table:     table,
					Many:      false,
					LeftJoins: joins,
					Where:     Where(EqualsQ(foreign_by), EqualsQ(by)),
				}))
				r.render(r.lang.RenderSelect(&SelectParams{
					Table:     table,
					Many:      true,
					LeftJoins: joins,
					Where: &WhereParams{
						Conditions: []*ConditionParams{
							EqualsQ(foreign_by),
						},
						PagingOn: by,
					},
				}))
			}
		}
	}
}

func (r *codeRenderer) render(code string, err error) {
	if !r.setError(err) {
		return
	}
	r.printf("%s\n", code)
}

func (r *codeRenderer) printTemplate(tmpl *template.Template, name string,
	data interface{}) (ok bool) {
	rendered, err := RenderTemplate(tmpl, name, data)
	if r.setError(err) {
		return false
	}
	return r.printf("%s\n", rendered)
}

func (r *codeRenderer) printf(format string, args ...interface{}) (ok bool) {
	if r.err != nil {
		return false
	}
	_, err := fmt.Fprintf(r.w, format, args...)
	return r.setError(err)
}

func (r *codeRenderer) setError(err error) (ok bool) {
	if err != nil && r.err == nil {
		r.err = err
	}
	return r.err == nil
}
