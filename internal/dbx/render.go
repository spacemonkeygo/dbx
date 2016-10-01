// Copyright (C) 2016 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbx

import (
	"fmt"
	"io"
)

func RenderCode(w io.Writer, schema *Schema, dialect Dialect, lang Language) (
	err error) {

	renderer := newCodeRenderer(w, dialect, lang)
	return renderer.RenderCode(schema)
}

type codeRenderer struct {
	w            io.Writer
	dialect      Dialect
	lang         Language
	err          error
	queries_seen map[string]bool
}

func newCodeRenderer(w io.Writer, dialect Dialect, lang Language) (
	r *codeRenderer) {

	return &codeRenderer{
		w:            w,
		dialect:      dialect,
		lang:         lang,
		queries_seen: map[string]bool{},
	}
}

func (r *codeRenderer) RenderCode(schema *Schema) (err error) {
	r.err = nil

	r.setError(r.lang.RenderHeader(r.w, schema))

	for _, table := range schema.Tables {
		r.renderTable(table)
	}

	for _, query := range schema.Queries {
		r.renderQuery(query)
	}

	return r.err
}

func equalsConditionsForColumns(columns []*Column) []*ConditionParams {
	conditions := make([]*ConditionParams, 0, len(columns))
	for _, column := range columns {
		conditions = append(conditions, ColumnEquals(column))
	}
	return conditions
}

func (r *codeRenderer) renderTable(table *Table) {
	r.renderInsert(&InsertParams{
		Table:   table,
		Columns: table.InsertableColumns(),
	})
	if len(table.UpdatableColumns()) > 0 {
		for _, columns := range table.UpdatableBy() {
			r.renderUpdate(&UpdateParams{
				Table:      table,
				Conditions: equalsConditionsForColumns(columns),
			})
		}
	}
	r.renderQuery(&Query{
		Table: table,
		Start: table.PrimaryKey,
	})
}

func selectFromQuery(query *Query) *SelectParams {
	start := equalsConditionsForColumns(query.Start)
	many := !query.Table.ColumnSetUnique(query.Start)
	end := equalsConditionsForColumns(query.End)
	joins := Joins(query.Joins...)

	var conditions []*ConditionParams
	conditions = append(conditions, start...)
	conditions = append(conditions, end...)

	return &SelectParams{
		Table:      query.Table,
		Many:       many,
		LeftJoins:  joins,
		Conditions: conditions,
	}
}

func selectsFromQuery(query *Query) (out []*SelectParams) {
	out = append(out, selectFromQuery(query))

	if out[0].Many && len(query.Table.PrimaryKey) == 1 {
		paged_select := *out[0]
		paged_select.PagedOn = query.Table.PrimaryKey[0]
		out = append(out, &paged_select)
	}

	return out
}

func deletesFromQuery(query *Query) (out []*DeleteParams) {
	params := &DeleteParams{
		Table: query.Table,
	}

	many := !query.Table.ColumnSetUnique(query.Start)
	if !many {
		params.Conditions = equalsConditionsForColumns(query.Start)
	}

	if len(query.Joins) > 0 {
		left := query.Joins[0].Left
		in := &SelectParams{
			Table:      query.Joins[0].Right.Table,
			What:       What(query.Joins[0].Right),
			Many:       len(query.Joins) > 1 || !query.Joins[0].Right.Table.ColumnSetUnique(query.End),
			LeftJoins:  Joins(query.Joins[1:]...),
			Conditions: equalsConditionsForColumns(query.End),
		}
		params.Conditions = append(params.Conditions, ColumnIn(left, in))
	}

	return append(out, params)
}

func (r *codeRenderer) querySeen(query *Query) bool {
	key := query.String()
	if r.queries_seen[key] {
		return true
	}
	r.queries_seen[key] = true
	return false
}

func (r *codeRenderer) renderQuery(query *Query) {
	if r.querySeen(query) {
		return
	}

	for _, params := range selectsFromQuery(query) {
		r.renderSelect(params)
	}
	for _, params := range deletesFromQuery(query) {
		r.renderDelete(params)
	}
}

func (r *codeRenderer) renderInsert(params *InsertParams) {
	if r.err != nil {
		return
	}
	sql, err := r.dialect.RenderInsert(params)
	if !r.setError(err) {
		return
	}
	r.setError(r.lang.RenderInsert(r.w, sql, params))
}

func (r *codeRenderer) renderUpdate(params *UpdateParams) {
	if r.err != nil {
		return
	}
	sql, err := r.dialect.RenderUpdate(params)
	if !r.setError(err) {
		return
	}
	r.setError(r.lang.RenderUpdate(r.w, sql, params))
}

func (r *codeRenderer) renderSelect(params *SelectParams) {
	if r.err != nil {
		return
	}
	sql, err := r.dialect.RenderSelect(params)
	if !r.setError(err) {
		return
	}
	if !r.setError(r.lang.RenderSelect(r.w, sql, params)) {
		return
	}

	if params.PagedOn == nil {
		sql, err = r.dialect.RenderCount(params)
		if !r.setError(err) {
			return
		}
		if !r.setError(r.lang.RenderCount(r.w, sql, params)) {
			return
		}
	}
}

func (r *codeRenderer) renderDelete(params *DeleteParams) {
	if r.err != nil {
		return
	}
	sql, err := r.dialect.RenderDelete(params)
	if !r.setError(err) {
		return
	}
	r.setError(r.lang.RenderDelete(r.w, sql, params))
}

func (r *codeRenderer) render(code string, err error) {
	if !r.setError(err) {
		return
	}
	r.printf("%s\n", code)
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
