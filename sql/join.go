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

package sql

import (
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

type Join struct {
	Type  string
	Table string
	Left  string
	Right string
}

func JoinsFromIR(ir_joins []*ir.Join) (joins []Join) {
	for _, ir_join := range ir_joins {
		join := Join{
			Table: ir_join.Right.Model.Table,
			Left:  ir_join.Left.ColumnRef(),
			Right: ir_join.Right.ColumnRef(),
		}
		switch ir_join.Type {
		case consts.InnerJoin:
		default:
			panic(fmt.Sprintf("unhandled join type %q", join.Type))
		}
		joins = append(joins, join)
	}
	return joins
}
