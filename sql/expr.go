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

	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

func ExprSQL(expr *ir.Expr, dialect Dialect) sqlgen.SQL {
	switch {
	case expr.Null:
		return L("NULL")
	case expr.StringLit != nil:
		return J("", L("'"), L(dialect.EscapeString(*expr.StringLit)), L("'"))
	case expr.NumberLit != nil:
		return L(*expr.NumberLit)
	case expr.BoolLit != nil:
		return L(dialect.BoolLit(*expr.BoolLit))
	case expr.Placeholder:
		return L("?")
	case expr.Field != nil:
		return L(expr.Field.ColumnRef())
	case expr.FuncCall != nil:
		var args []sqlgen.SQL
		for _, arg := range expr.FuncCall.Args {
			args = append(args, ExprSQL(arg, dialect))
		}
		return J("", L(expr.FuncCall.Name), L("("), J(", ", args...), L(")"))
	default:
		panic(fmt.Sprintf("unhandled expression variant: %+v", expr))
	}
}
