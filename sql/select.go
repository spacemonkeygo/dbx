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

const (
	selectTmpl = `SELECT {{ range $i, $f:= .Fields }}{{ if $i }}, {{ end }}{{ $f }}{{ end }}
	FROM {{ .From }}
	{{ range .Joins }}{{ .Type }} JOIN {{ .Table }} ON {{ .Left }} = {{ if .Right }}{{ .Right }}{{ else }}?{{ end }}{{- end -}}
	{{ if .Where }} WHERE {{- range $i, $w := .Where }}{{ if $i }} AND{{ end }} {{ $w.Left }} {{ $w.Op }} {{ $w.Right }}{{ end }} {{- end -}}
	{{ if .OrderBy }} ORDER BY {{- range $i, $field := .OrderBy.Fields }}{{ if $i }}, {{ end }} {{ $field }}{{ end }}{{ if .OrderBy.Descending }} DESC{{ end }} {{- end -}}
	{{ if .Limit }} LIMIT {{ .Limit }} {{- end -}}`

	hasTmpl = `SELECT COALESCE((` + selectTmpl + `), 0)`
)

var (
	countFields = []string{"COUNT(*)"}
	hasFields   = []string{"1"}
)

func RenderSelect(dialect Dialect, ir_sel *ir.Select) string {
	return render(selectTmpl, SelectFromIR(ir_sel, dialect))
}

func RenderCount(dialect Dialect, ir_sel *ir.Select) string {
	sel := SelectFromIR(ir_sel, dialect)
	sel.Fields = countFields
	return render(selectTmpl, sel)
}

func RenderHas(dialect Dialect, ir_sel *ir.Select) string {
	sel := SelectFromIR(ir_sel, dialect)
	sel.Fields = hasFields
	return render(hasTmpl, sel)
}

func RenderGetLast(dialect Dialect, ir_model *ir.Model) string {
	sel := Select{
		Fields: ir_model.SelectRefs(),
		From:   ir_model.TableName(),
		Where: []Where{
			{
				Left:  dialect.RowId(),
				Op:    "=",
				Right: "?",
			},
		},
	}
	return render(selectTmpl, sel)
}

func SelectFromIR(ir_sel *ir.Select, dialect Dialect) *Select {
	sel := &Select{
		From:  ir_sel.From.TableName(),
		Where: WheresFromIR(ir_sel.Where),
	}

	for _, ir_field := range ir_sel.Fields {
		sel.Fields = append(sel.Fields, ir_field.SelectRefs()...)
	}

	for _, ir_join := range ir_sel.Joins {
		join := Join{
			Table: ir_join.Right.Model.TableName(),
			Left:  ir_join.Left.ColumnRef(),
			Right: ir_join.Right.ColumnRef(),
		}
		switch ir_join.Type {
		case ast.LeftJoin:
			join.Type = "LEFT"
		default:
			panic(fmt.Sprintf("unhandled join type %d", join.Type))
		}
		sel.Joins = append(sel.Joins, join)
	}

	if ir_sel.OrderBy != nil {
		order_by := &OrderBy{
			Descending: ir_sel.OrderBy.Descending,
		}
		for _, ir_field := range ir_sel.OrderBy.Fields {
			order_by.Fields = append(order_by.Fields, ir_field.ColumnRef())
		}
		sel.OrderBy = order_by
	}

	if ir_sel.Limit != nil {
		if ir_sel.Limit.Amount <= 0 {
			sel.Limit = "?"
		} else {
			sel.Limit = fmt.Sprint(ir_sel.Limit.Amount)
		}
	}

	return sel
}

type Select struct {
	From    string
	Fields  []string
	Joins   []Join
	Where   []Where
	OrderBy *OrderBy
	Limit   string
}

type Where struct {
	Left  string
	Op    string
	Right string
}

func WhereFromIR(ir_where *ir.Where) Where {
	where := Where{
		Left: ir_where.Left.ColumnRef(),
		Op:   string(ir_where.Op),
	}
	if ir_where.Right != nil {
		where.Right = ir_where.Right.ColumnRef()
	} else {
		where.Right = "?"
	}
	return where
}

func WheresFromIR(ir_wheres []*ir.Where) (wheres []Where) {
	wheres = make([]Where, len(ir_wheres))
	for _, ir_where := range ir_wheres {
		wheres = append(wheres, WhereFromIR(ir_where))
	}
	return wheres
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
