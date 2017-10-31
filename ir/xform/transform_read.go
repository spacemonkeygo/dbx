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

func transformRead(lookup *lookup, ast_read *ast.Read) (
	reads []*ir.Read, err error) {

	tmpl := &ir.Read{
		Suffix: transformSuffix(ast_read.Suffix),
	}

	if ast_read.Select == nil || len(ast_read.Select.Refs) == 0 {
		return nil, errutil.New(ast_read.Pos, "no fields defined to select")
	}

	// Figure out which models are needed for the fields and that the field
	// references aren't repetetive.
	selected := map[string]map[string]*ast.FieldRef{}
	for _, ast_fieldref := range ast_read.Select.Refs {
		fields := selected[ast_fieldref.Model.Value]
		if fields == nil {
			fields = map[string]*ast.FieldRef{}
			selected[ast_fieldref.Model.Value] = fields
		}

		existing := fields[""]
		if existing == nil {
			existing = fields[ast_fieldref.Field.Get()]
		}
		if existing != nil {
			return nil, errutil.New(ast_fieldref.Pos,
				"field %q already selected by field %q",
				ast_fieldref, existing)
		}
		fields[ast_fieldref.Field.Get()] = ast_fieldref

		if ast_fieldref.Field.Get() == "" {
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

	models, joins, err := transformJoins(lookup, ast_read.Joins)
	if err != nil {
		return nil, err
	}

	tmpl.Joins = joins

	if len(joins) > 0 {
		tmpl.From = joins[0].Left.Model
	} else if len(selected) == 1 {
		sel := ast_read.Select.Refs[0]
		from, err := lookup.FindModel(sel.ModelRef())
		if err != nil {
			return nil, err
		}
		tmpl.From = from
		models[sel.Model.Value] = sel.Pos
	} else {
		return nil, errutil.New(ast_read.Select.Pos,
			"cannot select from multiple models without a join")
	}

	// Make sure all of the fields are accounted for in the set of models
	for _, ast_fieldref := range ast_read.Select.Refs {
		if _, ok := models[ast_fieldref.Model.Value]; !ok {
			return nil, errutil.New(ast_fieldref.Pos,
				"cannot select %q; model %q is not joined",
				ast_fieldref, ast_fieldref.Model.Value)
		}
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	tmpl.Where, err = transformWheres(lookup, models, ast_read.Where)
	if err != nil {
		return nil, err
	}

	// Finalize OrderBy and make sure referenced fields are part of the select
	if ast_read.OrderBy != nil {
		fields, err := resolveFieldRefs(lookup, ast_read.OrderBy.Fields.Refs)
		if err != nil {
			return nil, err
		}
		for _, order_by_field := range ast_read.OrderBy.Fields.Refs {
			if _, ok := models[order_by_field.Model.Value]; !ok {
				return nil, errutil.New(order_by_field.Pos,
					"invalid orderby field %q; model %q is not joined",
					order_by_field, order_by_field.Model.Value)
			}
		}

		tmpl.OrderBy = &ir.OrderBy{
			Fields:     fields,
			Descending: ast_read.OrderBy.Descending.Get(),
		}
	}

	// Finalize GroupBy and make sure referenced fields are part of the select
	if ast_read.GroupBy != nil {
		fields, err := resolveFieldRefs(lookup, ast_read.GroupBy.Fields.Refs)
		if err != nil {
			return nil, err
		}
		for _, group_by_field := range ast_read.GroupBy.Fields.Refs {
			if _, ok := models[group_by_field.Model.Value]; !ok {
				return nil, errutil.New(group_by_field.Pos,
					"invalid groupby field %q; model %q is not joined",
					group_by_field, group_by_field.Model.Value)
			}
		}

		tmpl.GroupBy = &ir.GroupBy{
			Fields: fields,
		}
	}

	// Now emit one select per view type (or one for all if unspecified)
	view := ast_read.View
	if view == nil {
		view = &ast.View{
			All: &ast.Bool{Value: true},
		}
	}

	addView := func(v ir.View) {
		read_copy := *tmpl
		read_copy.View = v
		if read_copy.Suffix == nil {
			read_copy.Suffix = DefaultReadSuffix(&read_copy)
		}
		reads = append(reads, &read_copy)
	}

	if view.All.Get() {
		if tmpl.Distinct() {
			return nil, errutil.New(view.All.Pos,
				"cannot limit/offset unique select")
		}
		addView(ir.All)
	}
	if view.Count.Get() {
		addView(ir.Count)
	}
	if view.Has.Get() {
		addView(ir.Has)
	}
	if view.LimitOffset.Get() {
		if tmpl.Distinct() {
			return nil, errutil.New(view.LimitOffset.Pos,
				"cannot use limitoffset view with distinct read")
		}
		addView(ir.LimitOffset)
	}
	if view.Paged.Get() {
		if tmpl.Distinct() {
			return nil, errutil.New(view.LimitOffset.Pos,
				"cannot use paged view with distinct read")
		}
		if tmpl.OrderBy != nil {
			return nil, errutil.New(view.Paged.Pos,
				"cannot page on model %q with order by",
				tmpl.From.Name)
		}
		if tmpl.GroupBy != nil {
			// Unless the primary key is part of the group by, then you can't
			// know which row the primary key would be chosen by. Not sure
			// this type of query would be useful, even if we could verify
			// that it was ok, so disabling for now.
			return nil, errutil.New(view.Paged.Pos,
				"cannot page on model %q with group by",
				tmpl.From.Name)
		}
		if tmpl.From.BasicPrimaryKey() == nil {
			return nil, errutil.New(view.Paged.Pos,
				"cannot page on model %q with composite primary key",
				tmpl.From.Name)
		}
		addView(ir.Paged)
	}
	if view.Scalar.Get() {
		addView(ir.Scalar)
	}
	if view.One.Get() {
		addView(ir.One)
	}
	if view.First.Get() {
		addView(ir.First)
	}

	return reads, nil
}
