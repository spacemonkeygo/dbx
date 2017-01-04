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

type Where struct {
	Left  *Field
	Op    ast.Operator
	Right *Field
}

func WhereFieldEquals(field *Field) *Where {
	return &Where{
		Left: field,
		Op:   ast.EQ,
	}
}

func WhereFieldsEquals(fields ...*Field) (wheres []*Where) {
	if len(fields) == 0 {
		return nil
	}
	for _, field := range fields {
		wheres = append(wheres, WhereFieldEquals(field))
	}
	return wheres
}

func FilterWhere(wheres []*Where, op ast.Operator) (filtered []*Where) {
	for _, where := range wheres {
		if where.Op == op {
			filtered = append(filtered, where)
		}
	}
	return filtered
}

func WhereSetUnique(wheres []*Where) bool {
	// Aggregate fields involved in EQ relationships
	fields := map[*Model][]*Field{}
	for _, eq := range FilterWhere(wheres, ast.EQ) {
		fields[eq.Left.Model] = append(fields[eq.Left.Model], eq.Left)
		if eq.Right != nil {
			fields[eq.Right.Model] = append(fields[eq.Right.Model], eq.Right)
		}
	}

	// No where conditions that can provide unique contraints
	if len(fields) == 0 {
		return false
	}

	// If any of the where conditions for a given model do not uniquely identify
	// a single entry for that model, then the select can return more than one.
	for m, fs := range fields {
		if !m.FieldSetUnique(fs) {
			return false
		}
	}

	return true
}
