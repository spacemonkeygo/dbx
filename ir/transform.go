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

import (
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
)

func Transform(ast_root *ast.Root) (*Root, error) {
	return newTransformer(ast_root).Transform()
}

func newTransformer(ast_root *ast.Root) *transformer {
	return &transformer{
		linker: newLinker(),

		ast_root: ast_root,
		ir_root:  new(Root),
	}
}

type transformer struct {
	*linker

	ast_root *ast.Root
	ir_root  *Root
}

func (t *transformer) Transform() (*Root, error) {
	// pass 1. build up the ir structures in a way where we can uniquely
	// reference them.
	for _, ast_model := range t.ast_root.Models {
		link, err := t.AddModel(ast_model)
		if err != nil {
			return nil, err
		}
		for _, ast_field := range ast_model.Fields {
			if err := link.AddField(ast_field); err != nil {
				return nil, err
			}
		}
	}

	for _, ast_model := range t.ast_root.Models {
		model_link := t.GetModel(ast_model.Name)
		if err := t.transformModel(model_link); err != nil {
			return nil, err
		}
		t.ir_root.Models = append(t.ir_root.Models, model_link.model)
	}

	for _, ast_sel := range t.ast_root.Selects {
		if err := t.transformSelect(ast_sel); err != nil {
			return nil, err
		}
	}

	return t.ir_root, nil
}

func (t *transformer) transformModel(model_link *modelLink) (err error) {
	model := model_link.model
	ast_model := model_link.ast_model

	model.Name = ast_model.Name
	model.Table = ast_model.Table

	for _, ast_field := range ast_model.Fields {
		field_link := model_link.GetField(ast_field.Name)
		if err := t.transformField(field_link); err != nil {
			return err
		}
		model.Fields = append(model.Fields, field_link.field)
	}

	if len(ast_model.PrimaryKey) == 0 {
		return Error.New("%s: no primary key defined", ast_model.Pos)
	}

	for _, ast_fieldref := range ast_model.PrimaryKey {
		field, err := model_link.FindField(ast_fieldref)
		if err != nil {
			return err
		}
		model.PrimaryKey = append(model.PrimaryKey, field)
	}

	for _, ast_unique := range ast_model.Unique {
		fields, err := resolveRelativeFieldRefs(model_link, ast_unique)
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

		fields, err := resolveRelativeFieldRefs(model_link, ast_index.Fields)
		if err != nil {
			return err
		}
		model.Indexes = append(model.Indexes, &Index{
			Name:   ast_index.Name,
			Fields: fields,
		})
	}

	return nil
}

func (t *transformer) transformField(field_link *fieldLink) (err error) {
	field := field_link.field
	ast_field := field_link.ast_field

	field.Name = ast_field.Name
	field.Type = ast_field.Type
	field.Column = ast_field.Column
	field.Nullable = ast_field.Nullable
	field.Updatable = ast_field.Updatable
	field.AutoInsert = ast_field.AutoInsert
	field.AutoUpdate = ast_field.AutoUpdate
	field.Length = ast_field.Length

	if ast_field.Relation != nil {
		related, err := t.FindField(ast_field.Relation.FieldRef)
		if err != nil {
			return err
		}
		field.Relation = &Relation{
			Field: related,
		}
		field.Type = related.Type.AsLink()
	}

	return nil
}

func (t *transformer) transformSelect(ast_sel *ast.Select) (err error) {
	sel := new(Select)

	var func_suffix []string
	if ast_sel.Fields == nil || len(ast_sel.Fields.Refs) == 0 {
		return Error.New("%s: no fields defined to select", ast_sel.Pos)
	}

	// Figure out which models are needed for the fields and that the field
	// references aren't repetetive.
	selected := map[string]map[string]*ast.FieldRef{}
	for _, ast_fieldref := range ast_sel.Fields.Refs {
		fields := selected[ast_fieldref.Model]
		if fields == nil {
			fields = map[string]*ast.FieldRef{}
			selected[ast_fieldref.Model] = fields
		}

		existing := fields[""]
		if existing == nil {
			existing = fields[ast_fieldref.Field]
		}
		if existing != nil {
			return Error.New(
				"%s: field %s already selected by field %s",
				ast_fieldref.Pos, ast_fieldref, existing)
		}
		fields[ast_fieldref.Field] = ast_fieldref

		if ast_fieldref.Field == "" {
			model, err := t.FindModel(ast_fieldref)
			if err != nil {
				return err
			}
			sel.Fields = append(sel.Fields, model)
			func_suffix = append(func_suffix, ast_fieldref.Model)
		} else {
			field, err := t.FindField(ast_fieldref)
			if err != nil {
				return err
			}
			sel.Fields = append(sel.Fields, field)
			func_suffix = append(func_suffix,
				ast_fieldref.Model, ast_fieldref.Field)
		}
	}

	// Figure out set of models that are included in the select. These come from
	// explicit joins, or implicitly if there is only a single model referenced
	// in the fields.
	models := map[string]*ast.FieldRef{}
	switch {
	case len(ast_sel.Joins) > 0:
		next := ast_sel.Joins[0].Left.Model
		for _, join := range ast_sel.Joins {
			left, err := t.FindField(join.Left)
			if err != nil {
				return err
			}
			if join.Left.Model != next {
				return Error.New(
					"%s: model order must be consistent; expected %q; got %q",
					join.Left.Pos, next, join.Left.Model)
			}
			right, err := t.FindField(join.Right)
			if err != nil {
				return err
			}
			next = join.Right.Model
			if sel.From == nil {
				sel.From = left.Model
				models[join.Left.Model] = join.Left
			}
			sel.Joins = append(sel.Joins, &Join{
				Type:  join.Type,
				Left:  left,
				Right: right,
			})
			if existing := models[join.Right.Model]; existing != nil {
				return Error.New("%s: model %q already joined at %s",
					join.Right.Pos, join.Right.Model, existing.Pos)
			}
			models[join.Right.Model] = join.Right
		}
	case len(selected) == 1:
		from, err := t.FindModel(ast_sel.Fields.Refs[0])
		if err != nil {
			return err
		}
		sel.From = from
		models[from.Name] = ast_sel.Fields.Refs[0]
	default:
		return Error.New(
			"%s: cannot select from multiple models without a join",
			ast_sel.Fields.Pos)
	}

	// Make sure all of the fields are accounted for in the set of models
	for _, ast_fieldref := range ast_sel.Fields.Refs {
		if models[ast_fieldref.Model] == nil {
			return Error.New(
				"%s: cannot select field/model %q; model %q is not joined",
				ast_fieldref.Pos, ast_fieldref, ast_fieldref.Model)
		}
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	if len(ast_sel.Where) > 0 {
		func_suffix = append(func_suffix, "by")
	}
	for _, ast_where := range ast_sel.Where {
		left, err := t.FindField(ast_where.Left)
		if err != nil {
			return err
		}
		if models[ast_where.Left.Model] == nil {
			return Error.New(
				"%s: invalid where condition %q; model %q is not joined",
				ast_where.Pos, ast_where, ast_where.Left.Model)
		}

		var right *Field
		if ast_where.Right != nil {
			right, err = t.FindField(ast_where.Right)
			if err != nil {
				return err
			}
			if models[ast_where.Right.Model] == nil {
				return Error.New(
					"%s: invalid where condition %q; model %q is not joined",
					ast_where.Pos, ast_where, ast_where.Right.Model)
			}
		} else {
			func_suffix = append(func_suffix,
				ast_where.Left.Model, ast_where.Left.Field)
		}

		sel.Where = append(sel.Where, &Where{
			Op:    ast_where.Op,
			Left:  left,
			Right: right,
		})
	}

	// Finalize OrderBy and make sure referenced fields are part of the select
	if ast_sel.OrderBy != nil {
		fields, err := t.resolveFields(ast_sel.OrderBy.Fields.Refs)
		if err != nil {
			return err
		}
		for _, order_by_field := range ast_sel.OrderBy.Fields.Refs {
			if models[order_by_field.Model] == nil {
				return Error.New(
					"%s: invalid orderby field %q; model %q is not joined",
					order_by_field.Pos, order_by_field, order_by_field.Model)
			}
		}

		sel.OrderBy = &OrderBy{
			Fields:     fields,
			Descending: ast_sel.OrderBy.Descending,
		}
	}

	if sel.FuncSuffix == "" {
		sel.FuncSuffix = strings.Join(func_suffix, "_")
	}

	t.ir_root.Selects = append(t.ir_root.Selects, sel)
	return nil
}

func (t *transformer) resolveFields(ast_refs []*ast.FieldRef) (
	fields []*Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := t.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func resolveRelativeFieldRefs(model_link *modelLink,
	ast_refs []*ast.RelativeFieldRef) (fields []*Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := model_link.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}
