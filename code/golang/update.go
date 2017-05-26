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
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlembedgo"
)

type Update struct {
	PartitionedArgs
	Info              sqlembedgo.Info
	InfoGet           sqlembedgo.Info
	Suffix            string
	Struct            *ModelStruct
	Return            *Var
	AutoFields        []*Var
	SupportsReturning bool
	NeedsNow          bool
}

func UpdateFromIR(ir_upd *ir.Update, dialect sql.Dialect) *Update {
	update_sql := sql.UpdateSQL(ir_upd, dialect)
	upd := &Update{
		PartitionedArgs:   PartitionedArgsFromWheres(ir_upd.Where),
		Info:              sqlembedgo.Embed("__", update_sql),
		Suffix:            convertSuffix(ir_upd.Suffix),
		Struct:            ModelStructFromIR(ir_upd.Model),
		SupportsReturning: dialect.Features().Returning,
	}
	if !ir_upd.NoReturn {
		upd.Return = VarFromModel(ir_upd.Model)
	}

	for _, field := range ir_upd.AutoUpdatableFields() {
		upd.NeedsNow = upd.NeedsNow || field.IsTime()
		upd.AutoFields = append(upd.AutoFields, VarFromField(field))
	}

	if !upd.SupportsReturning {
		select_sql := sql.SelectSQL(&ir.Read{
			From:        ir_upd.Model,
			Selectables: []ir.Selectable{ir_upd.Model},
			Joins:       ir_upd.Joins,
			Where:       ir_upd.Where,
			View:        ir.All,
		}, dialect)
		upd.InfoGet = sqlembedgo.Embed("__", select_sql)
	}

	return upd
}
