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

package xform

import (
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func transformWheres(lookup *lookup, models map[string]scanner.Position,
	ast_wheres []*ast.Where) (wheres []*ir.Where, err error) {
	for _, ast_where := range ast_wheres {
		where, err := transformWhere(lookup, models, ast_where)
		if err != nil {
			return nil, err
		}

		wheres = append(wheres, where)
	}
	return wheres, nil
}

func transformWhere(lookup *lookup, models map[string]scanner.Position,
	ast_where *ast.Where) (where *ir.Where, err error) {

	lexpr, err := transformExpr(lookup, models, ast_where.Left, true)
	if err != nil {
		return nil, err
	}

	rexpr, err := transformExpr(lookup, models, ast_where.Right, false)
	if err != nil {
		return nil, err
	}

	return &ir.Where{
		Left:  lexpr,
		Op:    ast_where.Op.Value,
		Right: rexpr,
	}, nil
}
