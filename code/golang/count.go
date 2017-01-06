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

type Count struct {
	Suffix string
	Args   []*Var
	SQL    string
}

func CountFromIR(ir_count *ir.Count, dialect sql.Dialect) *Count {
	count := &Count{
		Suffix: inflect.Camelize(ir_count.FuncSuffix),
		SQL:    sql.RenderCount(dialect, ir_count),
	}

	for _, where := range ir_count.Where {
		if where.Right == nil {
			count.Args = append(count.Args, VarFromField(where.Left))
		}
	}

	return count
}
