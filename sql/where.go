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
	"fmt"
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

type Where struct {
	Left     string
	Op       string
	Right    string
	Nullable bool
}

func WhereFromIRWhere(ir_where *ir.Where) Where {
	where := Where{
		Left: ir_where.Left.ColumnRef(),
		Op:   strings.ToUpper(string(ir_where.Op)),
	}
	if ir_where.Right != nil {
		where.Right = ir_where.Right.ColumnRef()
	} else {
		where.Right = "?"
	}
	if (where.Op == "=" || where.Op == "!=") && ir_where.Left.Nullable {
		where.Nullable = true
	}
	return where
}

func WheresFromIRWheres(ir_wheres []*ir.Where) (wheres []Where) {
	wheres = make([]Where, 0, len(ir_wheres))
	for _, ir_where := range ir_wheres {
		wheres = append(wheres, WhereFromIRWhere(ir_where))
	}
	return wheres
}

func SQLFromWheres(wheres []Where) (out []sqlgen.SQL) {
	conditions := 0
	for _, where := range wheres {
		var where_sql sqlgen.SQL
		if where.Nullable {
			where_sql = &sqlgen.Condition{
				Name:  fmt.Sprintf("cond_%d", conditions),
				Field: where.Left,
				Equal: where.Op == "=",
			}
			conditions++
		} else {
			where_sql = J(" ", L(where.Left), L(where.Op), L(where.Right))
		}
		out = append(out, where_sql)
	}
	return out
}
