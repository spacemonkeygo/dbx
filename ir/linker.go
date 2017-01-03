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

type linker struct {
	models map[string]*modelLink
}

func newLinker() *linker {
	return &linker{
		models: make(map[string]*modelLink),
	}
}

func (l *linker) AddModel(ast_model *ast.Model) (link *modelLink, err error) {
	if existing, ok := l.models[ast_model.Name]; ok {
		return nil, Error.New("%s: model %q already defined at %s",
			ast_model.Pos, ast_model.Name, existing.ast.Pos)
	}

	link = newModelLink(ast_model)
	l.models[ast_model.Name] = link
	return link, nil
}

func (l *linker) GetModel(name string) *modelLink {
	return l.models[name]
}

func (l *linker) FindModel(ref *ast.FieldRef) (*Model, error) {
	link := l.models[ref.Model]
	if link != nil {
		return link.model, nil
	}
	return nil, Error.New("%s: no model %q defined",
		ref.Pos, ref.Model)
}

func (l *linker) FindField(ref *ast.FieldRef) (*Field, error) {
	model_link := l.models[ref.Model]
	if model_link == nil {
		return nil, Error.New("%s: no model %q defined",
			ref.Pos, ref.Model)
	}
	return model_link.FindField(ref.Relative())
}

type modelLink struct {
	model  *Model
	ast    *ast.Model
	fields map[string]*fieldLink
}

func newModelLink(ast_model *ast.Model) *modelLink {
	return &modelLink{
		model:  &Model{},
		ast:    ast_model,
		fields: make(map[string]*fieldLink),
	}
}

func (m *modelLink) newFieldLink(ast_field *ast.Field) *fieldLink {
	field := &Field{
		Model: m.model,
	}
	m.model.Fields = append(m.model.Fields, field)

	return &fieldLink{
		field: field,
		ast:   ast_field,
	}
}

func (m *modelLink) AddField(ast_field *ast.Field) (err error) {
	if existing, ok := m.fields[ast_field.Name]; ok {
		return Error.New("%s: field %q already defined at %s",
			ast_field.Pos, ast_field.Name, existing.ast.Pos)
	}
	m.fields[ast_field.Name] = m.newFieldLink(ast_field)
	return nil
}

func (m *modelLink) GetField(name string) *fieldLink {
	return m.fields[name]
}

func (m *modelLink) FindField(ref *ast.RelativeFieldRef) (*Field, error) {
	field_link := m.fields[ref.Field]
	if field_link == nil {
		return nil, Error.New("%s: no field %q defined on model %q",
			ref.Pos, ref.Field, m.model.Name)
	}
	return field_link.field, nil
}

type fieldLink struct {
	field *Field
	ast   *ast.Field
}
