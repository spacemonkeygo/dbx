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
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
)

type Relation struct {
	Field *Field
	Kind  ast.RelationKind
}

type Field struct {
	Name       string
	Column     string
	Model      *Model
	Type       ast.FieldType
	Relation   *Relation
	Nullable   bool
	AutoInsert bool
	AutoUpdate bool
	Updatable  bool
	Length     int // Text only
}

func (f *Field) Insertable() bool {
	if f.Relation != nil {
		return true
	}
	return f.Type != ast.SerialField && f.Type != ast.Serial64Field
}

func (f *Field) Unique() bool {
	return f.Model.FieldUnique(f)
}

func (f *Field) IsInt() bool {
	switch f.Type {
	case ast.SerialField, ast.Serial64Field, ast.IntField, ast.Int64Field:
		return true
	default:
		return false
	}
}

func (f *Field) IsTime() bool {
	switch f.Type {
	case ast.TimestampField, ast.TimestampUTCField:
		return true
	default:
		return false
	}
}

func (f *Field) ColumnRef() string {
	return fmt.Sprintf("%s.%s", f.Model.Table, f.Column)
}

func (f *Field) UnderRef() string {
	return fmt.Sprintf("%s_%s", f.Model.Name, f.Name)
}

func (f *Field) SelectRefs() (refs []string) {
	return []string{f.ColumnRef()}
}

func (f *Field) selectable() {}
