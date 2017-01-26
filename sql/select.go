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

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

const (
	selectTmpl = `
	{{ if .Has }}SELECT COALESCE(({{ end }}
	SELECT {{ range $i, $f:= .Fields }}{{ if $i }}, {{ end }}{{ $f }}{{ end }}
	FROM {{ .From }}
	{{ range .Joins }}{{ .Type }} JOIN {{ .Table }} ON {{ .Left }} = {{ if .Right }}{{ .Right }}{{ else }}?{{ end }}{{- end -}}
	{{ if .Where }} WHERE {{- range $i, $w := .Where }}{{ if $i }} AND{{ end }} {{ $w.Left }} {{ $w.Op }} {{ $w.Right }}{{ end }} {{- end -}}
	{{ if .OrderBy }} ORDER BY {{- range $i, $field := .OrderBy.Fields }}{{ if $i }}, {{ end }} {{ $field }}{{ end }}{{ if .OrderBy.Descending }} DESC{{ end }} {{- end -}}
	{{ if .Limit }} LIMIT {{ .Limit }} {{- end -}}
	{{ if .Offset }} OFFSET {{ .Offset }} {{- end -}}
	{{ if .Has }}), 0){{ end }}`
)

var (
	countFields = []string{"COUNT(*)"}
	hasFields   = []string{"1"}
)

func RenderSelect(dialect Dialect, ir_read *ir.Read) string {
	return render(dialect, selectTmpl, SelectFromSelect(ir_read, dialect))
}

func RenderGetLast(dialect Dialect, ir_model *ir.Model) string {
	sel := Select{
		Fields: ir_model.SelectRefs(),
		From:   ir_model.Table,
		Where: []Where{
			{
				Left:  dialect.RowId(),
				Op:    "=",
				Right: "?",
			},
		},
	}
	return render(dialect, selectTmpl, sel)
}

type Select struct {
	From    string
	Fields  []string
	Joins   []Join
	Where   []Where
	OrderBy *OrderBy
	Limit   string
	Offset  string
	Has     bool
}

func SelectFromSelect(ir_read *ir.Read, dialect Dialect) *Select {
	sel := &Select{
		From:  ir_read.From.Table,
		Where: WheresFromIR(ir_read.Where),
		Joins: JoinsFromIR(ir_read.Joins),
	}

	for _, ir_selectable := range ir_read.Selectables {
		sel.Fields = append(sel.Fields, ir_selectable.SelectRefs()...)
	}

	if ir_read.OrderBy != nil {
		order_by := &OrderBy{
			Descending: ir_read.OrderBy.Descending,
		}
		for _, ir_field := range ir_read.OrderBy.Fields {
			order_by.Fields = append(order_by.Fields, ir_field.ColumnRef())
		}
		sel.OrderBy = order_by
	}

	switch ir_read.View {
	case ir.All:
	case ir.LimitOffset:
		sel.Limit = "?"
		sel.Offset = "?"
	case ir.Paged:
		pk := ir_read.From.BasicPrimaryKey()
		sel.Where = append(sel.Where, WhereFromIR(&ir.Where{
			Left: pk,
			Op:   consts.GT,
		}))
		sel.OrderBy = &OrderBy{
			Fields: []string{pk.ColumnRef()},
		}
		sel.Limit = "?"
		sel.Fields = append(sel.Fields, pk.SelectRefs()...)
	case ir.Has:
		sel.Has = true
		sel.Fields = hasFields
		sel.OrderBy = nil
	case ir.Count:
		sel.Fields = countFields
		sel.OrderBy = nil
	default:
		panic(fmt.Sprintf("unsupported select view %s", ir_read.View))
	}

	return sel
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
	wheres = make([]Where, 0, len(ir_wheres))
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

func JoinsFromIR(ir_joins []*ir.Join) (joins []Join) {
	for _, ir_join := range ir_joins {
		join := Join{
			Table: ir_join.Right.Model.Table,
			Left:  ir_join.Left.ColumnRef(),
			Right: ir_join.Right.ColumnRef(),
		}
		switch ir_join.Type {
		case consts.InnerJoin:
		default:
			panic(fmt.Sprintf("unhandled join type %d", join.Type))
		}
		joins = append(joins, join)
	}
	return joins
}
