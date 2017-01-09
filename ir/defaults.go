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

import "bitbucket.org/pkg/inflect"

func (root *Root) SetDefaults() {
	for _, model := range root.Models {
		model.SetDefaults()
	}
	for _, cre := range root.Creates {
		cre.SetDefaults()
	}
	for _, read := range root.Reads {
		read.SetDefaults()
	}
	for _, upd := range root.Updates {
		upd.SetDefaults()
	}
	for _, del := range root.Deletes {
		del.SetDefaults()
	}
}

func (model *Model) SetDefaults() {
	if model.Table == "" {
		model.Table = inflect.Pluralize(model.Name)
	}

	for _, field := range model.Fields {
		field.SetDefaults()
	}
}

func (field *Field) SetDefaults() {
	if field.Column == "" {
		field.Column = field.Name
	}
}

func (cre *Create) SetDefaults() {
}

func (read *Read) SetDefaults() {
}

func (udp *Update) SetDefaults() {
}

func (del *Delete) SetDefaults() {
}
