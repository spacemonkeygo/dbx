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

import "gopkg.in/spacemonkeygo/dbx.v1/ir"

const insertTmpl = `INSERT INTO {{ .Table -}}
	{{ if .Columns }}(
		{{- range $i, $col := .Columns }}
			{{- if $i }}, {{ end }}{{ $col }}
		{{- end }})
		VALUES(
		{{- range $i, $col := .Columns }}
			{{- if $i }}, {{ end }}?
		{{- end }})
	{{- else }}
		DEFAULT VALUES
	{{- end -}}
	{{ if .Returning }} RETURNING *{{ end }}`

func RenderInsert(dialect Dialect, ins *ir.Insert) string {
	return render(dialect, insertTmpl, InsertFromIR(ins, dialect))
}

type Insert struct {
	Table     string
	Columns   []string
	Returning bool
}

func InsertFromIR(ir_ins *ir.Insert, dialect Dialect) *Insert {
	ins := &Insert{
		Table:     ir_ins.Model.TableName(),
		Returning: dialect.Features().Returning,
	}
	for _, field := range ir_ins.Fields() {
		if field == ir_ins.Model.BasicPrimaryKey() && !ir_ins.Raw {
			continue
		}
		ins.Columns = append(ins.Columns, field.ColumnName())
	}
	return ins
}
