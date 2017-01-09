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

package golang

import (
	"fmt"

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

type Select struct {
	Suffix  string
	Row     *Var
	Args    []*Var
	SQL     string
	PagedOn string
}

func SelectFromIR(ir_sel *ir.Select, dialect sql.Dialect) *Select {
	sel := &Select{
		Suffix: inflect.Camelize(ir_sel.FuncSuffix),
		SQL:    sql.RenderSelect(dialect, ir_sel),
	}

	for _, where := range ir_sel.Where {
		if where.Right == nil {
			sel.Args = append(sel.Args, VarFromField(where.Left))
		}
	}

	vars := VarsFromSelectables(ir_sel.Fields)
	if len(vars) == 1 {
		sel.Row = vars[0]
	} else {
		sel.Row = StructVar("row", resultStructName(ir_sel.FuncSuffix), vars)
	}

	switch ir_sel.View {
	case ir.All:
	case ir.Limit:
		sel.Args = append(sel.Args, &Var{
			Name: "limit",
			Type: "int",
		})
	case ir.Offset:
		sel.Args = append(sel.Args, &Var{
			Name: "offset",
			Type: "int64",
		})
	case ir.LimitOffset:
		sel.Args = append(sel.Args, &Var{
			Name: "limit",
			Type: "int",
		})
		sel.Args = append(sel.Args, &Var{
			Name: "offset",
			Type: "int64",
		})
	case ir.Paged:
		sel.Args = append(sel.Args, &Var{
			Name: "ctoken",
			Type: "string",
		})
		sel.Args = append(sel.Args, &Var{
			Name: "limit",
			Type: "int",
		})
		paged_on := ModelFieldFromIR(ir_sel.From.BasicPrimaryKey()).Name
		if len(ir_sel.Fields) >= 2 {
			field := FieldFromSelectable(ir_sel.From)
			sel.PagedOn = field.Name + "." + paged_on
		} else {
			sel.PagedOn = paged_on
		}
	default:
		panic(fmt.Sprintf("unhandled view type %s", ir_sel.View))
	}

	return sel
}

func ResultStructFromSelect(ir_sel *ir.Select) *Struct {
	return &Struct{
		Name:   resultStructName(ir_sel.FuncSuffix),
		Fields: FieldsFromSelectables(ir_sel.Fields),
	}
}

func resultStructName(suffix string) string {
	return fmt.Sprintf("%sRow", inflect.Camelize(suffix))
}
