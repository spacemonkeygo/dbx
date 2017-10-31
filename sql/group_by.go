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

type GroupBy struct {
	Fields []string
}

func GroupByFromIRGroupBy(ir_group_by *ir.GroupBy) (group_by *GroupBy) {
	group_by = &GroupBy{}
	for _, ir_field := range ir_group_by.Fields {
		group_by.Fields = append(group_by.Fields, ir_field.ColumnRef())
	}
	return group_by
}

func SQLFromGroupBy(group_by *GroupBy) sqlgen.SQL {
	stmt := Build(L("GROUP BY"))
	stmt.Add(J(", ", Strings(group_by.Fields)...))
	return sqlcompile.Compile(stmt.SQL())
}
