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
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

type Insert struct {
	Suffix            string
	Return            *Var
	Args              []*Var
	Fields            []*Var
	SQL               string
	SupportsReturning bool
	NeedsNow          bool
}

type AutoInit struct {
	NeedsNow bool
	Vars     []*Var
}

func InsertFromIR(ir_ins *ir.Insert, dialect sql.Dialect) *Insert {
	suffix := inflect.Camelize(ir_ins.Model.Name)
	if ir_ins.Raw {
		suffix = "Raw" + suffix
	}

	ins := &Insert{
		Suffix:            suffix,
		Return:            VarFromModel(ir_ins.Model),
		SQL:               sql.RenderInsert(dialect, ir_ins),
		SupportsReturning: dialect.Features().Returning,
	}

	args := map[string]*Var{}

	// All of the manual fields are arguments to the function. The Field struct
	// type is used (pointer if nullable).
	for _, field := range ir_ins.ManualFields() {
		arg_type := fmt.Sprintf("%s%sField",
			inflect.Camelize(field.Model.Name),
			inflect.Camelize(field.Name))
		if field.Nullable {
			arg_type = "*" + arg_type
		}
		arg := &Var{
			Name: field.Name,
			Type: arg_type,
		}
		args[field.Name] = arg
		ins.Args = append(ins.Args, arg)
	}

	// Now for each field
	for _, field := range ir_ins.Fields() {
		v := VarFromField(field)
		v.Name = fmt.Sprintf("__%s_val", v.Name)
		if arg := args[field.Name]; arg != nil {
			v.InitVal = fmt.Sprintf("%s.value()", arg.Name)
		} else if field.IsTime() {
			ins.NeedsNow = true
		}
		ins.Fields = append(ins.Fields, v)
	}

	return ins
}
