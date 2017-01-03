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

import (
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

var selectTmpl = `SELECT {{ range $i, $f:= .Fields }}{{ if $i }}, {{ end }}{{ $f }}{{ end }}
	FROM {{ .From }}
	{{- range .Joins }}
	{{ .Type }} JOIN {{ .Table }} ON {{ .Left }} = {{ if .Right }}{{ .Right }}{{ else }}?{{ end }}
	{{- end -}}
	{{- if .Where }}
	WHERE {{ range $i, $w := .Where }}{{ if $i }} AND {{ end }}{{ $w.Left }} {{ $w.Op }} {{ $w.Right }}{{ end }}
	{{- end -}}
	{{- if .OrderBy }}
	ORDER BY {{ range $i, $field := .OrderBy.Fields }}{{ if $i }}, {{ end }}{{ $field }}{{ end }}{{ if .OrderBy.Descending }} DESC{{ end }}
	{{- end -}}
	{{- if .Limit }}
	LIMIT {{ .Limit }}
	{{- end }}
`

func RenderSelect(dialect Dialect, sel *ir.Select) string {
	return mustRender(selectTmpl, SelectFromAST(sel, dialect))
}

type Select struct {
	sel     *ir.Select
	dialect Dialect
}

func SelectFromAST(sel *ir.Select, dialect Dialect) *Select {
	return &Select{
		sel:     sel,
		dialect: dialect,
	}
}

func (s *Select) From() string {
	return s.sel.From.TableName()
}

func (s *Select) Fields() (fields []string) {
	for _, field := range s.sel.Fields {
		fields = append(fields, field.SelectRefs()...)
	}
	return fields
}

func (s *Select) Joins() (sqljoins []Join) {
	for _, join := range s.sel.Joins {
		sqljoin := Join{
			Table: join.Right.Model.TableName(),
			Left:  join.Left.ColumnRef(),
			Right: join.Right.ColumnRef(),
		}
		switch join.Type {
		case ast.LeftJoin:
			sqljoin.Type = "LEFT"
		default:
			panic(fmt.Sprintf("unhandled join type %d", join.Type))
		}
		sqljoins = append(sqljoins, sqljoin)
	}
	return sqljoins
}

func (s *Select) Where() (sqlwheres []Where) {
	for _, where := range s.sel.Where {
		sqlwhere := Where{
			Left: where.Left.ColumnRef(),
			Op:   string(where.Op),
		}
		if where.Right != nil {
			sqlwhere.Right = where.Right.ColumnRef()
		} else {
			sqlwhere.Right = "?"
		}

		sqlwheres = append(sqlwheres, sqlwhere)
	}
	return sqlwheres
}

func (s *Select) OrderBy() (order_by *OrderBy) {
	if s.sel.OrderBy == nil {
		return nil
	}
	order_by = &OrderBy{
		Descending: s.sel.OrderBy.Descending,
	}
	for _, field := range s.sel.OrderBy.Fields {
		order_by.Fields = append(order_by.Fields, field.ColumnRef())
	}

	return order_by
}

func (s *Select) Limit() string {
	if s.sel.Limit == nil {
		return ""
	}
	if s.sel.Limit.Amount <= 0 {
		return "?"
	}
	return fmt.Sprint(s.sel.Limit.Amount)
}

type Where struct {
	Left  string
	Op    string
	Right string
}

type OrderBy struct {
	Fields     []string
	Descending bool
}

type Join struct {
	Type  string
	Table string
	Left  string
	Right string
}
