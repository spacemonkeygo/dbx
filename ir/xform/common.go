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
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func resolveFieldRefs(lookup *lookup, ast_refs []*ast.FieldRef) (
	fields []*ir.Field, err error) {

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
	ast_refs []*ast.RelativeFieldRef) (fields []*ir.Field, err error) {

	for _, ast_ref := range ast_refs {
		field, err := model_entry.FindField(ast_ref)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func previouslyDefined(pos scanner.Position, kind string,
	where scanner.Position) error {

	return errutil.New(pos,
		"%s already defined. previous definition at %s",
		kind, where)
}

func duplicateQuery(pos scanner.Position, kind string,
	where scanner.Position) error {
	return errutil.New(pos,
		"%s: duplicate %s (first defined at %s)",
		pos, kind, where)
}
