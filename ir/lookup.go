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

import "gopkg.in/spacemonkeygo/dbx.v1/ast"

type lookup struct {
	models map[string]*modelEntry
}

func newLookup() *lookup {
	return &lookup{
		models: make(map[string]*modelEntry),
	}
}

func (l *lookup) AddModel(ast_model *ast.Model) (link *modelEntry, err error) {
	if existing, ok := l.models[ast_model.Name]; ok {
		return nil, Error.New("%s: model %q already defined at %s",
			ast_model.Pos, ast_model.Name, existing.ast.Pos)
	}

	link = newModelEntry(ast_model)
	l.models[ast_model.Name] = link
	return link, nil
}

func (l *lookup) GetModel(name string) *modelEntry {
	return l.models[name]
}

func (l *lookup) FindModel(ref *ast.ModelRef) (*Model, error) {
	link := l.models[ref.Model]
	if link != nil {
		return link.model, nil
	}
	return nil, Error.New("%s: no model %q defined",
		ref.Pos, ref.Model)
}

func (l *lookup) FindField(ref *ast.FieldRef) (*Field, error) {
	model_link := l.models[ref.Model]
	if model_link == nil {
		return nil, Error.New("%s: no model %q defined",
			ref.Pos, ref.Model)
	}
	return model_link.FindField(ref.Relative())
}

type modelEntry struct {
	model  *Model
	ast    *ast.Model
	fields map[string]*fieldEntry
}

func newModelEntry(ast_model *ast.Model) *modelEntry {
	return &modelEntry{
		model: &Model{
			Name: ast_model.Name,
		},
		ast:    ast_model,
		fields: make(map[string]*fieldEntry),
	}
}

func (m *modelEntry) newFieldEntry(ast_field *ast.Field) *fieldEntry {
	field := &Field{
		Name:  ast_field.Name,
		Type:  ast_field.Type,
		Model: m.model,
	}
	m.model.Fields = append(m.model.Fields, field)

	return &fieldEntry{
		field: field,
		ast:   ast_field,
	}
}

func (m *modelEntry) AddField(ast_field *ast.Field) (err error) {
	if existing, ok := m.fields[ast_field.Name]; ok {
		return Error.New("%s: field %q already defined at %s",
			ast_field.Pos, ast_field.Name, existing.ast.Pos)
	}
	m.fields[ast_field.Name] = m.newFieldEntry(ast_field)
	return nil
}

func (m *modelEntry) GetField(name string) *fieldEntry {
	return m.fields[name]
}

func (m *modelEntry) FindField(ref *ast.RelativeFieldRef) (*Field, error) {
	field_link := m.fields[ref.Field]
	if field_link == nil {
		return nil, Error.New("%s: no field %q defined on model %q",
			ref.Pos, ref.Field, m.model.Name)
	}
	return field_link.field, nil
}

type fieldEntry struct {
	field *Field
	ast   *ast.Field
}
