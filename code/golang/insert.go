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
	ins     *ir.Insert
	dialect sql.Dialect
}

func InsertFromIR(ins *ir.Insert, dialect sql.Dialect) *Insert {
	return &Insert{
		ins:     ins,
		dialect: dialect,
	}
}

func (i *Insert) Dialect() string {
	return i.dialect.Name()
}

func (i *Insert) FuncName() string {
	return inflect.Camelize(i.ins.FuncName())
}

func (i *Insert) SQL() string {
	return sql.RenderInsert(i.dialect, i.ins)
}

func (i *Insert) Args() []*Field {
	return FieldsFromIR(i.ins.ManualFields())
}

func (i *Insert) Fields() []*Field {
	return FieldsFromIR(i.ins.Fields())
}

func (i *Insert) AutoFields() []*Field {
	return FieldsFromIR(i.ins.AutoFields())
}

func (i *Insert) Struct() string {
	return structName(i.ins.Model)
}

func (i *Insert) ReturnBy() (return_by *ReturnBy) {
	if i.dialect.Features().Returning {
		return nil
	}
	if pk := FieldFromIR(i.ins.Model.BasicPrimaryKey()); pk != nil {
		return &ReturnBy{
			Pk: pk.Name,
		}
	}
	panic("returnby.getter")
	// TODO: with getter
	return nil
}

func (i *Insert) NeedsNow() bool {
	for _, field := range i.AutoFields() {
		switch field.Type {
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
