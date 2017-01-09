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

func Transform(ast_root *ast.Root) (root *Root, err error) {
	models, lookup, err := transformModels(ast_root.Models)
	if err != nil {
		return nil, err
	}

	root = &Root{
		Models: models,
	}

	for _, ast_update := range ast_root.Updates {
		upd, err := transformUpdate(lookup, ast_update)
		if err != nil {
			return nil, err
		}
		root.Updates = append(root.Updates, upd)
	}

	for _, ast_del := range ast_root.Deletes {
		del, err := transformDelete(lookup, ast_del)
		if err != nil {
			return nil, err
		}
		root.Deletes = append(root.Deletes, del)
	}

	for _, ast_sel := range ast_root.Selects {
		selects, err := transformSelect(lookup, ast_sel)
		if err != nil {
			return nil, err
		}
		root.Selects = append(root.Selects, selects...)
	}

	return root, nil
}

func resolveFieldRefs(lookup *lookup, ast_refs []*ast.FieldRef) (
	fields []*Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := lookup.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func resolveRelativeFieldRefs(model_entry *modelEntry,
	ast_refs []*ast.RelativeFieldRef) (fields []*Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := model_entry.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}
