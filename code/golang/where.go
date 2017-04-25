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

import "gopkg.in/spacemonkeygo/dbx.v1/ir"

type PartitionedArgs struct {
	AllArgs      []*Var
	StaticArgs   []*Var
	NullableArgs []*Var
}

func PartitionedArgsFromWheres(wheres []*ir.Where) (out PartitionedArgs) {
	for _, where := range wheres {
		if !where.Right.HasPlaceholder() {
			continue
		}

		arg := ArgFromWhere(where)
		out.AllArgs = append(out.AllArgs, arg)

		if where.NeedsCondition() {
			out.NullableArgs = append(out.NullableArgs, arg)
		} else {
			out.StaticArgs = append(out.StaticArgs, arg)
		}
	}
	return out
}
