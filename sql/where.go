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
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlcompile"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

type Where struct {
	Left  string
	Op    string
	Right string
}

func SQLFromWhere(where Where) sqlgen.SQL {
	return sqlcompile.Compile(
		J(" ", L(where.Left), L(where.Op), L(where.Right)))
}

func SQLFromWheres(wheres []Where) []sqlgen.SQL {
	var out []sqlgen.SQL
	for _, where := range wheres {
		out = append(out, SQLFromWhere(where))
	}
	return out
}

func WhereFromIR(ir_where *ir.Where) Where {
	where := Where{
		Left: ir_where.Left.ColumnRef(),
		Op:   strings.ToUpper(string(ir_where.Op)),
	}
	if ir_where.Right != nil {
		where.Right = ir_where.Right.ColumnRef()
	} else {
		where.Right = "?"
	}
	return where
}

func WheresFromIR(ir_wheres []*ir.Where) (wheres []Where) {
	wheres = make([]Where, 0, len(ir_wheres))
	for _, ir_where := range ir_wheres {
		wheres = append(wheres, WhereFromIR(ir_where))
	}
	return wheres
}
