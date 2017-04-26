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

type Expr struct {
	Null        bool
	Placeholder bool
	StringLit   *string
	NumberLit   *string
	BoolLit     *bool
	Field       *Field
	FuncCall    *FuncCall
}

func (e *Expr) Nullable() bool {
	switch {
	case e.Null, e.Placeholder:
		return true
	case e.Field != nil:
		return e.Field.Nullable
	case e.FuncCall != nil:
		return e.FuncCall.Nullable()
	default:
		return false
	}
}

func (e *Expr) HasPlaceholder() bool {
	if e.Placeholder {
		return true
	}
	if e.FuncCall != nil {
		return e.FuncCall.HasPlaceholder()
	}
	return false
}

func (e *Expr) Unique() bool {
	return e.Field != nil && e.Field.Unique()
}

type FuncCall struct {
	Name string
	Args []*Expr
}

func (fc *FuncCall) Nullable() bool {
	for _, arg := range fc.Args {
		if arg.Nullable() {
			return true
		}
	}
	return false
}

func (fc *FuncCall) HasPlaceholder() bool {
	for _, arg := range fc.Args {
		if arg.HasPlaceholder() {
			return true
		}
	}
	return false
}
