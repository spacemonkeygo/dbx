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
	"strings"
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

func (b *Bool) String() string {
	return fmt.Sprint(b.Value)
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
	Pos      scanner.Position
	Model    *ModelRef
	Joins    []*Join
	Where    []*Where
	NoReturn *Bool
	Suffix   *Suffix
}

type Create struct {
	Pos      scanner.Position
	Model    *ModelRef
	Raw      *Bool
	NoReturn *Bool
	Suffix   *Suffix
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
	Pos    scanner.Position
	Or     []*Where
	And    []*Where
	Clause *WhereClause
}

func (w *Where) String() string {
	switch {
	case w.Or != nil:
		out := make([]string, 0, len(w.Or))
		for _, where := range w.Or {
			out = append(out, where.String())
		}
		return fmt.Sprintf("(%s)", strings.Join(out, " or "))
	case w.And != nil:
		out := make([]string, 0, len(w.And))
		for _, where := range w.And {
			out = append(out, where.String())
		}
		return fmt.Sprintf("(%s)", strings.Join(out, " and "))
	case w.Clause != nil:
		return w.Clause.String()
	default:
		return "<invalid where>"
	}
}

type WhereClause struct {
	Pos   scanner.Position
	Left  *Expr
	Op    *Operator
	Right *Expr
}

func (w *WhereClause) String() string {
	return fmt.Sprintf("%s %s %s", w.Left, w.Op, w.Right)
}

type Expr struct {
	Pos scanner.Position
	// The following fields are mutually exclusive
	Null        *Null
	StringLit   *String
	NumberLit   *String
	BoolLit     *Bool
	Placeholder *Placeholder
	FieldRef    *FieldRef
	FuncCall    *FuncCall
}

func (e *Expr) String() string {
	switch {
	case e.Null != nil:
		return e.Null.String()
	case e.StringLit != nil:
		return fmt.Sprintf("%q", e.StringLit.Value)
	case e.NumberLit != nil:
		return e.NumberLit.Value
	case e.BoolLit != nil:
		return e.BoolLit.String()
	case e.Placeholder != nil:
		return e.Placeholder.String()
	case e.FieldRef != nil:
		return e.FieldRef.String()
	case e.FuncCall != nil:
		return e.FuncCall.String()
	default:
		return "<invalid expr>"
	}
}

type Null struct {
	Pos scanner.Position
}

func (p *Null) String() string {
	return "null"
}

type Placeholder struct {
	Pos scanner.Position
}

func (p *Placeholder) String() string {
	return "?"
}

type FuncCall struct {
	Pos  scanner.Position
	Name *String
	Args []*Expr
}

func (f *FuncCall) String() string {
	var args []string
	for _, arg := range f.Args {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Name.Value, strings.Join(args, ", "))
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
