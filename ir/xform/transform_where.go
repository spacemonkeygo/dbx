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
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
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

	where = new(ir.Where)
	switch {
	case ast_where.Clause != nil:
		where.Clause, err = transformWhereClause(lookup, models,
			ast_where.Clause)
	case ast_where.Or != nil:
		where.Or, err = transformWhereOr(lookup, models, ast_where.Or)
	case ast_where.And != nil:
		where.And, err = transformWhereAnd(lookup, models, ast_where.And)
	default:
		err = errutil.New(ast_read.Pos, "no fields defined to select")
	}
	if err != nil {
		return nil, err
	}
	return where, nil
}

func transformWhereClause(lookup *lookup, models map[string]scanner.Position,
	ast_where_clause *ast.Where) (where *ir.WhereClause, err error) {

	lexpr, err := transformExpr(lookup, models, ast_where_clause.Left, true)
	if err != nil {
		return nil, err
	}

	rexpr, err := transformExpr(lookup, models, ast_where_clause.Right, false)
	if err != nil {
		return nil, err
	}

	return &ir.Where{
		Left:  lexpr,
		Op:    ast_where_clause.Op.Value,
		Right: rexpr,
	}, nil
}

func transformWhereOr(lookup *lookup, models map[string]scanner.Position,
	ast_where_or *ast.WhereOr) (where *ir.WhereOr, err error) {

	lwhere, err := transformWhere(lookup, models, ast_where_or.Left)
	if err != nil {
		return nil, err
	}

	rwhere, err := transformWhere(lookup, models, ast_where_or.Right)
	if err != nil {
		return nil, err
	}

	return &ir.WhereOr{
		Left:  lwhere,
		Right: rwhere,
	}, nil
}

func transformWhereAnd(lookup *lookup, models map[string]scanner.Position,
	ast_where_or *ast.WhereAnd) (where *ir.WhereAnd, err error) {

	lwhere, err := transformWhere(lookup, models, ast_where_or.Left)
	if err != nil {
		return nil, err
	}

	rwhere, err := transformWhere(lookup, models, ast_where_or.Right)
	if err != nil {
		return nil, err
	}

	return &ir.WhereAnd{
		Left:  lwhere,
		Right: rwhere,
	}, nil
}
