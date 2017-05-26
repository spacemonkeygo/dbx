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
		Model:    model,
		NoReturn: ast_upd.NoReturn.Get(),
		Suffix:   transformSuffix(ast_upd.Suffix),
	}

	models, joins, err := transformJoins(lookup, ast_upd.Joins)
	if err != nil {
		return nil, err
	}
	models[model.Name] = ast_upd.Model.Pos

	upd.Joins = joins

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	upd.Where, err = transformWheres(lookup, models, ast_upd.Where)
	if err != nil {
		return nil, err
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
