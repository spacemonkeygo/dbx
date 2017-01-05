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

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func VarFromModel(model *ir.Model) *Var {
	return &Var{
		Name:   model.Name,
		Type:   structName(model),
		Fields: VarsFromFields(model.Fields),
	}
}

func VarFromField(field *ir.Field) *Var {
	return &Var{
		Name: field.Name,
		Type: "int64",
	}
}

func VarsFromFields(fields []*ir.Field) (vars []*Var) {
	for _, field := range fields {
		vars = append(vars, VarFromField(field))
	}
	return vars
}

type Var struct {
	Name   string
	Type   string
	Fields []*Var
}

func (v *Var) Value() string {
	return v.Name
}

func (v *Var) Arg() string {
	return v.Name
}

func (v *Var) Ptr() string {
	return fmt.Sprintf("&%s", v.Name)
}

func (v *Var) Param() string {
	if v.IsStruct() {
		return fmt.Sprintf("%s *%s", v.Name, v.Type)
	}
	return fmt.Sprintf("%s %s", v.Name, v.Type)
}

func (v *Var) Init() string {
	if v.IsStruct() {
		return fmt.Sprintf("%s = &%s{}", v.Name, v.Type)
	}
	var val string
	switch v.Type {
	case "int", "int64", "uint", "uint64", "float", "float64":
		val = "0"
	case "string":
		val = `""`
	case "bool":
		val = "false"
	case "time.Time":
		val = "time.Time{}"
	case "*time.Time":
		val = "nil"
	default:
		panic(fmt.Sprintf("unhandled var type %q", v.Type))
	}
	return fmt.Sprintf("%s = %s", v.Name, val)
}

func (v *Var) IsStruct() bool {
	return len(v.Fields) > 0
}

func (v *Var) Zero() string {
	if v.IsStruct() {
		return "nil"
	}
	switch v.Type {
	case "int", "int64", "uint", "uint64", "float", "float64":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "time.Time":
		return "now"
	case "*time.Time":
		return "&now"
	default:
		panic(fmt.Sprintf("unhandled var type %q", v.Type))
	}
}

func (v *Var) Flatten() (flattened []*Var) {
	if len(v.Fields) == 0 {
		// return a copy
		copy := *v
		return append(flattened, &copy)
	}

	for _, field := range v.Fields {
		field_vars := field.Flatten()
		for _, field_var := range field_vars {
			field_var.Name = v.Name + "." + inflect.Camelize(field_var.Name)
		}
		flattened = append(flattened, field_vars...)
	}
	return flattened
}
