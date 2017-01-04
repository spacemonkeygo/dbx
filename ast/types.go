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

package ast

import (
	"fmt"
	"text/scanner"
)

type Root struct {
	Models  []*Model
	Selects []*Select
}

type Model struct {
	Pos        scanner.Position
	Name       string
	Table      string
	Fields     []*Field
	PrimaryKey *RelativeFieldRefs
	Unique     []*RelativeFieldRefs
	Indexes    []*Index
}

type Field struct {
	Pos        scanner.Position
	Name       string
	Type       FieldType
	Relation   *Relation
	Column     string
	Nullable   bool
	Updatable  bool
	AutoInsert bool
	AutoUpdate bool
	Length     int
}

type FieldType int

const (
	UnsetField FieldType = iota
	SerialField
	Serial64Field
	IntField
	Int64Field
	UintField
	Uint64Field
	FloatField
	Float64Field
	TextField
	BoolField
	TimestampField
	TimestampUTCField
	BlobField
)

func (f FieldType) String() string {
	switch f {
	case UnsetField:
		return "<UNSET-FIELD>"
	case SerialField:
		return "serial"
	case Serial64Field:
		return "serial64"
	case IntField:
		return "int"
	case Int64Field:
		return "int64"
	case UintField:
		return "uint"
	case Uint64Field:
		return "uint64"
	case FloatField:
		return "float"
	case Float64Field:
		return "float64"
	case TextField:
		return "text"
	case BoolField:
		return "bool"
	case TimestampField:
		return "timestamp"
	case TimestampUTCField:
		return "utimestamp"
	case BlobField:
		return "blob"
	default:
		return "<UNKNOWN-FIELD>"
	}
}

func (f FieldType) AsLink() FieldType {
	switch f {
	case SerialField:
		return IntField
	case Serial64Field:
		return Int64Field
	default:
		return f
	}
}

type Relation struct {
	Pos      scanner.Position
	FieldRef *FieldRef
}

type FieldRef struct {
	Pos   scanner.Position
	Model string
	Field string
}

func (r *FieldRef) String() string {
	if r.Field == "" {
		return r.Model
	}
	if r.Model == "" {
		return r.Field
	}
	return fmt.Sprintf("%s.%s", r.Model, r.Field)
}

func (f *FieldRef) Relative() *RelativeFieldRef {
	return &RelativeFieldRef{
		Pos:   f.Pos,
		Field: f.Field,
	}
}

type RelativeFieldRefs struct {
	Pos  scanner.Position
	Refs []*RelativeFieldRef
}

type RelativeFieldRef struct {
	Pos   scanner.Position
	Field string
}

func (r *RelativeFieldRef) String() string { return r.Field }

type Index struct {
	Pos    scanner.Position
	Name   string
	Fields *RelativeFieldRefs
}

type Select struct {
	Pos        scanner.Position
	FuncSuffix string
	Limit      *Limit
	Fields     *FieldRefs
	Joins      []*Join
	Where      []*Where
	OrderBy    *OrderBy
}

type Limit struct {
	Pos    scanner.Position
	Amount int
}

type FieldRefs struct {
	Pos  scanner.Position
	Refs []*FieldRef
}

type Join struct {
	Pos   scanner.Position
	Left  *FieldRef
	Right *FieldRef
	Type  JoinType
}

type JoinType int

const (
	LeftJoin JoinType = iota
)

type Where struct {
	Pos   scanner.Position
	Left  *FieldRef
	Op    Operator
	Right *FieldRef
}

type Operator string

const (
	LT   Operator = "<"
	LE   Operator = "<="
	GT   Operator = ">"
	GE   Operator = ">="
	EQ   Operator = "="
	NE   Operator = "!="
	Like Operator = "like"
)

type OrderBy struct {
	Pos        scanner.Position
	Fields     *FieldRefs
	Descending bool
}
