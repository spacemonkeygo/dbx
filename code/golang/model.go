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

package golang

import (
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

// Struct is used for generating go structures
type ModelStruct struct {
	Name   string
	Table  string
	Fields []*ModelField
}

func ModelStructFromIR(model *ir.Model) *ModelStruct {
	name := structName(model)

	return &ModelStruct{
		Name:   name,
		Table:  model.Table,
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
	return s.Name + "_Update_Fields"
}

func (s *ModelStruct) CreateStructName() string {
	return s.Name + "_Create_Fields"
}

type ModelField struct {
	Name       string
	ModelName  string
	Type       string
	CtorValue  string
	MutateFn   string
	Column     string
	Nullable   bool
	Insertable bool
	AutoInsert bool
	Updatable  bool
	AutoUpdate bool
	TakeAddr   bool
}

func ModelFieldFromIR(field *ir.Field) *ModelField {
	return &ModelField{
		Name:       fieldName(field),
		ModelName:  structName(field.Model),
		Type:       valueType(field.Type, field.Nullable),
		CtorValue:  valueType(field.Type, false),
		MutateFn:   mutateFn(field.Type),
		Column:     field.Column,
		Nullable:   field.Nullable,
		Insertable: true,
		AutoInsert: field.AutoInsert,
		Updatable:  field.Updatable,
		AutoUpdate: field.AutoUpdate,
		TakeAddr:   field.Nullable && field.Type != consts.BlobField,
	}
}

func ModelFieldsFromIR(fields []*ir.Field) (out []*ModelField) {
	for _, field := range fields {
		out = append(out, ModelFieldFromIR(field))
	}
	return out
}

func (f *ModelField) StructName() string {
	return fmt.Sprintf("%s_%s_Field", f.ModelName, f.Name)
}

func (f *ModelField) ArgType() string {
	if f.Nullable {
		return "*" + f.StructName()
	}
	return f.StructName()
}

func valueType(t consts.FieldType, nullable bool) (value_type string) {
	switch t {
	case consts.TextField:
		value_type = "string"
	case consts.IntField, consts.SerialField:
		value_type = "int"
	case consts.UintField:
		value_type = "uint"
	case consts.Int64Field, consts.Serial64Field:
		value_type = "int64"
	case consts.Uint64Field:
		value_type = "uint64"
	case consts.BlobField:
		value_type = "[]byte"
	case consts.TimestampField:
		value_type = "time.Time"
	case consts.TimestampUTCField:
		value_type = "time.Time"
	case consts.BoolField:
		value_type = "bool"
	case consts.FloatField:
		value_type = "float32"
	case consts.Float64Field:
		value_type = "float64"
	case consts.DateField:
		value_type = "time.Time"
	default:
		panic(fmt.Sprintf("unhandled field type %q", t))
	}

	if nullable && t != consts.BlobField {
		return "*" + value_type
	}
	return value_type
}

func zeroVal(t consts.FieldType, nullable bool) string {
	if nullable {
		return "nil"
	}
	switch t {
	case consts.TextField:
		return `""`
	case consts.IntField, consts.SerialField:
		return `int(0)`
	case consts.UintField:
		return `uint(0)`
	case consts.Int64Field, consts.Serial64Field:
		return `int64(0)`
	case consts.Uint64Field:
		return `uint64(0)`
	case consts.BlobField:
		return `nil`
	case consts.TimestampField:
		return `time.Time{}`
	case consts.TimestampUTCField:
		return `time.Time{}`
	case consts.BoolField:
		return `false`
	case consts.FloatField:
		return `float32(0)`
	case consts.Float64Field:
		return `float64(0)`
	case consts.DateField:
		return `time.Time{}`
	default:
		panic(fmt.Sprintf("unhandled field type %q", t))
	}
}

func initVal(t consts.FieldType, nullable bool) string {
	switch t {
	case consts.TextField:
		if nullable {
			return `(*string)(nil)`
		}
		return `""`
	case consts.IntField, consts.SerialField:
		if nullable {
			return `(*int)(nil)`
		}
		return `int(0)`
	case consts.UintField:
		if nullable {
			return `(*uint)(nil)`
		}
		return `uint(0)`
	case consts.Int64Field, consts.Serial64Field:
		if nullable {
			return `(*int64)(nil)`
		}
		return `int64(0)`
	case consts.Uint64Field:
		if nullable {
			return `(*uint64)(nil)`
		}
		return `uint64(0)`
	case consts.BlobField:
		if nullable {
			return `[]byte(nil)`
		}
		return `nil`
	case consts.TimestampField:
		if nullable {
			return `(*time.Time)(nil)`
		}
		return `__now`
	case consts.TimestampUTCField:
		if nullable {
			return `(*time.Time)(nil)`
		}
		return `__now.UTC()`
	case consts.BoolField:
		if nullable {
			return `(*bool)(nil)`
		}
		return `false`
	case consts.FloatField:
		if nullable {
			return `(*float32)(nil)`
		}
		return `float32(0)`
	case consts.Float64Field:
		if nullable {
			return `(*float64)(nil)`
		}
		return `float64(0)`
	case consts.DateField:
		if nullable {
			return `(*time.Time)(nil)`
		}
		return `toDate(__now)`
	default:
		panic(fmt.Sprintf("unhandled field type %q", t))
	}
}

func mutateFn(field_type consts.FieldType) string {
	switch field_type {
	case consts.TimestampUTCField:
		return "toUTC"
	case consts.DateField:
		return "toDate"
	default:
		return ""
	}
}
