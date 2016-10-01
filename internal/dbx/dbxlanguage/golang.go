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

package dbxlanguage

import (
	"fmt"
	"go/format"
	"io"
	"text/template"

	"bitbucket.org/pkg/inflect"
	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx"
)

var (
	Error = errors.NewClass("dbxlang")
)

func GolangStructName(table *dbx.Table) string {
	if table == nil {
		return ""
	}
	return inflect.Camelize(inflect.Singularize(table.Name))
}

func GolangFieldName(column *dbx.Column) string {
	if column == nil {
		return ""
	}
	return inflect.Camelize(column.Name)
}

func GolangFieldType(column *dbx.Column) string {
	if column == nil {
		return ""
	}
	switch column.Type {
	case "text":
		if column.NotNull {
			return "string"
		} else {
			return "sql.NullString"
		}
	case "int", "serial":
		if column.NotNull {
			return "int64"
		} else {
			return "sql.NullInt64"
		}
	case "int64", "serial64":
		if column.NotNull {
			return "int64"
		} else {
			return "sql.NullInt64"
		}
	case "blob":
		return "[]byte"
	case "timestamp":
		if column.NotNull {
			return "time.Time"
		} else {
			return "*time.Time"
		}
	case "bool":
		if column.NotNull {
			return "bool"
		} else {
			return "sql.NullBool"
		}
	}
	panic(fmt.Sprintf("unhandled column type %q", column.Type))
}

func GolangFieldInit(column *dbx.Column) string {
	field_type := GolangFieldType(column)
	switch field_type {
	case "string":
		return `""`
	case "sql.NullString":
		return `sql.NullString{}`
	case "int64":
		return `0`
	case "sql.NullInt64":
		return `sql.NullInt64{}`
	case "[]byte":
		return "nil"
	case "time.Time":
		return "now"
	case "*time.Time":
		return "nil"
	case "bool":
		return "false"
	case "sql.NullBool":
		return "sql.NullBool{}"
	}
	panic(fmt.Sprintf("unhandled field init for type %q", field_type))
}

func GolangFieldTag(column *dbx.Column) string {
	return fmt.Sprintf("`"+`db:"%s"`+"`", column.Name)
}

type GolangStruct struct {
	Name   string
	Fields []GolangField
}

func GolangStructFromTable(table *dbx.Table) GolangStruct {
	return GolangStruct{
		Name:   GolangStructName(table),
		Fields: GolangFieldsFromColumns(table.Columns),
	}
}

func GolangStructsFromTables(tables []*dbx.Table) (structs []GolangStruct) {
	for _, table := range tables {
		structs = append(structs, GolangStructFromTable(table))
	}
	return structs
}

type GolangField struct {
	Name string
	Type string
	Tag  string
}

func GolangFieldFromColumn(column *dbx.Column) GolangField {
	return GolangField{Name: GolangFieldName(column),
		Type: GolangFieldType(column),
		Tag:  GolangFieldTag(column),
	}
}

func GolangFieldsFromColumns(columns []*dbx.Column) (fields []GolangField) {
	for _, column := range columns {
		fields = append(fields, GolangFieldFromColumn(column))
	}
	return fields
}

type GolangOptions struct {
	Package string
}

type Golang struct {
	dialect     dbx.Dialect
	options     *GolangOptions
	header_tmpl *template.Template
	tmpl        *template.Template
}

func NewGolang(loader dbx.Loader, dialect dbx.Dialect, options *GolangOptions) (
	*Golang, error) {

	tmpl, err := loader.Load("golang.tmpl")
	if err != nil {
		return nil, err
	}

	header_tmpl, err := loader.Load("golang.header.tmpl")
	if err != nil {
		return nil, err
	}

	return &Golang{
		tmpl:        tmpl,
		dialect:     dialect,
		header_tmpl: header_tmpl,
		options:     options,
	}, nil
}

func (lang *Golang) Name() string {
	return "golang"
}

func (lang *Golang) Format(in []byte) (out []byte, err error) {
	out, err = format.Source(in)
	return out, Error.Wrap(err)
}

func (lang *Golang) RenderHeader(w io.Writer, schema *dbx.Schema) (err error) {

	type headerParams struct {
		Package string
		Dialect string
		Structs []GolangStruct
	}

	params := headerParams{
		Package: lang.options.Package,
		Dialect: lang.dialect.Name(),
		Structs: GolangStructsFromTables(schema.Tables),
	}

	return dbx.RenderTemplate(lang.header_tmpl, w, "", params)
}

func GolangArgsFromConditions(conditions []*dbx.ConditionParams) (
	out []GolangArg) {

	for _, condition := range conditions {
		if params := condition.ColumnCmp; params != nil {
			out = append(out, GolangArgFromColumn(params.Left))
		}
		if params := condition.ColumnCmpColumn; params != nil {
			out = append(out, GolangArgFromColumn(params.Left))
			out = append(out, GolangArgFromColumn(params.Right))
		}
		if params := condition.ColumnIn; params != nil {
			out = append(out, GolangArgsFromConditions(params.In.Conditions)...)
		}
	}
	return out
}

func GolangArgName(column *dbx.Column) string {
	return inflect.Underscore(column.Table.Name + "_" + column.Name)
}

func GolangArgFromColumn(column *dbx.Column) GolangArg {
	return GolangArg{
		Name:        GolangArgName(column),
		Type:        GolangFieldType(column),
		from_column: column,
	}
}

func GolangFuncSuffix(table *dbx.Table, args []GolangArg) (suffix string) {
	if len(args) == 0 {
		return ""
	}
	// see if all args come from the same table
	args_table := args[0].from_column.Table
	suffix = inflect.Camelize(args[0].from_column.Name)
	for _, arg := range args[1:] {
		if arg.from_column.Table != args_table {
			args_table = nil
			break
		}
		suffix += inflect.Camelize(arg.from_column.Name)
	}

	if args_table != table {
		suffix = ""
		for _, arg := range args {
			suffix += inflect.Camelize(arg.Name)
		}
	}
	return "By" + suffix
}

type GolangAutoParam struct {
	Name string
	Init string
}

func GolangInsertParamsFromTable(columns []*dbx.Column) (
	all, params []GolangArg, autos []GolangAutoParam) {

	for _, column := range columns {
		param := GolangArgFromColumn(column)
		all = append(all, param)
		if column.AutoInsert {
			init := GolangFieldInit(column)
			if init != "" {
				autos = append(autos, GolangAutoParam{
					Name: param.Name,
					Init: init,
				})
			} else {
				params = append(params, param)
			}
		} else {
			params = append(params, param)
		}
	}
	return all, params, autos
}

type GolangArg struct {
	Name        string
	Type        string
	from_column *dbx.Column
}

type GolangFunc struct {
	Struct     string
	SQL        string
	FuncSuffix string
	Args       []GolangArg
}

type GolangSelect struct {
	GolangFunc
	PagedOn string
}

type GolangInsert struct {
	GolangFunc
	ReturnBy *string
	NeedsNow bool
	Inserts  []GolangArg
	Autos    []GolangAutoParam
}

func (lang *Golang) RenderSelect(w io.Writer, sql string,
	params *dbx.SelectParams) (err error) {

	var tmpl string
	switch {
	case !params.Many:
		tmpl = "select"
	case params.PagedOn == nil:
		tmpl = "select-all"
	default:
		tmpl = "select-paged"
	}

	return dbx.RenderTemplate(lang.tmpl, w, tmpl, GolangSelect{
		GolangFunc: MakeGolangFunc(
			params.Table, sql,
			GolangArgsFromConditions(params.Conditions),
		),
		PagedOn: GolangFieldName(params.PagedOn),
	})
}

func MakeGolangFunc(table *dbx.Table, sql string, args []GolangArg) GolangFunc {
	return GolangFunc{
		Struct:     GolangStructName(table),
		SQL:        sql,
		Args:       args,
		FuncSuffix: GolangFuncSuffix(table, args),
	}
}

func (lang *Golang) RenderCount(w io.Writer, sql string,
	params *dbx.SelectParams) (err error) {

	return dbx.RenderTemplate(lang.tmpl, w, "count",
		MakeGolangFunc(
			params.Table,
			sql,
			GolangArgsFromConditions(params.Conditions),
		))
}

func (lang *Golang) RenderDelete(w io.Writer, sql string,
	params *dbx.DeleteParams) error {

	tmpl := "delete"
	if params.Many {
		tmpl = "delete-all"
	}

	return dbx.RenderTemplate(lang.tmpl, w, tmpl,
		MakeGolangFunc(
			params.Table,
			sql,
			GolangArgsFromConditions(params.Conditions),
		))
}

func (lang *Golang) RenderInsert(w io.Writer, sql string,
	params *dbx.InsertParams) error {

	all, args, autos := GolangInsertParamsFromTable(params.Columns)

	insert := GolangInsert{
		GolangFunc: MakeGolangFunc(params.Table, sql, args),
		Inserts:    all,
		Autos:      autos,
	}

	for _, auto := range autos {
		if auto.Init == "now" {
			insert.NeedsNow = true
		}
	}

	if lang.dialect.InsertReturns() {
		return dbx.RenderTemplate(lang.tmpl, w, "insert", insert)
	} else if pk := params.Table.BasicPrimaryKey(); pk != nil {
		by := GolangFieldName(pk)
		insert.ReturnBy = &by
		return dbx.RenderTemplate(lang.tmpl, w, "insert", insert)
	} else {
		return dbx.RenderTemplate(lang.tmpl, w, "insert-no-return", insert)
	}
}
