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

func transformModels(lookup *lookup, ast_models []*ast.Model) (
	models []*ir.Model, err error) {

	// step 1. create all the Model and Field instances and set their pointers
	// to point at each other appropriately.
	for _, ast_model := range ast_models {
		link, err := lookup.AddModel(ast_model)
		if err != nil {
			return nil, err
		}
		for _, ast_field := range ast_model.Fields {
			if err := link.AddField(ast_field); err != nil {
				return nil, err
			}
		}
	}

	// step 2. resolve all of the other fields on the models and Fields
	// including references between them. also check for duplicate table names.
	table_names := map[string]*ast.Model{}
	for _, ast_model := range ast_models {
		model_entry := lookup.GetModel(ast_model.Name.Value)
		if err := transformModel(lookup, model_entry); err != nil {
			return nil, err
		}

		model := model_entry.model

		if existing := table_names[model.Table]; existing != nil {
			return nil, errutil.New(ast_model.Pos,
				"%s: table %q already used by model %q (%s)",
				model.Table, existing.Name.Get(), existing.Pos)
		}
		table_names[model.Table] = ast_model

		models = append(models, model_entry.model)
	}

	return models, nil
}
