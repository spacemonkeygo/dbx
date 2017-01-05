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
	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sql"
)

type Insert struct {
	Suffix            string
	Return            *Var
	Args              []*Var
	Fields            []*Var
	AutoFields        []*Var
	SQL               string
	SupportsReturning bool
}

func InsertFromIR(ir_ins *ir.Insert, dialect sql.Dialect) *Insert {
	suffix := inflect.Camelize(ir_ins.Model.Name)
	if ir_ins.Raw {
		suffix = "Raw" + suffix
	}
	return &Insert{
		Suffix:            suffix,
		Return:            VarFromModel(ir_ins.Model),
		SQL:               sql.RenderInsert(dialect, ir_ins),
		Args:              VarsFromFields(ir_ins.ManualFields()),
		Fields:            VarsFromFields(ir_ins.Fields()),
		AutoFields:        VarsFromFields(ir_ins.AutoFields()),
		SupportsReturning: dialect.Features().Returning,
	}
}

func (i *Insert) NeedsNow() bool {
	for _, v := range i.AutoFields {
		switch v.Type {
		case "time.Time", "*time.Time":
			return true
		}
	}
	return false
}

type ReturnBy struct {
	Pk     string
	Getter interface{} //*FuncBase
}
