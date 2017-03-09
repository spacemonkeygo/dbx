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
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func transformUpdate(lookup *lookup, ast_upd *ast.Update) (
	upd *ir.Update, err error) {

	model, err := lookup.FindModel(ast_upd.Model)
	if err != nil {
		return nil, err
	}

	if len(model.PrimaryKey) > 1 && len(ast_upd.Joins) > 0 {
		return nil, errutil.New(ast_upd.Joins[0].Pos,
			"update with joins unsupported on multicolumn primary key:")
	}

	upd = &ir.Update{
		Model:  model,
		Suffix: transformSuffix(ast_upd.Suffix),
	}

	// Figure out set of models that are included in the update.
	// These come from explicit joins.
	models := map[string]*ast.ModelRef{
		model.Name: ast_upd.Model,
	}

	next := model.Name
	for _, join := range ast_upd.Joins {
		left, err := lookup.FindField(join.Left)
		if err != nil {
			return nil, err
		}
		if join.Left.Model.Value != next {
			return nil, errutil.New(join.Left.Model.Pos,
				"model order must be consistent; expected %q; got %q",
				next, join.Left.Model.Value)
		}
		right, err := lookup.FindField(join.Right)
		if err != nil {
			return nil, err
		}
		next = join.Right.Model.Value
		upd.Joins = append(upd.Joins, &ir.Join{
			Type:  join.Type.Get(),
			Left:  left,
			Right: right,
		})
		if existing := models[join.Right.Model.Value]; existing != nil {
			return nil, errutil.New(join.Right.Model.Pos,
				"model %q already joined at %s",
				join.Right.Model.Value, existing.Pos)
		}
		models[join.Right.Model.Value] = join.Right.ModelRef()
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	for _, ast_where := range ast_upd.Where {
		left, err := lookup.FindField(ast_where.Left)
		if err != nil {
			return nil, err
		}
		if models[ast_where.Left.Model.Value] == nil {
			return nil, errutil.New(ast_where.Pos,
				"invalid where condition %q; model %q is not joined",
				ast_where, ast_where.Left.Model.Value)
		}

		var right *ir.Field
		if ast_where.Right != nil {
			right, err = lookup.FindField(ast_where.Right)
			if err != nil {
				return nil, err
			}
			if models[ast_where.Right.Model.Value] == nil {
				return nil, errutil.New(ast_where.Pos,
					"invalid where condition %q; model %q is not joined",
					ast_where, ast_where.Right.Model.Value)
			}
		}

		upd.Where = append(upd.Where, &ir.Where{
			Op:    ast_where.Op.Value,
			Left:  left,
			Right: right,
		})
	}

	if !upd.One() {
		return nil, errutil.New(ast_upd.Pos,
			"updates for more than one row are unsupported")
	}

	if upd.Suffix == nil {
		upd.Suffix = DefaultUpdateSuffix(upd)
	}

	return upd, nil
}
