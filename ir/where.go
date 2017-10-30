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

type WhereExpr struct {
	Left  *Expr
	Op    consts.Operator
	Right *Expr
}

type WhereOr struct {
	Left  *Where
	Right *Where
}

type WhereAnd struct {
	Left  *Where
	Right *Where
}

type Where struct {
	Expr *WhereExpr
	Or   *WhereOr
	And  *WhereAnd
}

func (w *WhereExpr) NeedsCondition() bool {
	// only EQ and NE need a condition to switch on "=" v.s. "is", etc.
	switch w.Op {
	case consts.EQ, consts.NE:
	default:
		return false
	}

	// null values are fixed and don't need a runtime condition to render
	// appropriately
	if w.Left.Null || w.Right.Null {
		return false
	}

	return w.Left.Nullable() && w.Right.Nullable()
}
