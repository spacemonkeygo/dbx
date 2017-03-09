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

package ast

import (
	"fmt"
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
)

type Root struct {
	Models  []*Model
	Creates []*Create
	Reads   []*Read
	Updates []*Update
	Deletes []*Delete
}

type String struct {
	Pos   scanner.Position
	Value string
}

func (s *String) Get() string {
	if s == nil {
		return ""
	}
	return s.Value
}

type Model struct {
	Pos        scanner.Position
	Name       *String
	Table      *String
	Fields     []*Field
	PrimaryKey *RelativeFieldRefs
	Unique     []*RelativeFieldRefs
	Indexes    []*Index
}

type Bool struct {
	Pos   scanner.Position
	Value bool
}

func (b *Bool) Get() bool {
	if b == nil {
		return false
	}
	return b.Value
}

type Int struct {
	Pos   scanner.Position
	Value int
}

func (i *Int) Get() int {
	if i == nil {
		return 0
	}
	return i.Value
}

type Suffix struct {
	Pos   scanner.Position
	Parts []*String
}

type Field struct {
	Pos  scanner.Position
	Name *String

	// Common to both regular and relation fields
	Column    *String
	Nullable  *Bool
	Updatable *Bool

	// Only make sense on a regular field
	Type       *FieldType
	AutoInsert *Bool
	AutoUpdate *Bool
	Length     *Int

	// Only make sense on a relation
	Relation     *FieldRef
	RelationKind *RelationKind
}

type RelationKind struct {
	Pos   scanner.Position
	Value consts.RelationKind
}

type FieldType struct {
	Pos   scanner.Position
	Value consts.FieldType
}

type FieldRef struct {
	Pos   scanner.Position
	Model *String
	Field *String
}

func (r *FieldRef) String() string {
	if r.Field == nil {
		return r.Model.Value
	}
	if r.Model == nil {
		return r.Field.Value
	}
	return fmt.Sprintf("%s.%s", r.Model.Value, r.Field.Value)
}

func (f *FieldRef) Relative() *RelativeFieldRef {
	return &RelativeFieldRef{
		Pos:   f.Pos,
		Field: f.Field,
	}
}

func (f *FieldRef) ModelRef() *ModelRef {
	return &ModelRef{
		Pos:   f.Pos,
		Model: f.Model,
	}
}

type RelativeFieldRefs struct {
	Pos  scanner.Position
	Refs []*RelativeFieldRef
}

type RelativeFieldRef struct {
	Pos   scanner.Position
	Field *String
}

func (r *RelativeFieldRef) String() string {
	return r.Field.Value
}

type ModelRef struct {
	Pos   scanner.Position
	Model *String
}

func (m *ModelRef) String() string {
	return m.Model.Value
}

type Index struct {
	Pos    scanner.Position
	Name   *String
	Fields *RelativeFieldRefs
	Unique *Bool
}

type Read struct {
	Pos     scanner.Position
	Select  *FieldRefs
	Joins   []*Join
	Where   []*Where
	OrderBy *OrderBy
	View    *View
	Suffix  *Suffix
}

type Delete struct {
	Pos    scanner.Position
	Model  *ModelRef
	Joins  []*Join
	Where  []*Where
	Suffix *Suffix
}

type Update struct {
	Pos    scanner.Position
	Model  *ModelRef
	Joins  []*Join
	Where  []*Where
	Suffix *Suffix
}

type Create struct {
	Pos    scanner.Position
	Model  *ModelRef
	Raw    *Bool
	Suffix *Suffix
}

type View struct {
	Pos         scanner.Position
	All         *Bool
	LimitOffset *Bool
	Paged       *Bool
	Count       *Bool
	Has         *Bool
	Scalar      *Bool
	One         *Bool
	First       *Bool
}

type FieldRefs struct {
	Pos  scanner.Position
	Refs []*FieldRef
}

type Join struct {
	Pos   scanner.Position
	Left  *FieldRef
	Right *FieldRef
	Type  *JoinType
}

type JoinType struct {
	Pos   scanner.Position
	Value consts.JoinType
}

func (j *JoinType) Get() consts.JoinType {
	if j == nil {
		return consts.InnerJoin
	}
	return j.Value
}

type Where struct {
	Pos   scanner.Position
	Left  *FieldRef
	Op    *Operator
	Right *FieldRef
}

func (w *Where) String() string {
	right := "?"
	if w.Right != nil {
		right = w.Right.String()
	}
	return fmt.Sprintf("%s %s %s", w.Left, w.Op, right)
}

type Operator struct {
	Pos   scanner.Position
	Value consts.Operator
}

func (o *Operator) String() string { return string(o.Value) }

type OrderBy struct {
	Pos        scanner.Position
	Fields     *FieldRefs
	Descending *Bool
}
