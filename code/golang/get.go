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
	"strings"

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
		Suffix: inflect.Camelize(ir_read.Suffix),
		SQL:    sql.RenderSelect(dialect, ir_read),
	}

	for _, where := range ir_read.Where {
		if where.Right == nil {
			get.Args = append(get.Args, ArgFromField(where.Left))
		}
	}

	if len(ir_read.Selectables) == 1 {
		vars := VarsFromSelectables(ir_read.Selectables)
		get.Row = vars[0]
	} else {
		result := ResultStructFromRead(ir_read)
		get.Row = StructVar("row", result.Name, result.FieldVars())
	}

	if ir_read.View == ir.Paged {
		paged_on := ModelFieldFromIR(ir_read.From.BasicPrimaryKey()).Name
		if len(ir_read.Selectables) >= 2 {
			field := FieldFromSelectable(ir_read.From)
			get.PagedOn = field.Name + "." + paged_on
		} else {
			get.PagedOn = paged_on
		}
	}

	return get
}

func ResultStructFromRead(ir_read *ir.Read) *Struct {
	fields := FieldsFromSelectables(ir_read.Selectables)

	var parts []string
	for _, field := range fields {
		parts = append(parts, inflect.Camelize(field.Name))
	}
	parts = append(parts, "Row")

	return &Struct{
		Name:   strings.Join(parts, "_"),
		Fields: fields,
	}
}
