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

package ir

import "gopkg.in/spacemonkeygo/dbx.v1/consts"

type Where struct {
	Left  *Field
	Op    consts.Operator
	Right *Field
}

func WhereFieldEquals(field *Field) *Where {
	return &Where{
		Left: field,
		Op:   consts.EQ,
	}
}

func WhereFieldsEquals(fields ...*Field) (wheres []*Where) {
	if len(fields) == 0 {
		return nil
	}
	for _, field := range fields {
		wheres = append(wheres, WhereFieldEquals(field))
	}
	return wheres
}

func FilterWhere(wheres []*Where, op consts.Operator) (filtered []*Where) {
	for _, where := range wheres {
		if where.Op == op {
			filtered = append(filtered, where)
		}
	}
	return filtered
}
