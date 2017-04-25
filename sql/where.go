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

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

func WhereSQL(wheres []*ir.Where, dialect Dialect) (out []sqlgen.SQL) {
	// we put all the condition wheres at the end for ease of template
	// generation later.

	for _, where := range wheres {
		if where.NeedsCondition() {
			continue
		}
		out = append(out,
			J(" ", ExprSQL(where.Left, dialect),
				opSQL(where.Op, where.Left, where.Right),
				ExprSQL(where.Right, dialect)))
	}

	conditions := 0
	for _, where := range wheres {
		if !where.NeedsCondition() {
			continue
		}
		out = append(out, &sqlgen.Condition{
			Name:  fmt.Sprintf("cond_%d", conditions),
			Left:  ExprSQL(where.Left, dialect).Render(),
			Equal: where.Op == "=",
			Right: ExprSQL(where.Right, dialect).Render(),
		})
		conditions++
	}

	return out
}

func opSQL(op consts.Operator, left, right *ir.Expr) sqlgen.SQL {
	switch op {
	case consts.EQ:
		if left.Null || right.Null {
			return L("is")
		}
	case consts.NE:
		if left.Null || right.Null {
			return L("is not")
		}
	}
	return L(strings.ToUpper(string(op)))
}
