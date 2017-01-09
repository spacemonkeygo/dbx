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

func transformModel(lookup *lookup, model_entry *modelEntry) (err error) {
	model := model_entry.model
	ast_model := model_entry.ast

	model.Name = ast_model.Name
	model.Table = ast_model.Table

	for _, ast_field := range ast_model.Fields {
		field_entry := model_entry.GetField(ast_field.Name)
		if err := transformField(lookup, field_entry); err != nil {
			return err
		}
	}

	if ast_model.PrimaryKey == nil || len(ast_model.PrimaryKey.Refs) == 0 {
		return Error.New("%s: no primary key defined", ast_model.Pos)
	}

	for _, ast_fieldref := range ast_model.PrimaryKey.Refs {
		field, err := model_entry.FindField(ast_fieldref)
		if err != nil {
			return err
		}
		if field.Nullable {
			return Error.New("%s: nullable field %q cannot be a primary key",
				ast_fieldref.Pos, ast_fieldref.Field)
		}
		if field.Updatable {
			return Error.New("%s: updatable field %q cannot be a primary key",
				ast_fieldref.Pos, ast_fieldref.Field)
		}
		model.PrimaryKey = append(model.PrimaryKey, field)
	}

	for _, ast_unique := range ast_model.Unique {
		fields, err := resolveRelativeFieldRefs(model_entry, ast_unique.Refs)
		if err != nil {
			return err
		}
		model.Unique = append(model.Unique, fields)
	}

	index_names := map[string]*ast.Index{}
	for _, ast_index := range ast_model.Indexes {
		if existing, ok := index_names[ast_index.Name]; ok {
			return Error.New("%s: index %q already defined at %s",
				ast_index.Pos, ast_index.Name, existing.Pos)
		}
		index_names[ast_index.Name] = ast_index

		if ast_index.Fields == nil || len(ast_index.Fields.Refs) < 1 {
			return Error.New("%s: index %q has no fields defined",
				ast_index.Pos, ast_index.Name)
		}

		fields, err := resolveRelativeFieldRefs(
			model_entry, ast_index.Fields.Refs)
		if err != nil {
			return err
		}
		model.Indexes = append(model.Indexes, &ir.Index{
			Name:   ast_index.Name,
			Model:  fields[0].Model,
			Fields: fields,
			Unique: ast_index.Unique,
		})
	}

	return nil
}
