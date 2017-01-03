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

package dbx

type JoinParams struct {
	Left  *Column
	Right *Column
}

func Joins(joins ...*Join) (out []*JoinParams) {
	for _, join := range joins {
		out = append(out, &JoinParams{
			Left:  join.Left,
			Right: join.Right,
		})
	}

	return out
}

func Where(conditions ...*ConditionParams) []*ConditionParams {
	return conditions
}

type ColumnCmpParams struct {
	Left     *Column
	Operator string
}

type ColumnCmpColumnParams struct {
	Left     *Column
	Right    *Column
	Operator string
}

type ColumnInParams struct {
	Left *Column
	In   *SelectParams
}

type ConditionParams struct {
	ColumnCmp       *ColumnCmpParams
	ColumnCmpColumn *ColumnCmpColumnParams
	ColumnIn        *ColumnInParams
}

func ColumnEquals(left *Column) *ConditionParams {
	return &ConditionParams{
		ColumnCmp: &ColumnCmpParams{
			Left:     left,
			Operator: "=",
		},
	}
}

func ColumnEqualsColumn(left, right *Column) *ConditionParams {
	return &ConditionParams{
		ColumnCmpColumn: &ColumnCmpColumnParams{
			Left:     left,
			Right:    right,
			Operator: "=",
		},
	}
}

func ColumnIn(left *Column, in *SelectParams) *ConditionParams {
	return &ConditionParams{
		ColumnIn: &ColumnInParams{
			Left: left,
			In:   in,
		},
	}
}

type SelectParams struct {
	Many       bool
	What       []*Column
	Table      *Table
	LeftJoins  []*JoinParams
	Conditions []*ConditionParams
	PagedOn    *Column
}

func What(columns ...*Column) []*Column {
	return columns
}

type DeleteParams struct {
	Many       bool
	Table      *Table
	Conditions []*ConditionParams
}

type InsertParams struct {
	Table   *Table
	Columns []*Column
}

type UpdateParams struct {
	Table      *Table
	Conditions []*ConditionParams
}
