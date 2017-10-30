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
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestParseWhere(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	tw.Run("Basic", func(tw *testutil.T) {
		tw.Parallel()

		assertWhere(tw, "foo.bar = 3",
			&ast.Where{Clause: &ast.WhereClause{
				Left: &ast.Expr{FieldRef: &ast.FieldRef{
					Model: &ast.String{Value: "foo"},
					Field: &ast.String{Value: "bar"},
				}},
				Op: &ast.Operator{Value: consts.EQ},
				Right: &ast.Expr{NumberLit: &ast.String{
					Value: "3",
				}},
			}})

		// TODO(jeff): add some exhaustive cases
	})

	tw.Run("Compound", func(tw *testutil.T) {
		tw.Parallel()

	})
}
