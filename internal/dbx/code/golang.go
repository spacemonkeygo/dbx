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

package code

import (
	"fmt"
	"go/format"
	"io"
	"text/template"

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/ast"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/sql"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx/templates"
)

var (
	GoError = Error.NewClass("golang")
)

type GolangOptions struct {
	Package string
}

type Golang struct {
	header  *template.Template
	funcs   *template.Template
	options GolangOptions
}

var _ Language = (*Golang)(nil)

func NewGolang(loader templates.Loader, options *GolangOptions) (
	*Golang, error) {

	header, err := loader.Load("golang.header.tmpl")
	if err != nil {
		return nil, err
	}

	funcs, err := loader.Load("golang.funcs.tmpl")
	if err != nil {
		return nil, err
	}

	return &Golang{
		header:  header,
		funcs:   funcs,
		options: *options,
	}, nil
}

func (g *Golang) Format(in []byte) (out []byte, err error) {
	out, err = format.Source(in)
	return out, GoError.Wrap(err)
}

func (g *Golang) RenderHeader(w io.Writer, root *ast.Root,
	dialects []sql.Dialect) error {

	type headerDialect struct {
		Name      string
		Driver    string
		SchemaSQL string
	}

	type headerParams struct {
		Package        string
		Dialects       []headerDialect
		Structs        []*GolangStruct
		StructsReverse []*GolangStruct
	}

	params := headerParams{
		Package: g.options.Package,
		Structs: GolangStructsFromModels(root.Models),
	}

	for i := len(params.Structs) - 1; i >= 0; i-- {
		params.StructsReverse = append(params.StructsReverse, params.Structs[i])
	}

	for _, dialect := range dialects {
		var schema_sql string
		//		schema_sql, err := dialect.RenderSchema(schema)
		//		if err != nil {
		//			return err
		//		}

		var driver string
		switch dialect.Name() {
		case "postgres":
			driver = "github.com/lib/pq"
		case "sqlite3":
			driver = "github.com/mattn/go-sqlite3"
		default:
			return Error.New("unsupported dialect %q", dialect.Name())
		}

		params.Dialects = append(params.Dialects, headerDialect{
			Name:      dialect.Name(),
			Driver:    driver,
			SchemaSQL: schema_sql,
		})
	}

	return templates.Render(g.header, w, "", params)
}

func (g *Golang) RenderInsert(w io.Writer, model *ast.Model,
	dialect sql.Dialect) (err error) {

	go_ins := GolangInsertFromModel(model, dialect, false)
	if err := templates.Render(g.funcs, w, "insert", go_ins); err != nil {
		return err
	}

	go_ins = GolangInsertFromModel(model, dialect, true)
	if err := templates.Render(g.funcs, w, "raw-insert", go_ins); err != nil {
		return err
	}

	return nil
}

func (g *Golang) RenderSelect(w io.Writer, sel *ast.Select,
	dialect sql.Dialect) error {

	go_sel := GolangSelectFromSelect(sel, dialect)
	if err := templates.Render(g.funcs, w, "select", go_sel); err != nil {
		return err
	}

	return nil
}

func (g *Golang) RenderDelete(w io.Writer, del *ast.Delete,
	dialect sql.Dialect) error {

	return nil
}

func (g *Golang) RenderFooter(w io.Writer, root *ast.Root,
	dialects []sql.Dialect) error {
	return nil
}

type GolangInsert struct {
	model   *ast.Model
	s       *GolangStruct
	dialect sql.Dialect
	raw     bool
}

func GolangInsertFromModel(model *ast.Model, dialect sql.Dialect, raw bool) (
	i *GolangInsert) {

	return &GolangInsert{
		model:   model,
		s:       GolangStructFromModel(model),
		dialect: dialect,
		raw:     raw,
	}
}

func (i *GolangInsert) Dialect() string {
	return i.dialect.Name()
}

func (i *GolangInsert) FuncSuffix() string {
	return i.s.Name()
}

func (i *GolangInsert) SQL() string {
	return sql.RenderInsert(i.dialect, i.model)
}

func (i *GolangInsert) Args() (fields []*GolangField) {
	if i.raw {
		return i.Inserts()
	}
	for _, field := range i.Inserts() {
		if !field.field.AutoInsert {
			fields = append(fields, field)
		}
	}
	return fields
}

func (i *GolangInsert) Inserts() (fields []*GolangField) {
	return GolangFieldsFromFields(i.model.InsertableFields())
}

func (i *GolangInsert) Autos() (fields []*GolangField) {
	if i.raw {
		return nil
	}
	for _, field := range i.Inserts() {
		if field.field.AutoInsert {
			fields = append(fields, field)
		}
	}
	return fields
}

func (i *GolangInsert) Struct() string {
	return i.s.Name()
}

func (i *GolangInsert) ReturnBy() (return_by *GolangReturnBy) {
	if i.dialect.SupportsReturning() {
		return nil
	}
	if pk := GolangFieldFromField(i.model.BasicPrimaryKey()); pk != nil {
		return &GolangReturnBy{
			Pk: pk.Name(),
		}
	}
	panic("returnby.getter")
	// TODO: with getter
	return nil
}

func (i *GolangInsert) NeedsNow() bool {
	for _, field := range i.Autos() {
		switch field.field.Type {
		case ast.TimestampField, ast.TimestampUTCField:
			return true
		}
	}
	return false
}

type GolangReturnBy struct {
	Pk     string
	Getter interface{} //*GolangFuncBase
}

type GolangStruct struct {
	model  *ast.Model
	fields []*GolangField
}

func GolangStructFromModel(model *ast.Model) *GolangStruct {
	if model == nil {
		return nil
	}
	s := &GolangStruct{
		model:  model,
		fields: GolangFieldsFromFields(model.Fields),
	}
	for _, field := range s.fields {
		field.gostruct = s
	}

	return s
}

func GolangStructsFromModels(models []*ast.Model) (out []*GolangStruct) {
	for _, model := range models {
		out = append(out, GolangStructFromModel(model))
	}
	return out
}

func (s *GolangStruct) Name() string {
	return inflect.Camelize(s.model.Name)
}

func (s *GolangStruct) Fields() []*GolangField {
	return s.fields
}

func (s *GolangStruct) Updatable() bool {
	return false
}

func (s *GolangStruct) Init() string {
	return fmt.Sprintf("&%s{}", s.Name())
}

func (s *GolangStruct) Arg() string {
	return inflect.Underscore(s.model.Name)
}

func (s *GolangStruct) Param() string {
	return fmt.Sprintf("%s *%s", s.Arg(), s.Name())
}

type GolangField struct {
	field    *ast.Field
	gostruct *GolangStruct
}

func GolangFieldFromField(field *ast.Field) *GolangField {
	if field == nil {
		return nil
	}
	return &GolangField{
		field: field,
	}
}

func GolangFieldsFromFields(fields []*ast.Field) (out []*GolangField) {
	for _, field := range fields {
		out = append(out, GolangFieldFromField(field))
	}
	return out
}

func (f *GolangField) Name() string {
	return inflect.Camelize(f.field.Name)
}

func (f *GolangField) Param() (string, error) {
	param_type, err := f.Type()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s", f.Arg(), param_type), nil
}

func (f *GolangField) Arg() string {
	if f.gostruct != nil {
		return f.field.Model.Name + "." + f.Name()
	} else {
		return inflect.Underscore(f.field.Model.Name + "_" + f.field.Name)
	}
}

func (s *GolangField) Fields() []*GolangField {
	return []*GolangField{s}
}

func (f *GolangField) Type() (string, error) {
	switch f.field.Type {
	case ast.TextField:
		if f.field.Nullable {
			return "sql.NullString", nil
		} else {
			return "string", nil
		}
	case ast.IntField, ast.SerialField:
		if f.field.Nullable {
			return "sql.NullInt64", nil
		} else {
			return "int64", nil
		}
	case ast.UintField:
		if !f.field.Nullable {
			return "uint", nil
		}
	case ast.Int64Field, ast.Serial64Field:
		if f.field.Nullable {
			return "sql.NullInt64", nil
		} else {
			return "int64", nil
		}
	case ast.Uint64Field:
		if !f.field.Nullable {
			return "uint64", nil
		}
	case ast.BlobField:
		return "[]byte", nil
	case ast.TimestampField, ast.TimestampUTCField:
		if f.field.Nullable {
			return "*time.Time", nil
		} else {
			return "time.Time", nil
		}
	case ast.BoolField:
		if f.field.Nullable {
			return "sql.NullBool", nil
		} else {
			return "bool", nil
		}
	}
	return "", GoError.New("unhandled type %q (nullable=%t)", f.field.Type,
		f.field.Nullable)
}

func (f *GolangField) Tag() string {
	return fmt.Sprintf("`"+`db:"%s"`+"`", f.field.Name)
}

func (f *GolangField) Updatable() bool {
	return f.field.Updatable
}

func (f *GolangField) Init() (string, error) {
	field_type, err := f.Type()
	if err != nil {
		return "", err
	}
	switch field_type {
	case "string":
		return `""`, nil
	case "sql.NullString":
		return `sql.NullString{}`, nil
	case "int64", "uint64":
		return `0`, nil
	case "sql.NullInt64":
		return `sql.NullInt64{}`, nil
	case "[]byte":
		return "nil", nil
	case "time.Time":
		return "now", nil
	case "*time.Time":
		return "nil", nil
	case "bool":
		return "false", nil
	case "sql.NullBool":
		return "sql.NullBool{}", nil
	default:
		return "", Error.New("unhandled field init for type %q", field_type)
	}
}

type GolangSelect struct {
	sel     *ast.Select
	dialect sql.Dialect
}

func GolangSelectFromSelect(sel *ast.Select,
	dialect sql.Dialect) *GolangSelect {

	return &GolangSelect{
		sel:     sel,
		dialect: dialect,
	}
}

func (g *GolangSelect) SQL() string {
	return sql.RenderSelect(g.dialect, g.sel)
}

func (g *GolangSelect) Dialect() string {
	return g.dialect.Name()
}

func (g *GolangSelect) FuncSuffix() string {
	return inflect.Camelize(g.sel.FuncSuffix)
}

func (g *GolangSelect) Args() (args []*GolangField) {
	for _, where := range g.sel.Where {
		if where.Right == nil {
			args = append(args, GolangFieldFromField(where.Left))
		}
	}
	return args
}

func (g *GolangSelect) Returns() (returns []interface{}) {
	for _, selectable := range g.sel.Fields {
		switch t := selectable.(type) {
		case *ast.Model:
			returns = append(returns, GolangStructFromModel(t))
		case *ast.Field:
			returns = append(returns, GolangFieldFromField(t))
		default:
			panic(fmt.Sprintf("unhandled selectable type %T", t))
		}
	}
	return returns
}
