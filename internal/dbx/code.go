package dbx

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

	r.render(r.lang.RenderHeader(schema))

	for _, table := range schema.Tables {
		r.renderTable(table)
	}

	return r.err
}

func equalsConditionsForColumns(columns []*Column) []*ConditionParams {
	conditions := make([]*ConditionParams, 0, len(columns))
	for _, column := range columns {
		conditions = append(conditions, EqualsQ(column))
	}
	return conditions
}

func (r *codeRenderer) renderTable(table *Table) {
	r.render(r.lang.RenderInsert(&InsertParams{
		Table:   table,
		Columns: table.InsertableColumns(),
	}))

	for _, query := range table.Queries {
		var joins []*LeftJoinParams
		for _, relation := range query.Joins {
			joins = append(joins, LeftJoin(relation))
		}

		start := equalsConditionsForColumns(query.Start)
		end := equalsConditionsForColumns(query.End)
		many := !table.ColumnSetUnique(query.Start)

		var conditions []*ConditionParams
		conditions = append(conditions, start...)
		conditions = append(conditions, end...)
		r.render(r.lang.RenderSelect(&SelectParams{
			Table:     table,
			Many:      many,
			LeftJoins: joins,
			Where:     Where(conditions...),
		}))
		r.render(r.lang.RenderDelete(&DeleteParams{
			Table:     table,
			Many:      many,
			LeftJoins: joins,
			Where:     Where(conditions...),
		}))

		if many && len(query.Start) == 1 {
			r.render(r.lang.RenderSelect(&SelectParams{
				Table:     table,
				Many:      true,
				LeftJoins: joins,
				Where: &WhereParams{
					Conditions: end,
					PagingOn:   query.Start[0],
				},
			}))
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
