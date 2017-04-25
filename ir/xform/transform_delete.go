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

func transformDelete(lookup *lookup, ast_del *ast.Delete) (
	del *ir.Delete, err error) {

	model, err := lookup.FindModel(ast_del.Model)
	if err != nil {
		return nil, err
	}

	if len(model.PrimaryKey) > 1 && len(ast_del.Joins) > 0 {
		return nil, errutil.New(ast_del.Joins[0].Pos,
			"delete with joins unsupported on multicolumn primary key")
	}

	del = &ir.Delete{
		Model:  model,
		Suffix: transformSuffix(ast_del.Suffix),
	}

	// Figure out set of models that are included in the delete.
	// These come from explicit joins.
	models := map[string]scanner.Position{
		model.Name: ast_del.Model.Pos,
	}

	next := model.Name
	for _, join := range ast_del.Joins {
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
		del.Joins = append(del.Joins, &ir.Join{
			Type:  join.Type.Get(),
			Left:  left,
			Right: right,
		})
		if existing_pos, ok := models[join.Right.Model.Value]; ok {
			return nil, errutil.New(join.Right.Model.Pos,
				"model %q already joined at %s",
				join.Right.Model.Value, existing_pos)
		}
		models[join.Right.Model.Value] = join.Right.Pos
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	del.Where, err = transformWheres(lookup, models, ast_del.Where)
	if err != nil {
		return nil, err
	}

	if del.Suffix == nil {
		del.Suffix = DefaultDeleteSuffix(del)
	}

	return del, nil
}
