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

import "fmt"

type Update struct {
	Suffix []string
	Model  *Model
	Joins  []*Join
	Where  []*Where
}

func (r *Update) Signature() string {
	return fmt.Sprintf("UPDATE(%q)", r.Suffix)
}

func (upd *Update) AutoUpdatableFields() (fields []*Field) {
	return upd.Model.AutoUpdatableFields()
}

func (upd *Update) One() bool {
	return queryUnique([]*Model{upd.Model}, upd.Joins, upd.Where)
}
