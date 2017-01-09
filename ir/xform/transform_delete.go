// Copyright (C) 2016 Space Monkey, Inc.
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
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func transformDelete(lookup *lookup, ast_del *ast.Delete) (
	del *ir.Delete, err error) {

	model, err := lookup.FindModel(ast_del.Model)
	if err != nil {
		return nil, err
	}

	if len(model.PrimaryKey) > 1 && len(ast_del.Joins) > 0 {
		return nil, Error.New(
			"%s: delete with joins unsupported on multicolumn primary key",
			ast_del.Pos)
	}

	del = &ir.Delete{
		Model: model,
	}

	// Figure out set of models that are included in the delete.
	// These come from explicit joins.
	models := map[string]*ast.ModelRef{
		model.Name: ast_del.Model,
	}

	next := model.Name
	for _, join := range ast_del.Joins {
		left, err := lookup.FindField(join.Left)
		if err != nil {
			return nil, err
		}
		if join.Left.Model != next {
			return nil, Error.New(
				"%s: model order must be consistent; expected %q; got %q",
				join.Left.Pos, next, join.Left.Model)
		}
		right, err := lookup.FindField(join.Right)
		if err != nil {
			return nil, err
		}
		next = join.Right.Model
		del.Joins = append(del.Joins, &ir.Join{
			Type:  join.Type,
			Left:  left,
			Right: right,
		})
		if existing := models[join.Right.Model]; existing != nil {
			return nil, Error.New("%s: model %q already joined at %s",
				join.Right.Pos, join.Right.Model, existing.Pos)
		}
		models[join.Right.Model] = join.Right.ModelRef()
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	for _, ast_where := range ast_del.Where {
		left, err := lookup.FindField(ast_where.Left)
		if err != nil {
			return nil, err
		}
		if models[ast_where.Left.Model] == nil {
			return nil, Error.New(
				"%s: invalid where condition %q; model %q is not joined",
				ast_where.Pos, ast_where, ast_where.Left.Model)
		}

		var right *ir.Field
		if ast_where.Right != nil {
			right, err = lookup.FindField(ast_where.Right)
			if err != nil {
				return nil, err
			}
			if models[ast_where.Right.Model] == nil {
				return nil, Error.New(
					"%s: invalid where condition %q; model %q is not joined",
					ast_where.Pos, ast_where, ast_where.Right.Model)
			}
		}

		del.Where = append(del.Where, &ir.Where{
			Op:    ast_where.Op,
			Left:  left,
			Right: right,
		})
	}

	return del, nil
}
