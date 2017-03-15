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
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
)

type Update struct {
	Suffix              string
	Struct              *ModelStruct
	Return              *Var
	Args                []*Var
	AutoFields          []*Var
	SQLPrefix           string
	SQLSuffix           string
	SupportsReturning   bool
	PositionalArguments bool
	ArgumentPrefix      string
	NeedsNow            bool
	GetSQL              string
}

func UpdateFromIR(ir_upd *ir.Update, dialect sql.Dialect) *Update {
	prefix, suffix := sql.UpdateSQL(ir_upd, dialect)
	prefix_sql := sqlgen.Render(dialect, prefix, sqlgen.NoTerminate) + " "
	suffix_sql := " " + sqlgen.Render(dialect, suffix)

	upd := &Update{
		Suffix:              convertSuffix(ir_upd.Suffix),
		Struct:              ModelStructFromIR(ir_upd.Model),
		SQLPrefix:           prefix_sql,
		SQLSuffix:           suffix_sql,
		Return:              VarFromModel(ir_upd.Model),
		SupportsReturning:   dialect.Features().Returning,
		PositionalArguments: dialect.Features().PositionalArguments,
		ArgumentPrefix:      dialect.ArgumentPrefix(),
	}

	for _, where := range ir_upd.Where {
		if where.Right == nil {
			upd.Args = append(upd.Args, ArgFromWhere(where))
		}
	}

	for _, field := range ir_upd.AutoUpdatableFields() {
		upd.NeedsNow = upd.NeedsNow || field.IsTime()
		upd.AutoFields = append(upd.AutoFields, VarFromField(field))
	}

	if !upd.SupportsReturning {
		upd.GetSQL = sqlgen.Render(dialect, sql.SelectSQL(&ir.Read{
			From:        ir_upd.Model,
			Selectables: []ir.Selectable{ir_upd.Model},
			Joins:       ir_upd.Joins,
			Where:       ir_upd.Where,
			View:        ir.All,
		}))
	}

	return upd
}
