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
)

func RenderInsert(dialect Dialect, cre *ir.Create) string {
	insert := InsertFromIR(cre, dialect)

	var sql sqlgen.SQL
	sql = sqlgen.Append(sql, sqlgen.Lf("INSERT INTO %s", insert.Table))

	if len(insert.Columns) > 0 {
		var columns, values []sqlgen.SQL
		for _, col := range insert.Columns {
			columns = append(columns, sqlgen.L(col))
			values = append(values, sqlgen.Param)
		}
		sql = sqlgen.Append(sql, sqlgen.Join(", ", columns...))
		sql = sqlgen.Append(sql,
			sqlgen.L("VALUES("),
			sqlgen.Join(", ", values...),
			sqlgen.L(")"))
	} else {
		sql = sqlgen.Append(sql, sqlgen.L("DEFAULT VALUES"))
	}

	if len(insert.Returning) > 0 {
		var returning []sqlgen.SQL
		for _, col := range insert.Returning {
			returning = append(returning, sqlgen.L(col))
		}
		sql = sqlgen.Append(sql,
			sqlgen.L("RETURNING"),
			sqlgen.Join(", ", returning...))
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
