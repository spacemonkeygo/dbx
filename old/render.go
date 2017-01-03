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

func RenderCode(w io.Writer, schema *Schema, dialects []Dialect, lang Language) (
	err error) {

	renderer := newCodeRenderer(w, dialects, lang)
	return renderer.RenderCode(schema)
}

type codeRenderer struct {
	w            io.Writer
	dialects     []Dialect
	lang         Language
	err          error
	queries_seen map[string]bool
}

func newCodeRenderer(w io.Writer, dialects []Dialect, lang Language) (
	r *codeRenderer) {

	return &codeRenderer{
		w:            w,
		dialects:     dialects,
		lang:         lang,
		queries_seen: map[string]bool{},
	}
}

func (r *codeRenderer) RenderCode(schema *Schema) (err error) {
	r.err = nil

	r.setError(r.lang.RenderHeader(r.w, r.dialects, schema))

	for _, table := range schema.Tables {
		r.renderTable(table)
	}

	for _, query := range schema.Queries {
		r.renderQuery(query)
	}

	r.setError(r.lang.RenderFooter(r.w))

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
	r.renderQuery(&Query{
		Table: table,
	})
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
		Table:      query.Table,
		Many:       !query.Table.ColumnSetUnique(query.Start),
		Conditions: equalsConditionsForColumns(query.Start),
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

func updatesFromQuery(query *Query) (out []*UpdateParams) {
	if !query.Table.Updatable() ||
		!query.Table.ColumnSetUnique(query.Start) ||
		len(query.Joins) > 0 ||
		len(query.End) > 0 {
		return nil
	}
	out = append(out, &UpdateParams{
		Table:      query.Table,
		Conditions: equalsConditionsForColumns(query.Start),
	})
	return out
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

	for _, params := range updatesFromQuery(query) {
		r.renderUpdate(params)
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
	r.setError(r.lang.RenderInsert(r.w, r.dialects, params))
}

func (r *codeRenderer) renderUpdate(params *UpdateParams) {
	if r.err != nil {
		return
	}
	r.setError(r.lang.RenderUpdate(r.w, r.dialects, params))
}

func (r *codeRenderer) renderSelect(params *SelectParams) {
	if r.err != nil {
		return
	}
	if !r.setError(r.lang.RenderSelect(r.w, r.dialects, params)) {
		return
	}

	if params.PagedOn == nil {
		if !r.setError(r.lang.RenderCount(r.w, r.dialects, params)) {
			return
		}
	}
}

func (r *codeRenderer) renderDelete(params *DeleteParams) {
	if r.err != nil {
		return
	}
	r.setError(r.lang.RenderDelete(r.w, r.dialects, params))
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
