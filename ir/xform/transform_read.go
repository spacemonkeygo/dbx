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

func transformRead(lookup *lookup, ast_read *ast.Read) (
	reads []*ir.Read, err error) {

	tmpl := new(ir.Read)

	if ast_read.Select == nil || len(ast_read.Select.Refs) == 0 {
		return nil, Error.New("%s: no fields defined to select", ast_read.Pos)
	}

	// Figure out which models are needed for the fields and that the field
	// references aren't repetetive.
	selected := map[string]map[string]*ast.FieldRef{}
	for _, ast_fieldref := range ast_read.Select.Refs {
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
			return nil, Error.New(
				"%s: field %s already selected by field %s",
				ast_fieldref.Pos, ast_fieldref, existing)
		}
		fields[ast_fieldref.Field] = ast_fieldref

		if ast_fieldref.Field == "" {
			model, err := lookup.FindModel(ast_fieldref.ModelRef())
			if err != nil {
				return nil, err
			}
			tmpl.Selectables = append(tmpl.Selectables, model)
		} else {
			field, err := lookup.FindField(ast_fieldref)
			if err != nil {
				return nil, err
			}
			tmpl.Selectables = append(tmpl.Selectables, field)
		}
	}

	// Figure out set of models that are included in the read. These come from
	// explicit joins, or implicitly if there is only a single model referenced
	// in the fields.
	models := map[string]*ast.FieldRef{}
	switch {
	case len(ast_read.Joins) > 0:
		next := ast_read.Joins[0].Left.Model
		for _, join := range ast_read.Joins {
			left, err := lookup.FindField(join.Left)
			if err != nil {
				return nil, err
			}
			if join.Left.Model != next {
				return nil, Error.New(
					"%s: model order must be consistent; expected %q; got %q",
					join.Left.Pos, next, join.Left.Model)
			}
			right, err := lookup.FindField(join.Right)
			if err != nil {
				return nil, err
			}
			next = join.Right.Model
			if tmpl.From == nil {
				tmpl.From = left.Model
				models[join.Left.Model] = join.Left
			}
			tmpl.Joins = append(tmpl.Joins, &ir.Join{
				Type:  join.Type,
				Left:  left,
				Right: right,
			})
			if existing := models[join.Right.Model]; existing != nil {
				return nil, Error.New("%s: model %q already joined at %s",
					join.Right.Pos, join.Right.Model, existing.Pos)
			}
			models[join.Right.Model] = join.Right
		}
	case len(selected) == 1:
		from, err := lookup.FindModel(ast_read.Select.Refs[0].ModelRef())
		if err != nil {
			return nil, err
		}
		tmpl.From = from
		models[from.Name] = ast_read.Select.Refs[0]
	default:
		return nil, Error.New(
			"%s: cannot select from multiple models without a join",
			ast_read.Select.Pos)
	}

	// Make sure all of the fields are accounted for in the set of models
	for _, ast_fieldref := range ast_read.Select.Refs {
		if models[ast_fieldref.Model] == nil {
			return nil, Error.New(
				"%s: cannot select field/model %q; model %q is not joined",
				ast_fieldref.Pos, ast_fieldref, ast_fieldref.Model)
		}
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	for _, ast_where := range ast_read.Where {
		left, err := lookup.FindField(ast_where.Left)
		if err != nil {
			return nil, err
		}
		if models[ast_where.Left.Model] == nil {
			return nil, Error.New(
				"%s: invalid where condition %q; model %q is not joined",
				ast_where.Pos, ast_where, ast_where.Left.Model)
		}

		var right *ir.Field
		if ast_where.Right != nil {
			right, err = lookup.FindField(ast_where.Right)
			if err != nil {
				return nil, err
			}
			if models[ast_where.Right.Model] == nil {
				return nil, Error.New(
					"%s: invalid where condition %q; model %q is not joined",
					ast_where.Pos, ast_where, ast_where.Right.Model)
			}
		}

		tmpl.Where = append(tmpl.Where, &ir.Where{
			Op:    ast_where.Op,
			Left:  left,
			Right: right,
		})
	}

	// Finalize OrderBy and make sure referenced fields are part of the select
	if ast_read.OrderBy != nil {
		fields, err := resolveFieldRefs(lookup, ast_read.OrderBy.Fields.Refs)
		if err != nil {
			return nil, err
		}
		for _, order_by_field := range ast_read.OrderBy.Fields.Refs {
			if models[order_by_field.Model] == nil {
				return nil, Error.New(
					"%s: invalid orderby field %q; model %q is not joined",
					order_by_field.Pos, order_by_field, order_by_field.Model)
			}
		}

		tmpl.OrderBy = &ir.OrderBy{
			Fields:     fields,
			Descending: ast_read.OrderBy.Descending,
		}
	}

	// Now emit one select per view type (or one for all if unspecified)
	view := ast_read.View
	if view == nil {
		view = &ast.View{
			All: true,
		}
	}

	addView := func(v ir.View) {
		read_copy := *tmpl
		read_copy.View = v
		reads = append(reads, &read_copy)
	}

	if view.All {
		// template is already sufficient for "all"
		addView(ir.All)
	}
	if view.Count {
		addView(ir.Count)
	}
	if view.Has {
		addView(ir.Has)
	}
	if view.LimitOffset {
		if tmpl.One() {
			return nil, Error.New("%s: cannot limit/offset unique select",
				view.Pos)
		}
		addView(ir.LimitOffset)
	}
	if view.Paged {
		if tmpl.OrderBy != nil {
			return nil, Error.New(
				"%s: cannot page on table %s with order by",
				view.Pos, tmpl.From)
		}
		if tmpl.From.BasicPrimaryKey() == nil {
			return nil, Error.New(
				"%s: cannot page on table %s with composite primary key",
				view.Pos, tmpl.From)
		}
		if tmpl.One() {
			return nil, Error.New("%s: cannot page unique select",
				view.Pos)
		}
		addView(ir.Paged)
	}

	return reads, nil
}
