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

	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

type RawCreate struct {
	Suffix            string
	Return            *Var
	Arg               *Var
	Fields            []*Var
	SQL               string
	SupportsReturning bool
}

func RawCreateFromIR(ir_cre *ir.Create, dialect sql.Dialect) *RawCreate {
	ins := &RawCreate{
		Suffix:            convertSuffix(ir_cre.Suffix),
		Return:            VarFromModel(ir_cre.Model),
		SQL:               sql.RenderInsert(dialect, ir_cre),
		SupportsReturning: dialect.Features().Returning,
	}

	// the model struct is the only arg.
	ins.Arg = VarFromModel(ir_cre.Model)
	ins.Arg.Name = "raw_" + ins.Arg.Name

	// each field in the model is initialized from the raw model struct.
	for _, field := range ir_cre.Fields() {
		f := ModelFieldFromIR(field)
		v := VarFromField(field)
		if field.Nullable {
			v.InitVal = fmt.Sprintf("%s_%s_Raw(%s.%s).value()",
				ins.Arg.Type, f.Name, ins.Arg.Name, f.Name)
		} else {
			v.InitVal = fmt.Sprintf("%s_%s(%s.%s).value()",
				ins.Arg.Type, f.Name, ins.Arg.Name, f.Name)
		}
		v.Name = fmt.Sprintf("__%s_val", v.Name)
		ins.Fields = append(ins.Fields, v)
	}

	return ins
}

type Create struct {
	Suffix            string
	Return            *Var
	Args              []*Var
	Fields            []*Var
	SQL               string
	SupportsReturning bool
	NeedsNow          bool
}

func CreateFromIR(ir_cre *ir.Create, dialect sql.Dialect) *Create {
	ins := &Create{
		Suffix:            convertSuffix(ir_cre.Suffix),
		Return:            VarFromModel(ir_cre.Model),
		SQL:               sql.RenderInsert(dialect, ir_cre),
		SupportsReturning: dialect.Features().Returning,
	}

	args := map[string]*Var{}

	// All of the manual fields are arguments to the function. The Field struct
	// type is used (pointer if nullable).
	has_nullable := false
	for _, field := range ir_cre.InsertableFields() {
		arg := ArgFromField(field)
		args[field.Name] = arg
		if !field.Nullable {
			ins.Args = append(ins.Args, arg)
		} else {
			has_nullable = true
		}
	}

	if has_nullable {
		ins.Args = append(ins.Args, &Var{
			Name: "optional",
			Type: ModelStructFromIR(ir_cre.Model).CreateStructName(),
		})
	}

	// Now for each field
	for _, field := range ir_cre.Fields() {
		if field == ir_cre.Model.BasicPrimaryKey() {
			continue
		}
		v := VarFromField(field)
		v.Name = fmt.Sprintf("__%s_val", v.Name)
		if arg := args[field.Name]; arg != nil {
			if field.Nullable {
				f := ModelFieldFromIR(field)
				v.InitVal = fmt.Sprintf("optional.%s.value()", f.Name)
			} else {
				v.InitVal = fmt.Sprintf("%s.value()", arg.Name)
			}
		} else if field.IsTime() {
			ins.NeedsNow = true
		}
		ins.Fields = append(ins.Fields, v)
	}

	return ins
}
