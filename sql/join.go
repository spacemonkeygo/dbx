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

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlcompile"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

type Join struct {
	Type  string
	Table string
	Left  string
	Right string
}

func JoinFromIRJoin(ir_join *ir.Join) Join {
	join := Join{
		Table: ir_join.Right.Model.Table,
		Left:  ir_join.Left.ColumnRef(),
		Right: ir_join.Right.ColumnRef(),
	}
	switch ir_join.Type {
	case consts.InnerJoin:
	default:
		panic(fmt.Sprintf("unhandled join type %q", join.Type))
	}
	return join
}

func JoinsFromIRJoins(ir_joins []*ir.Join) (joins []Join) {
	for _, ir_join := range ir_joins {
		joins = append(joins, JoinFromIRJoin(ir_join))
	}
	return joins
}

func SQLFromJoin(join Join) sqlgen.SQL {
	clause := Build(Lf("%s JOIN %s ON %s =", join.Type, join.Table, join.Left))
	if join.Right != "" {
		clause.Add(L(join.Right))
	} else {
		clause.Add(Placeholder)
	}
	return sqlcompile.Compile(clause.SQL())
}

func SQLFromJoins(joins []Join) []sqlgen.SQL {
	var out []sqlgen.SQL
	for _, join := range joins {
		out = append(out, SQLFromJoin(join))
	}
	return out
}
