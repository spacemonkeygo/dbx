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

package syntax

import (
	"fmt"
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestParseWhere(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	// helpers to construct ast values
	eq := &ast.Operator{Value: consts.EQ}
	number := func(n int) *ast.Expr {
		return &ast.Expr{NumberLit: &ast.String{Value: fmt.Sprint(n)}}
	}
	clause := func(l *ast.Expr, op *ast.Operator, r *ast.Expr) *ast.Where {
		return &ast.Where{Clause: &ast.WhereClause{Left: l, Op: op, Right: r}}
	}
	and := func(wheres ...*ast.Where) *ast.Where {
		return &ast.Where{And: wheres}
	}
	or := func(wheres ...*ast.Where) *ast.Where {
		return &ast.Where{Or: wheres}
	}
	field := func(model, field string) *ast.Expr {
		return &ast.Expr{FieldRef: &ast.FieldRef{
			Model: &ast.String{Value: model},
			Field: &ast.String{Value: field},
		}}
	}

	tw.Run("Basic", func(tw *testutil.T) {
		tw.Parallel()

		assertWhere(tw, "foo.bar = 3",
			clause(field("foo", "bar"), eq, number(3)),
		)
	})

	tw.Run("Compound", func(tw *testutil.T) {
		tw.Parallel()

		assertWhere(tw, "1 = 1 and (barf.baz = 2 or 3 = 3)",
			and(clause(number(1), eq, number(1)),
				or(clause(field("barf", "baz"), eq, number(2)),
					clause(number(3), eq, number(3)),
				),
			),
		)

		assertWhere(tw, "1 = 1 and 2 = 2 and 3 = 3",
			and(clause(number(1), eq, number(1)),
				clause(number(2), eq, number(2)),
				clause(number(3), eq, number(3)),
			),
		)

		assertWhere(tw, "1 = 1 or 2 = 2 or 3 = 3",
			or(clause(number(1), eq, number(1)),
				clause(number(2), eq, number(2)),
				clause(number(3), eq, number(3)),
			),
		)
	})

	tw.Run("CompoundGrouping", func(tw *testutil.T) {
		tw.Parallel()

		scanner, err := NewScanner("", []byte("1 = 1 and 2 = 2 or 3 = 3"))
		tw.AssertNoError(err)
		node, err := newTupleNode(scanner)
		tw.AssertNoError(err)
		_, err = parseWhere(node)
		tw.AssertError(err, "distinct compound statements must be grouped")
	})
}
