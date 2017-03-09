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

type Model struct {
	Name       string
	Table      string
	Fields     []*Field
	PrimaryKey []*Field
	Unique     [][]*Field
	Indexes    []*Index
}

func (m *Model) BasicPrimaryKey() *Field {
	if len(m.PrimaryKey) == 1 && m.PrimaryKey[0].IsInt() {
		return m.PrimaryKey[0]
	}
	return nil
}

func (m *Model) InsertableFields() (fields []*Field) {
	for _, field := range m.Fields {
		if field.Insertable() && !field.AutoInsert {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Model) AutoInsertableFields() (fields []*Field) {
	for _, field := range m.Fields {
		if field.Insertable() && field.AutoInsert {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Model) UpdatableFields() (fields []*Field) {
	for _, field := range m.Fields {
		if field.Updatable && !field.AutoUpdate {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Model) AutoUpdatableFields() (fields []*Field) {
	for _, field := range m.Fields {
		if field.Updatable && field.AutoUpdate {
			fields = append(fields, field)
		}
	}
	return fields
}

func (m *Model) FieldUnique(field *Field) bool {
	return m.FieldSetUnique([]*Field{field})
}

func (m *Model) FieldSetUnique(fields []*Field) bool {
	if fieldSetSubset(m.PrimaryKey, fields) {
		return true
	}
	for _, unique := range m.Unique {
		if fieldSetSubset(unique, fields) {
			return true
		}
	}
	return false
}

func (m *Model) ModelOf() *Model {
	return m
}

func (m *Model) UnderRef() string {
	return m.Name
}

func (m *Model) SelectRefs() (refs []string) {
	for _, field := range m.Fields {
		refs = append(refs, field.SelectRefs()...)
	}
	return refs
}

func (m *Model) selectable() {}
