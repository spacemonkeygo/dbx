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

package parser

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/scanner"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/ast"
)

func ParseFile(path string) (root *ast.Root, err error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	scanner, err := NewScanner(path, data)
	if err != nil {
		return nil, err
	}
	return parseRoot(scanner)
}

func Parse(data []byte) (root *ast.Root, err error) {
	scanner, err := NewScanner("", data)
	if err != nil {
		return nil, err
	}
	return parseRoot(scanner)
}

type parserRoot struct {
	Models  []*parserModel
	Selects []*parserSelect
	Deletes []*parserDelete

	ast    *ast.Root
	models map[string]*parserModel
}

func newParserRoot() *parserRoot {
	return &parserRoot{
		ast:    &ast.Root{},
		models: map[string]*parserModel{},
	}
}

func (r *parserRoot) AddModel(model *parserModel) (err error) {
	if existing := r.models[model.Name]; existing != nil {
		return Error.New("%s: model %q already defined at %s",
			model.Pos, existing.Pos)
	}

	r.models[model.Name] = model
	r.Models = append(r.Models, model)
	return nil
}

func (r *parserRoot) Model(ref *parserFieldRef) (*parserModel, error) {
	model := r.models[ref.Model]
	if model != nil {
		return model, nil
	}
	return nil, Error.New("%s: no model %q defined",
		ref.Pos, ref.Model)
}

func (r *parserRoot) Field(ref *parserFieldRef) (*parserField, error) {
	model, err := r.Model(ref)
	if err != nil {
		return nil, err
	}
	return model.Field(ref)
}

func (r *parserRoot) FinalizeAST() (err error) {
	for _, model := range r.Models {
		if err = model.FinalizeAST(r); err != nil {
			return err
		}
		r.ast.Models = append(r.ast.Models, model.ast)
	}
	for _, sel := range r.Selects {
		if err = sel.FinalizeAST(r); err != nil {
			return err
		}
		r.ast.Selects = append(r.ast.Selects, sel.ast)
	}
	for _, del := range r.Deletes {
		if err = del.FinalizeAST(r); err != nil {
			return err
		}
		r.ast.Deletes = append(r.ast.Deletes, del.ast)
	}
	return nil
}

type parserModel struct {
	Pos        scanner.Position
	Name       string
	Table      string
	Fields     []*parserField
	PrimaryKey []*parserFieldRef
	Unique     [][]*parserFieldRef
	Indexes    []*parserIndex

	ast     *ast.Model
	fields  map[string]*parserField
	indexes map[string]*parserIndex
}

func newParserModel() *parserModel {
	return &parserModel{
		ast:     &ast.Model{},
		fields:  map[string]*parserField{},
		indexes: map[string]*parserIndex{},
	}
}

func (m *parserModel) AddField(field *parserField) (err error) {
	if existing := m.fields[field.Name]; existing != nil {
		return Error.New("%s: field %q already defined at %s",
			field.Pos, existing.Pos)
	}
	m.fields[field.Name] = field
	m.Fields = append(m.Fields, field)
	return nil
}

func (m *parserModel) Field(ref *parserFieldRef) (*parserField, error) {
	field := m.fields[ref.Field]
	if field != nil {
		return field, nil
	}
	return nil, Error.New("%s: no field %q on model %q",
		ref.Pos, ref.Field, m.Name)
}

func (m *parserModel) AddIndex(index *parserIndex) (err error) {
	if existing := m.indexes[index.Name]; existing != nil {
		return Error.New("%s: index %q already defined at %s",
			index.Pos, existing.Pos)
	}
	m.indexes[index.Name] = index
	m.Indexes = append(m.Indexes, index)
	return nil
}

func (m *parserModel) Index(name string) *parserIndex {
	return m.indexes[name]
}

func (m *parserModel) FinalizeAST(root *parserRoot) (err error) {
	m.ast.Name = m.Name
	m.ast.Table = m.Table

	for _, field := range m.Fields {
		if err = field.FinalizeAST(root, m); err != nil {
			return err
		}
		m.ast.Fields = append(m.ast.Fields, field.ast)
	}

	if len(m.PrimaryKey) == 0 {
		return Error.New("%s: no primary key defined", m.Pos)
	}

	for _, fieldref := range m.PrimaryKey {
		if field, err := m.Field(fieldref); err == nil {
			m.ast.PrimaryKey = append(m.ast.PrimaryKey, field.ast)
		} else {
			return err
		}
	}

	for _, unique := range m.Unique {
		var fields []*ast.Field
		for _, name := range unique {
			if field, err := m.Field(name); err == nil {
				fields = append(fields, field.ast)
			} else {
				return err
			}
		}
		m.ast.Unique = append(m.ast.Unique, fields)
	}

	for _, index := range m.Indexes {
		if err = index.FinalizeAST(m); err != nil {
			return err
		}
		m.ast.Indexes = append(m.ast.Indexes, index.ast)
	}

	return nil
}

type parserField struct {
	Pos        scanner.Position
	Name       string
	Type       ast.FieldType
	Relation   *parserRelation
	Column     string
	Nullable   bool
	Updatable  bool
	AutoInsert bool
	AutoUpdate bool
	Length     int

	ast *ast.Field
}

func newParserField() *parserField {
	return &parserField{
		ast: &ast.Field{},
	}
}

func (f *parserField) FinalizeAST(root *parserRoot, model *parserModel) (
	err error) {

	f.ast.Name = f.Name
	f.ast.Type = f.Type
	f.ast.Model = model.ast
	if f.Relation != nil {
		if err = f.Relation.FinalizeAST(root); err != nil {
			return err
		}
		f.ast.Relation = f.Relation.ast
		f.ast.Type = f.ast.Relation.Field.Type.AsLink()
	}
	f.ast.Column = f.Column
	f.ast.Nullable = f.Nullable
	f.ast.Updatable = f.Updatable
	f.ast.AutoInsert = f.AutoInsert
	f.ast.AutoUpdate = f.AutoUpdate
	f.ast.Length = f.Length
	return nil
}

type parserRelation struct {
	Pos      scanner.Position
	FieldRef *parserFieldRef

	ast *ast.Relation
}

func newParserRelation() *parserRelation {
	return &parserRelation{
		ast: &ast.Relation{},
	}
}

func (r *parserRelation) FinalizeAST(root *parserRoot) (err error) {
	field, err := root.Field(r.FieldRef)
	if err != nil {
		return err
	}

	r.ast.Field = field.ast
	return nil
}

type parserIndex struct {
	Pos    scanner.Position
	Name   string
	Fields []*parserFieldRef

	ast *ast.Index
}

func newParserIndex() *parserIndex {
	return &parserIndex{
		ast: &ast.Index{},
	}
}

func (idx *parserIndex) FinalizeAST(model *parserModel) (err error) {
	idx.ast.Name = idx.Name
	for _, ref := range idx.Fields {
		if field, err := model.Field(ref); err == nil {
			idx.ast.Fields = append(idx.ast.Fields, field.ast)
		} else {
			return err
		}
	}
	return nil
}

type parserFields struct {
	Pos  scanner.Position
	Refs []*parserFieldRef
}

type parserSelect struct {
	Pos     scanner.Position
	Fields  *parserFields
	Joins   []*parserJoin
	Where   []*parserWhere
	OrderBy *parserOrderBy

	ast *ast.Select
}

func newParserSelect() *parserSelect {
	return &parserSelect{
		ast: &ast.Select{},
	}
}

func (s *parserSelect) FinalizeAST(root *parserRoot) (err error) {

	var func_suffix []string
	if s.Fields == nil || len(s.Fields.Refs) == 0 {
		return Error.New("%s: no fields defined to select", s.Pos)
	}

	// Figure out which models are needed for the fields and that the field
	// references aren't repetetive.
	selected := map[string]map[string]*parserFieldRef{}
	for _, fieldref := range s.Fields.Refs {
		fields := selected[fieldref.Model]
		if fields == nil {
			fields = map[string]*parserFieldRef{}
			selected[fieldref.Model] = fields
		}

		existing := fields[""]
		if existing == nil {
			existing = fields[fieldref.Field]
		}
		if existing != nil {
			return Error.New(
				"%s: field %s already selected by field %s",
				fieldref.Pos, fieldref, existing)
		}
		fields[fieldref.Field] = fieldref

		if fieldref.Field == "" {
			model, err := root.Model(fieldref)
			if err != nil {
				return err
			}
			s.ast.Fields = append(s.ast.Fields, model.ast)
			func_suffix = append(func_suffix, fieldref.Model)
		} else {
			field, err := root.Field(fieldref)
			if err != nil {
				return err
			}
			s.ast.Fields = append(s.ast.Fields, field.ast)
			func_suffix = append(func_suffix, fieldref.Model, fieldref.Field)
		}
	}

	// Figure out set of models that are included in the select. These come from
	// explicit joins, or implicitly if there is only a single model referenced
	// in the fields.
	models := map[string]*parserFieldRef{}
	switch {
	case len(s.Joins) > 0:
		next := s.Joins[0].Left.Model
		for _, join := range s.Joins {
			left, err := root.Field(join.Left)
			if err != nil {
				return err
			}
			if join.Left.Model != next {
				return Error.New(
					"%s: model order must be consistent; expected %q; got %q",
					left.Pos, next, join.Left.Model)
			}
			right, err := root.Field(join.Right)
			if err != nil {
				return err
			}
			next = join.Right.Model
			if s.ast.From == nil {
				s.ast.From = left.ast.Model
				models[join.Left.Model] = join.Left
			}
			s.ast.Joins = append(s.ast.Joins, &ast.Join{
				Type:  join.Type,
				Left:  left.ast,
				Right: right.ast,
			})
			if existing := models[join.Right.Model]; existing != nil {
				return Error.New("%s: model %q already joined at %s",
					join.Right.Pos, join.Right.Model, existing.Pos)
			}
			models[join.Right.Model] = join.Right
		}
	case len(selected) == 1:
		from, err := root.Model(s.Fields.Refs[0])
		if err != nil {
			return err
		}
		s.ast.From = from.ast
		models[from.Name] = s.Fields.Refs[0]
	default:
		return Error.New(
			"%s: cannot select from multiple models without a join",
			s.Fields.Pos)
	}

	// Make sure all of the fields are accounted for in the set of models
	for _, fieldref := range s.Fields.Refs {
		if models[fieldref.Model] == nil {
			return Error.New(
				"%s: cannot select field/model %q; model %q is not joined",
				fieldref.Pos, fieldref, fieldref.Model)
		}
	}

	// Finalize the where conditions and make sure referenced models are part
	// of the select.
	if len(s.Where) > 0 {
		func_suffix = append(func_suffix, "by")
	}
	for _, where := range s.Where {
		if err = where.FinalizeAST(root); err != nil {
			return err
		}
		if models[where.Left.Model] == nil {
			return Error.New(
				"%s: invalid where condition %q; model %q is not joined",
				where.Pos, where, where.Left.Model)
		}
		if where.Right != nil {
			if models[where.Right.Model] == nil {
				return Error.New(
					"%s: invalid where condition %q; model %q is not joined",
					where.Pos, where, where.Right.Model)
			}
		} else {
			func_suffix = append(func_suffix, where.Left.Model, where.Left.Field)
		}
		s.ast.Where = append(s.ast.Where, where.ast)
	}

	// Finalize OrderBy and make sure referenced fields are part of the select
	if s.OrderBy != nil {
		if err = s.OrderBy.FinalizeAST(root); err != nil {
			return err
		}
		for _, order_by_field := range s.OrderBy.Fields {
			if models[order_by_field.Model] == nil {
				return Error.New(
					"%s: invalid orderby field %q; model %q is not joined",
					order_by_field.Pos, order_by_field, order_by_field.Model)
			}
		}
		s.ast.OrderBy = s.OrderBy.ast
	}

	if s.ast.FuncSuffix == "" {
		s.ast.FuncSuffix = strings.Join(func_suffix, "_")
	}

	return nil
}

type parserJoin struct {
	Pos   scanner.Position
	Left  *parserFieldRef
	Right *parserFieldRef
	Type  ast.JoinType
}

type parserWhere struct {
	Pos   scanner.Position
	Left  *parserFieldRef
	Op    ast.Operator
	Right *parserFieldRef

	ast *ast.Where
}

func newParserWhere() *parserWhere {
	return &parserWhere{
		ast: &ast.Where{},
	}
}

func (s *parserWhere) FinalizeAST(root *parserRoot) (err error) {
	s.ast.Op = s.Op
	left, err := root.Field(s.Left)
	if err != nil {
		return err
	}
	s.ast.Left = left.ast
	if s.Right != nil {
		right, err := root.Field(s.Right)
		if err != nil {
			return err
		}
		s.ast.Right = right.ast
	}
	return nil
}

type parserOrderBy struct {
	Fields []*parserFieldRef
	ast    *ast.OrderBy
}

func newParserOrderBy() *parserOrderBy {
	return &parserOrderBy{
		ast: &ast.OrderBy{},
	}
}

func (s *parserOrderBy) FinalizeAST(root *parserRoot) (err error) {
	for _, order_by_field := range s.Fields {
		field, err := root.Field(order_by_field)
		if err != nil {
			return err
		}
		s.ast.Fields = append(s.ast.Fields, field.ast)
	}
	return nil
}

type parserDelete struct {
	Pos   scanner.Position
	Name  string
	Model string
	Where []*parserWhere

	ast *ast.Delete
}

func newParserDelete() *parserDelete {
	return &parserDelete{
		ast: &ast.Delete{},
	}
}

func (s *parserDelete) FinalizeAST(root *parserRoot) (err error) {
	s.ast.Name = s.Name
	s.ast.Model = s.Model

	for _, where := range s.Where {
		if err := where.FinalizeAST(root); err != nil {
			return err
		}
		s.ast.Where = append(s.ast.Where, where.ast)
	}
	return nil
}

type parserFieldRef struct {
	Pos   scanner.Position
	Model string
	Field string
}

func (r *parserFieldRef) String() string {
	if r.Field == "" {
		return r.Model
	}
	if r.Model == "" {
		return r.Field
	}
	return fmt.Sprintf("%s.%s", r.Model, r.Field)
}

func parseRoot(scanner *Scanner) (astroot *ast.Root, err error) {
	root := newParserRoot()
	for {
		token, pos, text, err := scanner.ScanOneOf(EOF, Ident)
		if err != nil {
			return nil, err
		}
		if token == EOF {
			if err = root.FinalizeAST(); err != nil {
				return nil, err
			}
			return root.ast, nil
		}

		switch strings.ToLower(text) {
		case "model":
			model, err := parseModel(scanner)
			if err != nil {
				return nil, err
			}

			if err = root.AddModel(model); err != nil {
				return nil, err
			}
		case "select":
			sel, err := parseSelect(scanner)
			if err != nil {
				return nil, err
			}

			root.Selects = append(root.Selects, sel)
		case "delete":
			del, err := parseDelete(scanner)
			if err != nil {
				return nil, err
			}
			root.Deletes = append(root.Deletes, del)
		default:
			return nil, expectedKeyword(pos, text, "model", "query")
		}
	}
}

func parseModel(scanner *Scanner) (model *parserModel, err error) {
	model = newParserModel()
	model.Pos, model.Name, err = scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	model.Name = strings.ToLower(model.Name)

	_, _, err = scanner.ScanExact(OpenParen)
	if err != nil {
		return nil, err
	}

	for {
		token, pos, text, err := scanner.ScanOneOf(CloseParen, Ident)
		if err != nil {
			return nil, err
		}
		if token == CloseParen {
			break
		}
		switch strings.ToLower(text) {
		case "table":
			_, model.Table, err = scanner.ScanExact(Ident)
			if err != nil {
				return nil, err
			}
		case "field":
			field, err := parseField(scanner)
			if err != nil {
				return nil, err
			}
			if err = model.AddField(field); err != nil {
				return nil, err
			}
		case "key":
			if model.PrimaryKey != nil {
				return nil, Error.New("%s: primary key already on model %q",
					pos, model.Name)
			}
			primary_key, err := parseRelativeFieldRefs(scanner)
			if err != nil {
				return nil, err
			}
			model.PrimaryKey = primary_key
		case "unique":
			unique, err := parseRelativeFieldRefs(scanner)
			if err != nil {
				return nil, err
			}
			model.Unique = append(model.Unique, unique)
		case "index":
			index, err := parseIndex(scanner)
			if err != nil {
				return nil, err
			}
			if err = model.AddIndex(index); err != nil {
				return nil, err
			}
		default:
			return nil, expectedKeyword(pos, text, "name", "field", "key",
				"unique", "index")
		}

	}

	return model, nil
}

var podFields = map[ast.FieldType]bool{
	ast.IntField:          true,
	ast.Int64Field:        true,
	ast.UintField:         true,
	ast.Uint64Field:       true,
	ast.BoolField:         true,
	ast.TextField:         true,
	ast.TimestampField:    true,
	ast.TimestampUTCField: true,
	ast.FloatField:        true,
	ast.Float64Field:      true,
}

var allowedAttributes = map[string]map[ast.FieldType]bool{
	"column":     nil,
	"nullable":   nil,
	"updatable":  nil,
	"default":    podFields,
	"sqldefault": podFields,
	"autoinsert": podFields,
	"autoupdate": podFields,
	"length": {
		ast.TextField: true,
	},
}

func parseField(scanner *Scanner) (field *parserField, err error) {
	field = newParserField()
	field.Pos, field.Name, err = scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	field.Name = strings.ToLower(field.Name)

	field.Type, field.Relation, err = parseFieldType(scanner)
	if err != nil {
		return nil, err
	}

	if _, _, ok := scanner.ScanIf(OpenParen); !ok {
		return field, nil
	}

	for {
		token, pos, raw_attr, err := scanner.ScanOneOf(Ident, CloseParen)
		if err != nil {
			return nil, err
		}
		if token == CloseParen {
			break
		}
		attr := strings.ToLower(raw_attr)

		// make sure the attribute is allowed for the field type
		allowed := allowedAttributes[attr]
		if allowed != nil {
			if !allowed[field.Type] {
				return nil, unallowedAttribute(pos, field.Type, raw_attr)
			}
		}

		switch attr {
		case "column":
			field.Column, err = parseAttribute(scanner)
		case "nullable":
			field.Nullable = true
		case "default":
			//field.Default, err = parseAttribute(scanner)
			_, err = parseAttribute(scanner)
			if err != nil {
				return nil, err
			}
		case "sqldefault":
			//field.SQLDefault, err = parseAttribute(scanner)
			_, err = parseAttribute(scanner)
			if err != nil {
				return nil, err
			}
		case "autoinsert":
			field.AutoInsert = true
		case "autoupdate":
			field.AutoUpdate = true
		case "updatable":
			field.Updatable = true
		case "length":
			field.Length, err = parseIntAttribute(scanner)
			if err != nil {
				return nil, err
			}
		case "large":
		default:
			return nil, expectedKeyword(pos, attr, "")
		}
	}
	return field, nil
}

func parseFieldType(scanner *Scanner) (field_type ast.FieldType,
	relation *parserRelation, err error) {

	pos, ident, err := scanner.ScanExact(Ident)
	if err != nil {
		return ast.UnsetField, nil, err
	}
	ident = strings.ToLower(ident)

	if _, _, ok := scanner.ScanIf(Dot); !ok {
		switch ident {
		case "serial":
			return ast.SerialField, nil, nil
		case "serial64":
			return ast.Serial64Field, nil, nil
		case "int":
			return ast.IntField, nil, nil
		case "int64":
			return ast.Int64Field, nil, nil
		case "uint":
			return ast.UintField, nil, nil
		case "uint64":
			return ast.Uint64Field, nil, nil
		case "bool":
			return ast.BoolField, nil, nil
		case "text":
			return ast.TextField, nil, nil
		case "timestamp":
			return ast.TimestampField, nil, nil
		case "utimestamp":
			return ast.TimestampUTCField, nil, nil
		case "float":
			return ast.FloatField, nil, nil
		case "float64":
			return ast.Float64Field, nil, nil
		case "blob":
			return ast.BlobField, nil, nil
		default:
			return ast.UnsetField, nil, expectedKeyword(pos, ident,
				"serial64", "int", "int64", "uint", "uint64", "bool", "text",
				"timestamp", "utimestamp", "float", "float64", "blob")
		}
	}

	_, suffix, err := scanner.ScanExact(Ident)
	if err != nil {
		return ast.UnsetField, nil, err
	}
	suffix = strings.ToLower(suffix)

	relation = newParserRelation()
	relation.FieldRef = &parserFieldRef{
		Pos:   pos,
		Model: ident,
		Field: suffix,
	}
	return ast.UnsetField, relation, nil
}

func parseAttribute(scanner *Scanner) (string, error) {
	_, _, err := scanner.ScanExact(Colon)
	if err != nil {
		return "", err
	}
	_, text, err := scanner.ScanExact(Ident)
	if err != nil {
		return "", err
	}
	return text, nil
}

func parseIntAttribute(scanner *Scanner) (int, error) {
	_, _, err := scanner.ScanExact(Colon)
	if err != nil {
		return 0, err
	}
	pos, text, err := scanner.ScanExact(Int)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, Error.New("unable to parse int at %s: %v", pos, err)
	}
	return value, nil
}

func parseRelativeFieldRefs(scanner *Scanner) (refs []*parserFieldRef,
	err error) {

	for {
		ref, err := parseRelativeFieldRef(scanner)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)

		if pos, _, ok := scanner.ScanIf(Comma); !ok {
			if len(refs) == 0 {
				return nil, Error.New(
					"%s: at least one field must be specified", pos)
			}
			return refs, nil
		}
	}
}

func parseRelativeFieldRef(scanner *Scanner) (ref *parserFieldRef, err error) {
	ref = &parserFieldRef{}
	ref.Pos, ref.Field, err = scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func parseIndex(scanner *Scanner) (index *parserIndex, err error) {
	index = newParserIndex()
	index.Pos = scanner.Pos()
	if scanner.Peek() == Ident {
		_, index.Name, err = scanner.ScanExact(Ident)
		if err != nil {
			return nil, err
		}
		index.Name = strings.ToLower(index.Name)
	}

	_, _, err = scanner.ScanExact(OpenParen)
	if err != nil {
		return nil, err
	}

	for {
		token, pos, text, err := scanner.ScanOneOf(CloseParen, Ident)
		if err != nil {
			return nil, err
		}
		if token == CloseParen {
			break
		}
		if strings.ToLower(text) != "field" {
			return nil, expectedKeyword(pos, text, "field")
		}

		field, err := parseRelativeFieldRef(scanner)
		if err != nil {
			return nil, err
		}
		index.Fields = append(index.Fields, field)
	}

	return index, nil
}

func parseSelect(scanner *Scanner) (sel *parserSelect, err error) {
	sel = newParserSelect()
	sel.Pos = scanner.Pos()

	if _, _, err := scanner.ScanExact(OpenParen); err != nil {
		return nil, err
	}

	for {
		token, pos, text, err := scanner.ScanOneOf(CloseParen, Ident)
		if err != nil {
			return nil, err
		}

		if token == CloseParen {
			return sel, nil
		}

		switch text {
		case "suffix":
			if sel.ast.FuncSuffix != "" {
				return nil, Error.New("%s: suffix can only be specified once",
					pos)
			}
			_, sel.ast.FuncSuffix, err = scanner.ScanExact(Ident)
			if err != nil {
				return nil, err
			}
		case "fields":
			if sel.Fields != nil {
				return nil, Error.New("%s: fields can only be specified once",
					pos)
			}
			sel.Fields = &parserFields{
				Pos: pos,
			}
			sel.Fields.Refs, err = parseFieldRefs(scanner, modelCentricRef)
			if err != nil {
				return nil, err
			}
		case "where":
			where, err := parseWhere(scanner)
			if err != nil {
				return nil, err
			}
			sel.Where = append(sel.Where, where)
		case "join":
			join, err := parseJoin(scanner)
			if err != nil {
				return nil, err
			}
			sel.Joins = append(sel.Joins, join)
		case "limit":
			sel.ast.Limit, err = parseLimit(scanner)
			if err != nil {
				return nil, err
			}
		case "orderby":
			if sel.OrderBy != nil {
				return nil, Error.New("%s: orderby can only be specified once",
					pos)
			}
			sel.OrderBy, err = parseOrderBy(scanner)
			if err != nil {
				return nil, err
			}
		default:
			return nil, expectedKeyword(pos, text,
				"fields", "where", "join", "limit")
		}
	}
}

func parseLimit(scanner *Scanner) (limit *ast.Limit, err error) {
	limit = new(ast.Limit)
	if _, _, ok := scanner.ScanIf(Question); ok {
		return limit, nil
	}
	_, text, err := scanner.ScanExact(Int)
	if err != nil {
		return nil, err
	}

	limit.Amount, err = strconv.Atoi(text)
	if err != nil {
		return nil, Error.Wrap(err)
	}
	if limit.Amount <= 0 {
		return nil, Error.New("limit amount must be greater than zero")
	}
	return limit, nil
}

func parseOrderBy(scanner *Scanner) (order_by *parserOrderBy, err error) {
	order_by = newParserOrderBy()
	pos, text, err := scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(text) {
	case "asc":
	case "desc":
		order_by.ast.Descending = true
	default:
		return nil, expectedKeyword(pos, text, "asc", "desc")
	}

	order_by.Fields, err = parseFieldRefs(scanner, fullRef)
	if err != nil {
		return nil, err
	}
	return order_by, nil
}

func parseDelete(scanner *Scanner) (del *parserDelete, err error) {
	del = newParserDelete()
	del.Pos = scanner.Pos()

	_, del.Name, _ = scanner.ScanIf(Ident)
	del.Name = strings.ToLower(del.Name)

	if _, _, ok := scanner.ScanIf(OpenParen); !ok {
		return nil, nil
	}

	for {
		token, _, text, err := scanner.ScanOneOf(CloseParen, Ident)
		if err != nil {
			return nil, err
		}

		if token == CloseParen {
			return del, nil
		}

		switch text {
		case "model":
			_, del.Model, err = scanner.ScanExact(Ident)
			if err != nil {
				return nil, err
			}
		case "where":
			where, err := parseWhere(scanner)
			if err != nil {
				return nil, err
			}
			del.Where = append(del.Where, where)
		}
	}
}

func parseWhere(scanner *Scanner) (where *parserWhere, err error) {
	where = newParserWhere()
	where.Pos = scanner.Pos()

	where.Left, err = parseFieldRef(scanner, fullRef)
	if err != nil {
		return nil, err
	}

	token, pos, text, err := scanner.ScanOneOf(Ident, Equal, LeftAngle,
		RightAngle)
	if err != nil {
		return nil, err
	}

	switch token {
	case Exclamation:
		_, _, err := scanner.ScanExact(Equal)
		if err != nil {
			return nil, err
		}
		where.Op = ast.NE
	case Ident:
		switch strings.ToLower(text) {
		case "like":
			where.Op = ast.Like
		default:
			return nil, expectedKeyword(pos, text, "like")
		}
	case Equal:
		where.Op = ast.EQ
	case LeftAngle:
		if _, _, eq := scanner.ScanIf(Equal); eq {
			where.Op = ast.LE
		} else {
			where.Op = ast.LT
		}
	case RightAngle:
		if _, _, eq := scanner.ScanIf(Equal); eq {
			where.Op = ast.GE
		} else {
			where.Op = ast.GT
		}
	}

	if _, _, ok := scanner.ScanIf(Question); ok {
		return where, nil
	}

	where.Right, err = parseFieldRef(scanner, fieldCentricRef)
	if err != nil {
		return nil, err
	}

	return where, nil
}

func parseJoin(scanner *Scanner) (join *parserJoin, err error) {
	join = &parserJoin{}
	join.Pos = scanner.Pos()

	pos, join_type, err := scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}

	switch join_type {
	case "left":
		join.Type = ast.LeftJoin
	default:
		return nil, expectedKeyword(pos, join_type, "left")
	}

	join.Left, err = parseFieldRef(scanner, fullRef)
	if err != nil {
		return nil, err
	}

	_, _, err = scanner.ScanExact(Equal)
	if err != nil {
		return nil, err
	}

	join.Right, err = parseFieldRef(scanner, fullRef)
	if err != nil {
		return nil, err
	}

	return join, nil
}

type fieldRefType int

const (
	fullRef fieldRefType = iota
	modelCentricRef
	fieldCentricRef
)

func parseFieldRefs(scanner *Scanner, ref_type fieldRefType) (
	refs []*parserFieldRef, err error) {

	for {
		ref, err := parseFieldRef(scanner, ref_type)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)

		if _, _, ok := scanner.ScanIf(Comma); !ok {
			return refs, nil
		}
	}
}

func parseFieldRef(scanner *Scanner, ref_type fieldRefType) (
	ref *parserFieldRef, err error) {

	pos, first, err := scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}

	var full bool
	if ref_type == fullRef {
		_, _, err := scanner.ScanExact(Dot)
		if err != nil {
			return nil, err
		}
		full = true
	} else {
		_, _, full = scanner.ScanIf(Dot)
	}

	if !full {
		switch ref_type {
		case modelCentricRef:
			return &parserFieldRef{
				Pos:   pos,
				Model: first,
			}, nil
		case fieldCentricRef:
			return &parserFieldRef{
				Pos:   pos,
				Field: first,
			}, nil
		default:
			return nil, errors.NotImplementedError.New(
				"unhandled ref type %s", ref_type)
		}
	}
	_, second, err := scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}

	return &parserFieldRef{
		Pos:   pos,
		Model: first,
		Field: second,
	}, nil
}
