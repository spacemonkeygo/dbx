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

package ir

import (
	"fmt"
	"sort"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
)

// returns true if left is a subset of right
func fieldSetSubset(left, right []*Field) bool {
	if len(left) > len(right) {
		return false
	}
lcols:
	for _, lcol := range left {
		for _, rcol := range right {
			if lcol == rcol {
				continue lcols
			}
		}
		return false
	}
	return true
}

// returns true if left and right are equivalent (order agnostic)
func fieldSetEquivalent(left, right []*Field) bool {
	if len(left) != len(right) {
		return false
	}
	return fieldSetSubset(left, right)
}

// returns true if left is a subset of right
func fieldSetPrune(all, bad []*Field) (out []*Field) {
	for i := range all {
		if fieldSetSubset(all[i:i+1], bad) {
			continue
		}
		out = append(out, all[i])
	}
	return out
}

func whereUnique(wheres []*Where) (unique map[string]bool) {
	fields := map[*Model][]*Field{}
	for _, eq := range FilterWhere(wheres, consts.EQ) {
		fields[eq.Left.Model] = append(fields[eq.Left.Model], eq.Left)
		if eq.Right != nil {
			fields[eq.Right.Model] = append(fields[eq.Right.Model], eq.Right)
		}
	}

	unique = map[string]bool{}
	for m, fs := range fields {
		unique[m.Name] = m.FieldSetUnique(fs)
	}
	return unique
}

func queryUnique(targets []*Model, joins []*Join, wheres []*Where) (out bool) {
	// Build up a list of models involved in the query.
	unique := map[string]bool{}

	// Contrain based on the where conditions
	for model_name, model_unique := range whereUnique(wheres) {
		unique[model_name] = model_unique
	}

	// Constrain based on joins with unique columns
	for _, join := range joins {
		switch join.Type {
		case consts.InnerJoin:
			if unique[join.Left.Model.Name] {
				if join.Right.Unique() {
					unique[join.Right.Model.Name] = true
				}
				if join.Right.Relation != nil &&
					join.Right.Relation.Field.Unique() {
					unique[join.Right.Relation.Field.Model.Name] = true
				}
			}
			if unique[join.Right.Model.Name] {
				if join.Left.Unique() {
					unique[join.Left.Model.Name] = true
				}
				if join.Left.Relation != nil &&
					join.Left.Relation.Field.Unique() {
					unique[join.Left.Relation.Field.Model.Name] = true
				}
			}
		default:
			panic(fmt.Sprintf("unhandled join type %q", join.Type))
		}
	}

	// if any table from the set of targets is unique, then only one row would
	// ever be returned, so it is a "unique" query.
	for _, target := range targets {
		if unique[target.Name] {
			return true
		}
	}

	return false
}

func SortModels(models []*Model) (sorted []*Model) {
	// sort the slice copy
	sorted = append([]*Model(nil), models...)
	sort.Sort(byModelDepth(sorted))
	return sorted
}

type byModelDepth []*Model

func (by byModelDepth) Len() int {
	return len(by)
}

func (by byModelDepth) Swap(a, b int) {
	by[a], by[b] = by[b], by[a]
}

func (by byModelDepth) Less(a, b int) bool {
	adepth := modelDepth(by[a])
	bdepth := modelDepth(by[b])
	if adepth < bdepth {
		return true
	}
	if adepth > bdepth {
		return false
	}
	return by[a].Name < by[b].Name
}

func modelDepth(model *Model) (depth int) {
	for _, field := range model.Fields {
		if field.Relation == nil {
			continue
		}
		reldepth := modelDepth(field.Relation.Field.Model) + 1
		if reldepth > depth {
			depth = reldepth
		}
	}
	return depth
}

func distinctModels(models []*Model) (distinct []*Model) {
	set := map[string]*Model{}
	var names []string
	for _, model := range models {
		if set[model.Name] != nil {
			continue
		}
		set[model.Name] = model
		names = append(names, model.Name)
	}
	sort.Sort(sort.StringSlice(names))

	for _, name := range names {
		distinct = append(distinct, set[name])
	}
	return distinct
}
