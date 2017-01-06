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
	v := &Var{
		Name:   model.Name,
		Type:   structName(model),
		Fields: VarsFromFields(model.Fields),
	}
	v.InitVal = fmt.Sprintf("&%s{}", v.Type)
	v.ZeroVal = "nil"
	return v
}

func VarFromField(field *ir.Field) *Var {
	v := &Var{
		Name: field.Name,
		Type: fieldType(field),
	}

	// zero val
	switch v.Type {
	case "int", "int64", "uint", "uint64", "float", "float64":
		v.ZeroVal = "0"
	case "string":
		v.ZeroVal = `""`
	case "sql.NullString":
		v.ZeroVal = `sql.NullString{}`
	case "bool":
		v.ZeroVal = "false"
	case "time.Time":
		v.ZeroVal = "time.Time{}"
	case "*time.Time":
		v.ZeroVal = "nil"
	default:
		panic(fmt.Sprintf("unhandled var type %q", v.Type))
	}

	// init val
	switch v.Type {
	case "time.Time":
		v.InitVal = "__now"
	case "*time.Time":
		v.InitVal = "&__now"
	}
	return v
}

func VarsFromFields(fields []*ir.Field) (vars []*Var) {
	for _, field := range fields {
		vars = append(vars, VarFromField(field))
	}
	return vars
}

type Var struct {
	Name    string
	Type    string
	ZeroVal string
	InitVal string
	Fields  []*Var
}

func (v *Var) Value() string {
	return v.Name
}

func (v *Var) Arg() string {
	return v.Name
}

func (v *Var) Init() string {
	return fmt.Sprintf("%s = %s", v.Name, v.initVal())
}

func (v *Var) InitNew() string {
	return fmt.Sprintf("%s := %s", v.Name, v.initVal())
}

func (v *Var) initVal() string {
	val := v.InitVal
	if val == "" {
		val = v.ZeroVal
	}
	return val
}

func (v *Var) Zero() string {
	return v.ZeroVal
}

func (v *Var) AddrOf() string {
	return fmt.Sprintf("&%s", v.Name)
}

func (v *Var) Param() string {
	if v.IsStruct() {
		return fmt.Sprintf("%s *%s", v.Name, v.Type)
	}
	return fmt.Sprintf("%s %s", v.Name, v.Type)
}

func (v *Var) IsStruct() bool {
	return len(v.Fields) > 0
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
