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

type Insert struct {
	Model *Model
	Raw   bool
}

func (ins *Insert) Fields() []*Field {
	if ins.Raw {
		return ins.Model.Fields
	}
	return ins.Model.InsertableFields()
}

func (ins *Insert) ManualFields() (fields []*Field) {
	if ins.Raw {
		return ins.Fields()
	}
	for _, field := range ins.Fields() {
		if !field.AutoInsert {
			fields = append(fields, field)
		}
	}
	return fields
}

func (ins *Insert) AutoFields() (fields []*Field) {
	if ins.Raw {
		return nil
	}
	for _, field := range ins.Fields() {
		if field.AutoInsert {
			fields = append(fields, field)
		}
	}
	return fields
}
