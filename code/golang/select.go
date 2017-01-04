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

type Select struct {
	sel     *ir.Select
	dialect sql.Dialect
}

func SelectFromIR(sel *ir.Select,
	dialect sql.Dialect) *Select {

	return &Select{
		sel:     sel,
		dialect: dialect,
	}
}

func (g *Select) SQL() string {
	return sql.RenderSelect(g.dialect, g.sel)
}

func (g *Select) Dialect() string {
	return g.dialect.Name()
}

func (g *Select) FuncSuffix() string {
	return inflect.Camelize(g.sel.FuncSuffix)
}

func (g *Select) Args() (args []*Field) {
	for _, where := range g.sel.Where {
		if where.Right == nil {
			args = append(args, FieldFromIR(where.Left))
		}
	}
	return args
}

func (g *Select) Returns() (returns []interface{}) {
	for _, selectable := range g.sel.Fields {
		switch t := selectable.(type) {
		case *ir.Model:
			returns = append(returns, StructFromIR(t))
		case *ir.Field:
			returns = append(returns, FieldFromIR(t))
		default:
			panic(fmt.Sprintf("unhandled selectable type %T", t))
		}
	}
	return returns
}
