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

type Field struct {
	Name       string
	Type       string
	Column     string
	Nullable   bool
	Insertable bool
	AutoInsert bool
	Updatable  bool
	AutoUpdate bool
}

func FieldFromIR(field *ir.Field) *Field {
	return &Field{
		Name:       fieldName(field),
		Type:       fieldType(field),
		Column:     field.ColumnName(),
		Nullable:   field.Nullable,
		Insertable: true,
		AutoInsert: field.AutoInsert,
		Updatable:  field.Updatable,
		AutoUpdate: field.AutoUpdate,
	}
}

func FieldsFromIR(fields []*ir.Field) (out []*Field) {
	for _, field := range fields {
		out = append(out, FieldFromIR(field))
	}
	return out
}

//func (f *Field) Init() string {
//	switch field_type {
//	case "string":
//		return `""`
//	case "sql.NullString":
//		return `sql.NullString{}`
//	case "int64", "uint64":
//		return `0`
//	case "sql.NullInt64":
//		return `sql.NullInt64{}`
//	case "[]byte":
//		return `nil`
//	case "time.Time":
//		return `now`
//	case "*time.Time":
//		return `nil`
//	case "bool":
//		return `false`
//	case "sql.NullBool":
//		return `sql.NullBool{}`
//	default:
//		panic(fmt.Sprintf("unhandled field init for type %q", f.Type)
//	}
//}

//type Field struct {
//	field    *ir.Field
//	gostruct *Struct
//}
//
//func FieldFromIR(field *ir.Field) *Field {
//	if field == nil {
//		return nil
//	}
//	return &Field{
//		field: field,
//	}
//}
//
//func FieldsFromIR(fields []*ir.Field) (out []*Field) {
//	for _, field := range fields {
//		out = append(out, FieldFromIR(field))
//	}
//	return out
//}
//
//func (f *Field) Name() string {
//	return inflect.Camelize(f.field.Name)
//}
//
//func (f *Field) Column() string {
//	return f.field.ColumnName()
//}
//
//func (f *Field) Param() (string, error) {
//	param_type, err := f.Type()
//	if err != nil {
//		return "", err
//	}
//	return fmt.Sprintf("%s %s", f.Arg(), param_type), nil
//}
//
//func (f *Field) Arg() string {
//	if f.gostruct != nil {
//		return f.field.Model.Name + "." + f.Name()
//	} else {
//		return inflect.Underscore(f.field.Model.Name + "_" + f.field.Name)
//	}
//}
//
//func (s *Field) Fields() []*Field {
//	return []*Field{s}
//}
//
//func (f *Field) Type() (string, error) {
//	switch f.field.Type {
//	case ast.TextField:
//		if f.field.Nullable {
//			return "sql.NullString", nil
//		} else {
//			return "string", nil
//		}
//	case ast.IntField, ast.SerialField:
//		if f.field.Nullable {
//			return "sql.NullInt64", nil
//		} else {
//			return "int64", nil
//		}
//	case ast.UintField:
//		if !f.field.Nullable {
//			return "uint", nil
//		}
//	case ast.Int64Field, ast.Serial64Field:
//		if f.field.Nullable {
//			return "sql.NullInt64", nil
//		} else {
//			return "int64", nil
//		}
//	case ast.Uint64Field:
//		if !f.field.Nullable {
//			return "uint64", nil
//		}
//	case ast.BlobField:
//		return "[]byte", nil
//	case ast.TimestampField, ast.TimestampUTCField:
//		if f.field.Nullable {
//			return "*time.Time", nil
//		} else {
//			return "time.Time", nil
//		}
//	case ast.BoolField:
//		if f.field.Nullable {
//			return "sql.NullBool", nil
//		} else {
//			return "bool", nil
//		}
//	}
//	return "", Error.New("unhandled type %q (nullable=%t)", f.field.Type,
//		f.field.Nullable)
//}
//
//func (f *Field) Tag() string {
//	return fmt.Sprintf("`"+`db:"%s"`+"`", f.field.Name)
//}

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
