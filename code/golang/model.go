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

package golang

import (
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

// Struct is used for generating go structures
type ModelStruct struct {
	Name   string
	Fields []*ModelField
}

func ModelStructFromIR(model *ir.Model) *ModelStruct {
	name := structName(model)

	return &ModelStruct{
		Name:   name,
		Fields: ModelFieldsFromIR(model.Fields),
	}
}

func ModelStructsFromIR(models []*ir.Model) (out []*ModelStruct) {
	for _, model := range models {
		out = append(out, ModelStructFromIR(model))
	}
	return out
}

func (s *ModelStruct) UpdatableFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Updatable && !field.AutoUpdate {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) OptionalInsertFields() (fields []*ModelField) {
	for _, field := range s.Fields {
		if field.Insertable && !field.AutoInsert && field.Nullable {
			fields = append(fields, field)
		}
	}
	return fields
}

func (s *ModelStruct) UpdateStructName() string {
	return "Update" + s.Name
}

func (s *ModelStruct) CreateStructName() string {
	return "Create" + s.Name
}

type ModelField struct {
	Name       string
	ModelName  string
	Type       string
	Column     string
	Nullable   bool
	Insertable bool
	AutoInsert bool
	Updatable  bool
	AutoUpdate bool
}

func ModelFieldFromIR(field *ir.Field) *ModelField {
	return &ModelField{
		Name:       fieldName(field),
		ModelName:  structName(field.Model),
		Type:       fieldType(field),
		Column:     field.ColumnName(),
		Nullable:   field.Nullable,
		Insertable: true,
		AutoInsert: field.AutoInsert,
		Updatable:  field.Updatable,
		AutoUpdate: field.AutoUpdate,
	}
}

func ModelFieldsFromIR(fields []*ir.Field) (out []*ModelField) {
	for _, field := range fields {
		out = append(out, ModelFieldFromIR(field))
	}
	return out
}

func fieldType(field *ir.Field) string {
	switch field.Type {
	case ast.TextField:
		if field.Nullable {
			return "sql.NullString"
		} else {
			return "string"
		}
	case ast.IntField, ast.SerialField:
		if field.Nullable {
			return "sql.NullInt64"
		} else {
			return "int64"
		}
	case ast.UintField:
		if !field.Nullable {
			return "uint"
		}
	case ast.Int64Field, ast.Serial64Field:
		if field.Nullable {
			return "sql.NullInt64"
		} else {
			return "int64"
		}
	case ast.Uint64Field:
		if !field.Nullable {
			return "uint64"
		}
	case ast.BlobField:
		return "[]byte"
	case ast.TimestampField, ast.TimestampUTCField:
		if field.Nullable {
			return "*time.Time"
		} else {
			return "time.Time"
		}
	case ast.BoolField:
		if field.Nullable {
			return "sql.NullBool"
		} else {
			return "bool"
		}
	}
	panic(fmt.Sprintf("unhandled field type %q (nullable=%t)",
		field.Type, field.Nullable))
}

func (f *ModelField) UpdateStructName() string {
	return fmt.Sprintf("%s_%sField", f.ModelName, f.Name)
}
