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
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
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

func parseRoot(scanner *Scanner) (root *ast.Root, err error) {
	root = new(ast.Root)

	for {
		token, pos, text, err := scanner.ScanOneOf(EOF, Ident)
		if err != nil {
			return nil, err
		}
		if token == EOF {
			return root, nil
		}

		switch strings.ToLower(text) {
		case "model":
			model, err := parseModel(scanner)
			if err != nil {
				return nil, err
			}
			root.Models = append(root.Models, model)
		case "create":
			cre, err := parseCreate(scanner)
			if err != nil {
				return nil, err
			}
			root.Creates = append(root.Creates, cre)
		case "read":
			read, err := parseRead(scanner)
			if err != nil {
				return nil, err
			}
			root.Reads = append(root.Reads, read)
		case "update":
			upd, err := parseUpdate(scanner)
			if err != nil {
				return nil, err
			}
			root.Updates = append(root.Updates, upd)
		case "delete":
			del, err := parseDelete(scanner)
			if err != nil {
				return nil, err
			}
			root.Deletes = append(root.Deletes, del)
		default:
			return nil, expectedKeyword(pos, text, "model", "read", "update",
				"delete")
		}
	}
}

func parseModel(scanner *Scanner) (model *ast.Model, err error) {
	model = new(ast.Model)
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
			model.Fields = append(model.Fields, field)
		case "key":
			if model.PrimaryKey != nil {
				return nil, Error.New(
					"%s: primary key already on model %q at %s",
					pos, model.Name, model.PrimaryKey.Pos)
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
			model.Indexes = append(model.Indexes, index)
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

func parseField(scanner *Scanner) (field *ast.Field, err error) {
	field = new(ast.Field)
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
	relation *ast.Relation, err error) {

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

	relation = &ast.Relation{
		FieldRef: &ast.FieldRef{
			Pos:   pos,
			Model: ident,
			Field: suffix,
		},
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

func parseRelativeFieldRefs(scanner *Scanner) (refs *ast.RelativeFieldRefs,
	err error) {

	refs = new(ast.RelativeFieldRefs)
	refs.Pos = scanner.Pos()

	for {
		ref, err := parseRelativeFieldRef(scanner)
		if err != nil {
			return nil, err
		}
		refs.Refs = append(refs.Refs, ref)

		if pos, _, ok := scanner.ScanIf(Comma); !ok {
			if len(refs.Refs) == 0 {
				return nil, Error.New(
					"%s: at ir one field must be specified", pos)
			}
			return refs, nil
		}
	}
}

func parseRelativeFieldRef(scanner *Scanner) (ref *ast.RelativeFieldRef,
	err error) {

	ref = new(ast.RelativeFieldRef)
	ref.Pos, ref.Field, err = scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func parseIndex(scanner *Scanner) (index *ast.Index, err error) {
	index = new(ast.Index)
	index.Pos = scanner.Pos()

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
		case "name":
			if index.Name != "" {
				return nil, Error.New(
					"%s: name can only be defined once", pos)
			}
			_, index.Name, err = scanner.ScanExact(Ident)
			if err != nil {
				return nil, err
			}
			index.Name = strings.ToLower(index.Name)
		case "fields":
			if index.Fields != nil {
				return nil, Error.New(
					"%s: fields already defined on index at %s",
					pos, index.Fields.Pos)
			}
			fields, err := parseRelativeFieldRefs(scanner)
			if err != nil {
				return nil, err
			}
			index.Fields = fields
		case "unique":
			index.Unique = true
		default:
			return nil, expectedKeyword(pos, text, "name", "fields", "unique")
		}
	}

	return index, nil
}

func parseCreate(scanner *Scanner) (cre *ast.Create, err error) {
	cre = new(ast.Create)
	cre.Pos = scanner.Pos()

	cre.Model, err = parseModelRef(scanner)
	if err != nil {
		return nil, err
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

		switch strings.ToLower(text) {
		case "raw":
			if cre.Raw {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			cre.Raw = true
		default:
			return nil, expectedKeyword(pos, text, "raw")
		}
	}

	return cre, nil
}

func parseUpdate(scanner *Scanner) (upd *ast.Update, err error) {
	upd = new(ast.Update)
	upd.Pos = scanner.Pos()

	upd.Model, err = parseModelRef(scanner)
	if err != nil {
		return nil, err
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

		switch strings.ToLower(text) {
		case "where":
			where, err := parseWhere(scanner)
			if err != nil {
				return nil, err
			}
			upd.Where = append(upd.Where, where)
		case "join":
			join, err := parseJoin(scanner)
			if err != nil {
				return nil, err
			}
			upd.Joins = append(upd.Joins, join)
		default:
			return nil, expectedKeyword(pos, text, "where", "join")
		}
	}

	return upd, nil
}

func parseDelete(scanner *Scanner) (del *ast.Delete, err error) {
	del = new(ast.Delete)
	del.Pos = scanner.Pos()

	del.Model, err = parseModelRef(scanner)
	if err != nil {
		return nil, err
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

		switch strings.ToLower(text) {
		case "where":
			where, err := parseWhere(scanner)
			if err != nil {
				return nil, err
			}
			del.Where = append(del.Where, where)
		case "join":
			join, err := parseJoin(scanner)
			if err != nil {
				return nil, err
			}
			del.Joins = append(del.Joins, join)
		default:
			return nil, expectedKeyword(pos, text, "where", "join")
		}
	}

	return del, nil
}

func parseRead(scanner *Scanner) (read *ast.Read, err error) {
	read = new(ast.Read)
	read.Pos = scanner.Pos()

	if _, _, err := scanner.ScanExact(OpenParen); err != nil {
		return nil, err
	}

	for {
		token, pos, text, err := scanner.ScanOneOf(CloseParen, Ident)
		if err != nil {
			return nil, err
		}

		if token == CloseParen {
			return read, nil
		}

		switch text {
		case "select":
			if read.Select != nil {
				return nil, Error.New("%s: select can only be specified once",
					pos)
			}
			read.Select, err = parseFieldRefs(scanner, modelCentricRef)
			if err != nil {
				return nil, err
			}
		case "where":
			where, err := parseWhere(scanner)
			if err != nil {
				return nil, err
			}
			read.Where = append(read.Where, where)
		case "join":
			join, err := parseJoin(scanner)
			if err != nil {
				return nil, err
			}
			read.Joins = append(read.Joins, join)
		case "view":
			if read.View != nil {
				return nil, Error.New("%s: views can only be specified once",
					pos)
			}
			read.View, err = parseView(scanner)
			if err != nil {
				return nil, err
			}
		case "orderby":
			if read.OrderBy != nil {
				return nil, Error.New("%s: orderby can only be specified once",
					pos)
			}
			read.OrderBy, err = parseOrderBy(scanner)
			if err != nil {
				return nil, err
			}
		default:
			return nil, expectedKeyword(pos, text,
				"select", "where", "join", "view", "orderby")
		}
	}
}

func parseView(scanner *Scanner) (view *ast.View, err error) {
	view = new(ast.View)
	view.Pos = scanner.Pos()

	for {
		pos, text, err := scanner.ScanExact(Ident)
		if err != nil {
			return nil, err
		}

		switch strings.ToLower(text) {
		case "all":
			if view.All {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.All = true
		case "limit":
			if view.Limit {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.Limit = true
		case "limitoffset":
			if view.LimitOffset {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.LimitOffset = true
		case "offset":
			if view.Offset {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.Offset = true
		case "paged":
			if view.Paged {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.Paged = true
		case "count":
			if view.Count {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.Count = true
		case "has":
			if view.Has {
				return nil, Error.New("%s: %q already specified", pos, text)
			}
			view.Has = true
		default:
			return nil, expectedKeyword(pos, text, "all", "limit",
				"limitoffset", "offset", "paged")
		}

		if _, _, ok := scanner.ScanIf(Comma); !ok {
			return view, nil
		}
	}
}

func parseOrderBy(scanner *Scanner) (order_by *ast.OrderBy, err error) {
	order_by = new(ast.OrderBy)
	order_by.Pos = scanner.Pos()

	pos, text, err := scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(text) {
	case "asc":
	case "desc":
		order_by.Descending = true
	default:
		return nil, expectedKeyword(pos, text, "asc", "desc")
	}

	order_by.Fields, err = parseFieldRefs(scanner, fullRef)
	if err != nil {
		return nil, err
	}
	return order_by, nil
}

func parseWhere(scanner *Scanner) (where *ast.Where, err error) {
	where = new(ast.Where)
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

	where.Right, err = parseFieldRef(scanner, fullRef)
	if err != nil {
		return nil, err
	}

	return where, nil
}

func parseJoin(scanner *Scanner) (join *ast.Join, err error) {
	join = new(ast.Join)
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
)

func parseFieldRefs(scanner *Scanner, ref_type fieldRefType) (
	refs *ast.FieldRefs, err error) {

	refs = new(ast.FieldRefs)
	refs.Pos = scanner.Pos()

	for {
		ref, err := parseFieldRef(scanner, ref_type)
		if err != nil {
			return nil, err
		}
		refs.Refs = append(refs.Refs, ref)

		if _, _, ok := scanner.ScanIf(Comma); !ok {
			return refs, nil
		}
	}
}

func parseFieldRef(scanner *Scanner, ref_type fieldRefType) (
	ref *ast.FieldRef, err error) {

	ref = new(ast.FieldRef)
	ref.Pos = scanner.Pos()

	_, first, err := scanner.ScanExact(Ident)
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
			ref.Model = first
			return ref, nil
		default:
			return nil, errors.NotImplementedError.New(
				"unhandled ref type %s", ref_type)
		}
	}
	_, second, err := scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}

	ref.Model = first
	ref.Field = second
	return ref, nil
}

func parseModelRef(scanner *Scanner) (ref *ast.ModelRef, err error) {
	ref = new(ast.ModelRef)
	ref.Pos, ref.Model, err = scanner.ScanExact(Ident)
	if err != nil {
		return nil, err
	}
	return ref, nil
}
