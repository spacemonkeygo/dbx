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

	models, joins, err := transformJoins(lookup, ast_del.Joins)
	if err != nil {
		return nil, err
	}
	models[model.Name] = ast_del.Model.Pos

	del.Joins = joins

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
