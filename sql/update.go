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

func UpdateSQL(ir_upd *ir.Update, dialect Dialect) (prefix, suffix sqlgen.SQL) {
	return SQLFromUpdate(UpdateFromIRUpdate(ir_upd, dialect))
}

type Update struct {
	Table     string
	Where     []Where
	Returning []string
	In        sqlgen.SQL
}

func UpdateFromIRUpdate(ir_upd *ir.Update, dialect Dialect) *Update {
	var returning []string
	if dialect.Features().Returning {
		returning = ir_upd.Model.SelectRefs()
	}

	if len(ir_upd.Joins) == 0 {
		return &Update{
			Table:     ir_upd.Model.Table,
			Where:     WheresFromIRWheres(ir_upd.Where),
			Returning: returning,
		}
	}

	pk_column := ir_upd.Model.PrimaryKey[0].Column
	sel := SQLFromSelect(&Select{
		From:   ir_upd.Model.Table,
		Fields: []string{pk_column},
		Joins:  JoinsFromIRJoins(ir_upd.Joins),
		Where:  WheresFromIRWheres(ir_upd.Where),
	})
	in := J("", L(pk_column), L(" IN ("), sel, L(")"))

	return &Update{
		Table:     ir_upd.Model.Table,
		Returning: returning,
		In:        in,
	}
}

func SQLFromUpdate(upd *Update) (prefix, suffix sqlgen.SQL) {
	// TODO(jeff): holes instead of prefix and suffix.

	{ // build prefix
		prefix = Lf("UPDATE %s SET", upd.Table)
	}

	{ // build suffix
		stmt := Build(nil)

		wheres := SQLFromWheres(upd.Where)
		if upd.In != nil {
			wheres = append(wheres, upd.In)
		}
		if len(wheres) > 0 {
			stmt.Add(L("WHERE"), J(" AND ", wheres...))
		}

		if len(upd.Returning) > 0 {
			stmt.Add(L("RETURNING"), J(", ", Strings(upd.Returning)...))
		}

		suffix = stmt.SQL()
	}

	return sqlcompile.Compile(prefix), sqlcompile.Compile(suffix)
}
