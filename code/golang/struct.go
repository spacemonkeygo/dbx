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

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

type Struct struct {
	Name   string
	Fields []Field
}

func (s *Struct) FieldVars() (vars []*Var) {
	for _, field := range s.Fields {
		vars = append(vars, &Var{
			Name: field.Name,
			Type: field.Type,
		})
	}
	return vars
}

type Field struct {
	Name string
	Type string
	Tags []Tag
}

type Tag struct {
	Key   string
	Value string
}

func FieldFromSelectable(selectable ir.Selectable, full_name bool) Field {
	field := Field{}
	switch obj := selectable.(type) {
	case *ir.Model:
		field.Name = inflect.Camelize(obj.Name)
		field.Type = field.Name
	case *ir.Field:
		if full_name {
			field.Name = inflect.Camelize(obj.Model.Name) + "_" +
				inflect.Camelize(obj.Name)
		} else {
			field.Name = inflect.Camelize(obj.Name)
		}
		field.Type = valueType(obj.Type, obj.Nullable)
	default:
		panic(fmt.Sprintf("unhandled selectable type %T", obj))
	}
	return field
}

func FieldsFromSelectables(selectables []ir.Selectable) (fields []Field) {
	// count the number of models being selected
	nmodels := 0
	for _, selectable := range selectables {
		if _, ok := selectable.(*ir.Model); ok {
			nmodels++
		}
	}

	for _, selectable := range selectables {
		field := FieldFromSelectable(selectable, nmodels > 1)
		fields = append(fields, field)
	}
	return fields
}
