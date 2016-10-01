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

import "fmt"

type RelationKind string

const (
	HasA    RelationKind = "has_a"
	OwnedBy RelationKind = "owned_by"
)

type Relation struct {
	Column *Column
	Kind   RelationKind
}

type Column struct {
	Table      *Table
	Name       string
	Type       string
	NotNull    bool
	Relation   *Relation
	AutoInsert bool
}

func (c *Column) String() string {
	return fmt.Sprintf("%s.%s", c.Table.Name, c.Name)
}

func (c *Column) IsInt() bool {
	switch c.Type {
	case "serial", "serial64", "int", "int64":
		return true
	}
	return false
}

type Table struct {
	Name       string
	Columns    []*Column
	Unique     [][]*Column
	PrimaryKey []*Column
}

func (t *Table) Depth() (depth int) {
	for _, table := range t.Columns {
		if table.Relation != nil {
			reldepth := table.Relation.Column.Table.Depth() + 1
			if reldepth > depth {
				depth = reldepth
			}
		}
	}
	return depth
}

type Query struct {
	Table *Table
	Start []*Column
	Joins []*Join
	End   []*Column
}

func (q *Query) String() (s string) {
	s += q.Table.Name + "("
	if len(q.Start) > 0 {
		s += "start=("
		for i, start := range q.Start {
			if i != 0 {
				s += ", "
			}
			s += start.String()
		}
		s += ")"
	}
	if len(q.Joins) > 0 {
		s += "joins=("
		for i, join := range q.Joins {
			if i != 0 {
				s += ", "
			}
			s += join.Left.String() + "|" + join.Right.String()
		}
		s += ")"
	}
	if len(q.End) > 0 {
		s += "end=("
		for i, end := range q.End {
			if i != 0 {
				s += ", "
			}
			s += end.String()
		}
		s += ")"
	}
	s += ")"
	return s
}

type Join struct {
	Left  *Column
	Right *Column
}

type Schema struct {
	Tables  []*Table
	Queries []*Query
}

func (c *Column) RelationLeft() *Join {
	if c.Relation == nil {
		return nil
	}
	return &Join{
		Left:  c,
		Right: c.Relation.Column,
	}
}

func (c *Column) RelationRight() *Join {
	if c.Relation == nil {
		return nil
	}
	return &Join{
		Left:  c.Relation.Column,
		Right: c,
	}
}

func (c *Column) Insertable() bool {
	if c.Relation != nil {
		return true
	}
	return c.Type != "serial" && c.Type != "serial64"
}

func (t *Table) GetColumn(name string) *Column {
	for _, column := range t.Columns {
		if column.Name == name {
			return column
		}
	}
	return nil
}

func (t *Table) GetColumns(names ...string) (out []*Column) {
	for _, name := range names {
		column := t.GetColumn(name)
		if column == nil {
			return nil
		}
		out = append(out, column)
	}
	return out
}

func (t *Table) InsertableColumns() (out []*Column) {
	for _, column := range t.Columns {
		if !column.Insertable() {
			continue
		}
		out = append(out, column)
	}
	return out
}

// returns true if left is a subset of right
func columnSetSubset(left, right []*Column) bool {
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
func columnSetEquivalent(left, right []*Column) bool {
	if len(left) != len(right) {
		return false
	}
	return columnSetSubset(left, right)
}

func (t *Table) BasicPrimaryKey() *Column {
	if len(t.PrimaryKey) == 1 && t.PrimaryKey[0].IsInt() {
		return t.PrimaryKey[0]
	}
	return nil
}

func (t *Table) ColumnSetUnique(columns []*Column) bool {
	if columnSetSubset(t.PrimaryKey, columns) {
		return true
	}
	for _, unique := range t.Unique {
		if columnSetSubset(unique, columns) {
			return true
		}
	}
	return false
}
