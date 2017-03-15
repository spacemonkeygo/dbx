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
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlcompile"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

func RenderDelete(dialect Dialect, ir_del *ir.Delete) string {
	del := DeleteFromIR(ir_del, dialect)
	sql := SQLFromDelete(del)
	return sqlgen.Render(dialect, sql)
}

func SQLFromDelete(del *Delete) sqlgen.SQL {
	stmt := Build(Lf("DELETE FROM %s", del.From))

	if wheres := SQLFromWheres(del.Where); len(wheres) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", wheres...))
	}

	return sqlcompile.Compile(stmt.SQL())
}

type Delete struct {
	From  string
	Where []Where
}

func DeleteFromIR(ir_del *ir.Delete, dialect Dialect) *Delete {
	if len(ir_del.Joins) == 0 {
		return &Delete{
			From:  ir_del.Model.Table,
			Where: WheresFromIR(ir_del.Where),
		}
	}

	pk_column := ir_del.Model.PrimaryKey[0].ColumnRef()

	// TODO(jeff): we should have the where optionally have a SQL for the right
	// maybe, or just make it SQL always that we stuff a literal in, but the
	// wrong thing is rendering here.

	sel := sqlgen.Render(dialect, SQLFromSelect(&Select{
		From:   ir_del.Model.Table,
		Fields: []string{pk_column},
		Joins:  JoinsFromIR(ir_del.Joins),
		Where:  WheresFromIR(ir_del.Where),
	}), sqlgen.NoTerminate)

	return &Delete{
		From: ir_del.Model.Table,
		Where: []Where{{
			Left:  pk_column,
			Op:    "IN",
			Right: "(" + sel + ")",
		}},
	}
}
