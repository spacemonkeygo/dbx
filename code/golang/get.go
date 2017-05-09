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

package golang

import (
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlembedgo"
)

type Get struct {
	PartitionedArgs
	Info   sqlembedgo.Info
	Suffix string
	Row    *Var
	LastPk *Var
}

func GetFromIR(ir_read *ir.Read, dialect sql.Dialect) *Get {
	select_sql := sql.SelectSQL(ir_read, dialect)
	get := &Get{
		PartitionedArgs: PartitionedArgsFromWheres(ir_read.Where),
		Info:            sqlembedgo.Embed("__", select_sql),
		Suffix:          convertSuffix(ir_read.Suffix),
	}

	get.Row = GetRowFromIR(ir_read)

	if ir_read.View == ir.Paged {
		pk_var := VarFromField(ir_read.From.BasicPrimaryKey())
		pk_var.Name = "__" + pk_var.Name
		get.LastPk = pk_var
	}

	return get
}

func GetRowFromIR(ir_read *ir.Read) *Var {
	if model := ir_read.SelectedModel(); model != nil {
		return VarFromModel(model)
	}

	return MakeResultVar(ir_read.Selectables)
}

func MakeResultVar(selectables []ir.Selectable) *Var {
	vars := VarsFromSelectables(selectables)

	// construct the aggregate struct name
	var parts []string
	for _, v := range vars {
		parts = append(parts, v.Name)
	}
	parts = append(parts, "Row")
	name := strings.Join(parts, "_")
	return StructVar("row", name, vars)
}

func ResultStructFromRead(ir_read *ir.Read) *Struct {
	// no result struct if there is just a single model selected
	if ir_read.SelectedModel() != nil {
		return nil
	}

	result := MakeResultVar(ir_read.Selectables)

	s := &Struct{
		Name: result.Type,
	}

	for _, field := range result.Fields {
		s.Fields = append(s.Fields, Field{
			Name: field.Name,
			Type: field.Type,
		})
	}

	return s
}
