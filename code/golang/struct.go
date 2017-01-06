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

import "gopkg.in/spacemonkeygo/dbx.v1/ir"

// Struct is used for generating go structures
type Struct struct {
	Name   string
	Fields []*Field
}

func StructFromIR(model *ir.Model) *Struct {
	return &Struct{
		Name:   structName(model),
		Fields: FieldsFromIR(model.Fields),
	}
}

func StructsFromIR(models []*ir.Model) (out []*Struct) {
	for _, model := range models {
		out = append(out, StructFromIR(model))
	}
	return out
}

func (s *Struct) UpdatableFields() (fields []*Field) {
	for _, field := range s.Fields {
		if field.Updatable {
			fields = append(fields, field)
		}
	}
	return fields
}
