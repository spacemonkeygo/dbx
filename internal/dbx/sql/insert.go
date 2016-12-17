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

package sql

import "gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/ast"

var insertTmpl = `INSERT INTO {{ .Table }}(
	{{- range $i, $col := .Columns }}
		{{- if $i }}, {{ end }}{{ $col }}
	{{- end }}) VALUES(
	{{- range $i, $col := .Columns }}
		{{- if $i }}, {{ end }}?
	{{- end }}){{ if .SupportsReturning }} RETURNING *{{ end }}`

func RenderInsert(dialect Dialect, model *ast.Model) string {
	return mustRender(insertTmpl, SQLInsertFromModel(model, dialect))
}

type SQLInsert struct {
	model   *ast.Model
	dialect Dialect
}

func SQLInsertFromModel(model *ast.Model, dialect Dialect) *SQLInsert {
	return &SQLInsert{
		model:   model,
		dialect: dialect,
	}
}

func (i *SQLInsert) Table() string {
	return i.model.Table
}

func (i *SQLInsert) Columns() (cols []string) {
	for _, field := range i.model.InsertableFields() {
		cols = append(cols, field.Column)
	}
	return cols
}

func (i *SQLInsert) SupportsReturning() bool {
	return i.dialect.SupportsReturning()
}
