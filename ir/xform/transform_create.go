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
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func transformCreate(lookup *lookup, ast_cre *ast.Create) (
	cre *ir.Create, err error) {

	model, err := lookup.FindModel(ast_cre.Model)
	if err != nil {
		return nil, err
	}

	cre = &ir.Create{
		Model:  model,
		Raw:    ast_cre.Raw.Get(),
		Suffix: transformSuffix(ast_cre.Suffix),
	}
	if cre.Suffix == nil {
		cre.Suffix = DefaultCreateSuffix(cre)
	}

	return cre, nil
}
