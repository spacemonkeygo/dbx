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

func DeleteSQL(ir_del *ir.Delete) sqlgen.SQL {
	return SQLFromDelete(DeleteFromIRDelete(ir_del))
}

type Delete struct {
	From  string
	Where []Where
	In    sqlgen.SQL
}

func DeleteFromIRDelete(ir_del *ir.Delete) *Delete {
	if len(ir_del.Joins) == 0 {
		return &Delete{
			From:  ir_del.Model.Table,
			Where: WheresFromIRWheres(ir_del.Where),
		}
	}

	pk_column := ir_del.Model.PrimaryKey[0].ColumnRef()
	sel := SQLFromSelect(&Select{
		From:   ir_del.Model.Table,
		Fields: []string{pk_column},
		Joins:  JoinsFromIRJoins(ir_del.Joins),
		Where:  WheresFromIRWheres(ir_del.Where),
	})
	in := J("", L(pk_column), L(" IN ("), sel, L(")"))

	return &Delete{
		From: ir_del.Model.Table,
		In:   in,
	}
}

func SQLFromDelete(del *Delete) sqlgen.SQL {
	stmt := Build(Lf("DELETE FROM %s", del.From))

	wheres := SQLFromWheres(del.Where)
	if del.In != nil {
		wheres = append(wheres, del.In)
	}
	if len(wheres) > 0 {
		stmt.Add(L("WHERE"), J(" AND ", wheres...))
	}

	return sqlcompile.Compile(stmt.SQL())
}
