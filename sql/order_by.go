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

type OrderBy struct {
	Fields     []string
	Descending bool
}

func OrderByFromIROrderBy(ir_order_by *ir.OrderBy) (order_by *OrderBy) {
	order_by = &OrderBy{
		Descending: ir_order_by.Descending,
	}
	for _, ir_field := range ir_order_by.Fields {
		order_by.Fields = append(order_by.Fields, ir_field.ColumnRef())
	}
	return order_by
}

func SQLFromOrderBy(order_by *OrderBy) sqlgen.SQL {
	stmt := Build(L("ORDER BY"))
	stmt.Add(J(", ", Strings(order_by.Fields)...))
	if order_by.Descending {
		stmt.Add(L("DESC"))
	}
	return sqlcompile.Compile(stmt.SQL())
}
