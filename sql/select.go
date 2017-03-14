// Copyright (C) 2017 Space Monkey, Inc.
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
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
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
	sel := SelectFromIR(ir_read, dialect)
	sql := SQLFromSelect(sel)
	return sqlgen.Render(dialect, sql)
}

func RenderGetLast(dialect Dialect, ir_model *ir.Model) string {
	sql := SQLFromSelect(&Select{
		Fields: ir_model.SelectRefs(),
		From:   ir_model.Table,
		Where: []Where{
			{
				Left:  dialect.RowId(),
				Op:    "=",
				Right: "?",
			},
		},
	})
	return sqlgen.Render(dialect, sql)
}

func SQLFromSelect(sel *Select) sqlgen.SQL {
	stmt := Build(nil)

	if sel.Has {
		stmt.Add(L("SELECT COALESCE(("))
	}

	stmt.Add(
		L("SELECT"),
		J(", ", Strings(sel.Fields)...),
		Lf("FROM %s", sel.From),
	)

	for _, join := range sel.Joins {
		j := Build(Lf("%s JOIN %s ON %s =", join.Type, join.Table, join.Left))
		if join.Right != "" {
			j.Add(L(join.Right))
		} else {
			j.Add(Placeholder)
		}
		stmt.Add(j.SQL())
	}

	if wheres := SQLFromWheres(sel.Where); len(wheres) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", wheres...))
	}

	if ob := sel.OrderBy; ob != nil {
		stmt.Add(L("ORDER BY"), J(", ", Strings(ob.Fields)...))
		if ob.Descending {
			stmt.Add(L("DESC"))
		}
	}

	if sel.Limit != "" {
		stmt.Add(Lf("LIMIT %s", sel.Limit))
	}

	if sel.Offset != "" {
		stmt.Add(Lf("OFFSET %s", sel.Offset))
	}

	if sel.Has {
		stmt.Add(L("), 0)"))
	}

	return sqlgen.Compile(stmt.SQL())
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

func SelectFromIR(ir_read *ir.Read, dialect Dialect) *Select {
	sel := &Select{
		From:  ir_read.From.Table,
		Where: WheresFromIR(ir_read.Where),
		Joins: JoinsFromIR(ir_read.Joins),
	}

	for _, ir_selectable := range ir_read.Selectables {
		sel.Fields = append(sel.Fields, ir_selectable.SelectRefs()...)
	}

	if ir_read.OrderBy != nil {
		sel.OrderBy = OrderByFromIR(ir_read.OrderBy)
	}

	switch ir_read.View {
	case ir.All:
	case ir.One, ir.Scalar:
		if !ir_read.Distinct() {
			sel.Limit = "2"
		}
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
	case ir.First:
		sel.Limit = "1"
		sel.Offset = "0"
	default:
		panic(fmt.Sprintf("unsupported select view %s", ir_read.View))
	}

	return sel
}
