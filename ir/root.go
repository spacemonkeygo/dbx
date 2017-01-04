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

package ir

import "gopkg.in/spacemonkeygo/dbx.v1/ast"

type Root struct {
	Models  *Models
	Inserts []*Insert
	Updates []*Update
	Selects []*Select
	Deletes []*Delete
}

func Transform(ast_root *ast.Root) (root *Root, err error) {
	root = new(Root)

	root.Models, err = TransformModels(ast_root.Models)
	if err != nil {
		return nil, err
	}

	for _, ast_sel := range ast_root.Selects {
		sel, err := root.Models.CreateSelect(ast_sel)
		if err != nil {
			return nil, err
		}
		root.Selects = append(root.Selects, sel)
	}

	return root, nil
}
