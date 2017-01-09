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

type Get struct {
	Suffix  string
	Row     *Var
	Args    []*Var
	SQL     string
	PagedOn string
}

func GetFromIR(ir_read *ir.Read, dialect sql.Dialect) *Get {
	get := &Get{
		Suffix: inflect.Camelize(ir_read.FuncSuffix),
		SQL:    sql.RenderSelect(dialect, ir_read),
	}

	for _, where := range ir_read.Where {
		if where.Right == nil {
			get.Args = append(get.Args, ArgFromField(where.Left))
		}
	}

	vars := VarsFromSelectables(ir_read.Selectables)
	if len(vars) == 1 {
		get.Row = vars[0]
	} else {
		get.Row = StructVar("row", resultStructName(ir_read.FuncSuffix), vars)
	}

	switch ir_read.View {
	case ir.All, ir.Count, ir.Has:
	case ir.Limit:
		get.Args = append(get.Args, &Var{
			Name: "limit",
			Type: "int",
		})
	case ir.Offset:
		get.Args = append(get.Args, &Var{
			Name: "offset",
			Type: "int64",
		})
	case ir.LimitOffset:
		get.Args = append(get.Args, &Var{
			Name: "limit",
			Type: "int",
		})
		get.Args = append(get.Args, &Var{
			Name: "offset",
			Type: "int64",
		})
	case ir.Paged:
		get.Args = append(get.Args, &Var{
			Name: "ctoken",
			Type: "string",
		})
		get.Args = append(get.Args, &Var{
			Name: "limit",
			Type: "int",
		})
		paged_on := ModelFieldFromIR(ir_read.From.BasicPrimaryKey()).Name
		if len(ir_read.Selectables) >= 2 {
			field := FieldFromSelectable(ir_read.From)
			get.PagedOn = field.Name + "." + paged_on
		} else {
			get.PagedOn = paged_on
		}
	default:
		panic(fmt.Sprintf("unhandled view type %s", ir_read.View))
	}

	return get
}

func ResultStructFromRead(ir_read *ir.Read) *Struct {
	return &Struct{
		Name:   resultStructName(ir_read.FuncSuffix),
		Fields: FieldsFromSelectables(ir_read.Selectables),
	}
}

func resultStructName(suffix string) string {
	return fmt.Sprintf("%sRow", inflect.Camelize(suffix))
}
