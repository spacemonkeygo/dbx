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
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

func RenderInsert(dialect Dialect, cre *ir.Create) string {
	insert := InsertFromIR(cre, dialect)

	var sql sqlgen.SQL
	sql = Append(sql, Lf("INSERT INTO %s", insert.Table))

	if len(insert.Columns) > 0 {
		var columns, values []sqlgen.SQL
		for _, col := range insert.Columns {
			columns = append(columns, L(col))
			values = append(values, Param)
		}
		sql = Append(sql, Ls(", ", columns...))
		sql = Append(sql, L("VALUES("), Ls(", ", values...), L(")"))
	} else {
		sql = Append(sql, L("DEFAULT VALUES"))
	}

	if len(insert.Returning) > 0 {
		var returning []sqlgen.SQL
		for _, col := range insert.Returning {
			returning = append(returning, L(col))
		}
		sql = Append(sql, L("RETURNING"), Ls(", ", returning...))
	}

	return sqlgen.Render(dialect, sql)
}

type Insert struct {
	Table     string
	Columns   []string
	Returning []string
}

func InsertFromIR(ir_cre *ir.Create, dialect Dialect) *Insert {
	ins := &Insert{
		Table: ir_cre.Model.Table,
	}
	if dialect.Features().Returning {
		ins.Returning = ir_cre.Model.SelectRefs()
	}
	for _, field := range ir_cre.Fields() {
		if field == ir_cre.Model.BasicPrimaryKey() && !ir_cre.Raw {
			continue
		}
		ins.Columns = append(ins.Columns, field.Column)
	}
	return ins
}
