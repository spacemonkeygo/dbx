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

func transformSelect(lookup *lookup, ast_sel *ast.Select) (
	selects []*Select, err error) {

	sel_tmpl := &Select{
		FuncSuffix: "",
	}

	var func_suffix []string
	if ast_sel.Fields == nil || len(ast_sel.Fields.Refs) == 0 {
		return nil, Error.New("%s: no fields defined to select", ast_sel.Pos)
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
			sel_tmpl.Fields = append(sel_tmpl.Fields, model)
			func_suffix = append(func_suffix, ast_fieldref.Model)
		} else {
			field, err := lookup.FindField(ast_fieldref)
			if err != nil {
				return nil, err
			}
			sel_tmpl.Fields = append(sel_tmpl.Fields, field)
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
			if sel_tmpl.From == nil {
				sel_tmpl.From = left.Model
				models[join.Left.Model] = join.Left
			}
			sel_tmpl.Joins = append(sel_tmpl.Joins, &Join{
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
		from, err := lookup.FindModel(ast_sel.Fields.Refs[0].ModelRef())
		if err != nil {
			return nil, err
		}
		sel_tmpl.From = from
		models[from.Name] = ast_sel.Fields.Refs[0]
	default:
		return nil, Error.New(
			"%s: cannot select from multiple models without a join",
			ast_sel.Fields.Pos)
	}

	// Make sure all of the fields are accounted for in the set of models
	for _, ast_fieldref := range ast_sel.Fields.Refs {
		if models[ast_fieldref.Model] == nil {
			return nil, Error.New(
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
		left, err := lookup.FindField(ast_where.Left)
		if err != nil {
			return nil, err
		}
		if models[ast_where.Left.Model] == nil {
			return nil, Error.New(
				"%s: invalid where condition %q; model %q is not joined",
				ast_where.Pos, ast_where, ast_where.Left.Model)
		}

		var right *Field
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
		} else {
			func_suffix = append(func_suffix,
				ast_where.Left.Model, ast_where.Left.Field)
		}

		sel_tmpl.Where = append(sel_tmpl.Where, &Where{
			Op:    ast_where.Op,
			Left:  left,
			Right: right,
		})
	}

	// Finalize OrderBy and make sure referenced fields are part of the select
	if ast_sel.OrderBy != nil {
		fields, err := resolveFieldRefs(lookup, ast_sel.OrderBy.Fields.Refs)
		if err != nil {
			return nil, err
		}
		for _, order_by_field := range ast_sel.OrderBy.Fields.Refs {
			if models[order_by_field.Model] == nil {
				return nil, Error.New(
					"%s: invalid orderby field %q; model %q is not joined",
					order_by_field.Pos, order_by_field, order_by_field.Model)
			}
		}

		sel_tmpl.OrderBy = &OrderBy{
			Fields:     fields,
			Descending: ast_sel.OrderBy.Descending,
		}
	}

	if sel_tmpl.FuncSuffix == "" {
		sel_tmpl.FuncSuffix = strings.Join(func_suffix, "_")
	}

	// Now emit one select per view type (or one for all if unspecified)
	view := ast_sel.View
	if view == nil {
		view = &ast.View{
			All: true,
		}
	}

	appendsel := func(v View, suffix string) {
		sel_copy := *sel_tmpl
		sel_copy.View = v
		sel_copy.FuncSuffix += suffix
		selects = append(selects, &sel_copy)
	}

	if view.All {
		// template is already sufficient for "all"
		appendsel(All, "")
	}
	if view.Limit {
		if sel_tmpl.One() {
			return nil, Error.New("%s: cannot limit unique select",
				view.Pos)
		}
		appendsel(Limit, "_limit")
	}
	if view.LimitOffset {
		if sel_tmpl.One() {
			return nil, Error.New("%s: cannot limit/offset unique select",
				view.Pos)
		}
		appendsel(LimitOffset, "_limit_offset")
	}
	if view.Offset {
		if sel_tmpl.One() {
			return nil, Error.New("%s: cannot offset unique select",
				view.Pos)
		}
		appendsel(Offset, "_offset")
	}
	if view.Paged {
		if sel_tmpl.OrderBy != nil {
			return nil, Error.New(
				"%s: cannot page on table %s with order by",
				view.Pos, sel_tmpl.From)
		}
		if sel_tmpl.From.BasicPrimaryKey() == nil {
			return nil, Error.New(
				"%s: cannot page on table %s with composite primary key",
				view.Pos, sel_tmpl.From)
		}
		if sel_tmpl.One() {
			return nil, Error.New("%s: cannot page unique select",
				view.Pos)
		}
		appendsel(Paged, "_paged")
	}

	return selects, nil
}
