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

package ir

import "gopkg.in/spacemonkeygo/dbx.v1/consts"

type Selectable interface {
	SelectRefs() []string
	UnderRef() string
	ModelOf() *Model
	selectable()
}

type Read struct {
	Suffix      string
	Selectables []Selectable
	From        *Model
	Joins       []*Join
	Where       []*Where
	OrderBy     *OrderBy
	View        View
}

func (r *Read) One() bool {
	var targets []*Model
	for _, selectable := range r.Selectables {
		targets = append(targets, selectable.ModelOf())
	}
	return queryUnique(distinctModels(targets), r.Joins, r.Where)
}

type View int

const (
	All View = iota
	LimitOffset
	Paged
	Count
	Has
)

type Join struct {
	Type  consts.JoinType
	Left  *Field
	Right *Field
}

type OrderBy struct {
	Fields     []*Field
	Descending bool
}
