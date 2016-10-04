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
	Name            string
	Fields          []GolangField
	Updatable       bool
	UpdatableFields []GolangField
}

type GolangField struct {
	Name   string
	Type   string
	Tag    string
	Column string
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

	header_tmpl, err := loader.Load("golang.header.tmpl")
	if err != nil {
		return nil, err
	}

	funcs_tmpl, err := loader.Load("golang.funcs.tmpl")
	if err != nil {
		return nil, err
	}

	return &Golang{
		tmpl:        funcs_tmpl,
		dialect:     dialect,
		header_tmpl: header_tmpl,
		options:     options,
	}, nil
}

func (g *Golang) Name() string {
	return "golang"
}

func (g *Golang) Format(in []byte) (out []byte, err error) {
	out, err = format.Source(in)
	return out, Error.Wrap(err)
}

func (g *Golang) RenderHeader(w io.Writer, schema *dbx.Schema) (err error) {

	type headerParams struct {
		Package string
		Dialect string
		Structs []GolangStruct
	}

	params := headerParams{
		Package: g.options.Package,
		Dialect: g.dialect.Name(),
		Structs: g.structsFromTables(schema.Tables),
	}

	return dbx.RenderTemplate(g.header_tmpl, w, "", params)
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
	Name   string
	Init   string
	Column string
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

type GolangUpdate struct {
	GolangFunc
	ReturnBy *string
	NeedsNow bool
	Autos    []GolangAutoParam
}

func (g *Golang) RenderSelect(w io.Writer, sql string,
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

	return dbx.RenderTemplate(g.tmpl, w, tmpl, GolangSelect{
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

func (g *Golang) RenderCount(w io.Writer, sql string,
	params *dbx.SelectParams) (err error) {

	return dbx.RenderTemplate(g.tmpl, w, "count",
		MakeGolangFunc(
			params.Table,
			sql,
			GolangArgsFromConditions(params.Conditions),
		))
}

func (g *Golang) RenderDelete(w io.Writer, sql string,
	params *dbx.DeleteParams) error {

	tmpl := "delete"
	if params.Many {
		tmpl = "delete-all"
	}

	return dbx.RenderTemplate(g.tmpl, w, tmpl,
		MakeGolangFunc(
			params.Table,
			sql,
			GolangArgsFromConditions(params.Conditions),
		))
}

func (g *Golang) RenderInsert(w io.Writer, sql string,
	params *dbx.InsertParams) error {

	var all []GolangArg
	var args []GolangArg
	var autos []GolangAutoParam
	var needs_now bool
	for _, column := range params.Columns {
		arg := GolangArgFromColumn(column)
		all = append(all, arg)
		if column.AutoInsert {
			if init := GolangFieldInit(column); init != "" {
				if init == "now" {
					needs_now = true
				}
				autos = append(autos, GolangAutoParam{
					Name:   arg.Name,
					Init:   init,
					Column: g.dialect.ColumnName(column),
				})
			} else {
				args = append(args, arg)
			}
		} else {
			args = append(args, arg)
		}
	}

	insert := GolangInsert{
		GolangFunc: MakeGolangFunc(params.Table, sql, args),
		Inserts:    all,
		Autos:      autos,
		NeedsNow:   needs_now,
	}

	if g.dialect.SupportsReturning() {
		return dbx.RenderTemplate(g.tmpl, w, "insert", insert)
	} else if pk := params.Table.BasicPrimaryKey(); pk != nil {
		by := GolangFieldName(pk)
		insert.ReturnBy = &by
		return dbx.RenderTemplate(g.tmpl, w, "insert", insert)
	} else {
		return dbx.RenderTemplate(g.tmpl, w, "insert-no-return", insert)
	}
}

func (g *Golang) RenderUpdate(w io.Writer, sql string,
	params *dbx.UpdateParams) error {

	// For updates, the only thing we need to know is which fields to
	// auto generate. The rest will be passed in via ColumnUpdate interfaces
	// passed variadically to the update function.
	var autos []GolangAutoParam
	var needs_now bool
	for _, column := range params.Table.Columns {
		if !column.AutoUpdate {
			continue
		}
		arg := GolangArgFromColumn(column)
		if init := GolangFieldInit(column); init != "" {
			if init == "now" {
				needs_now = true
			}
			autos = append(autos, GolangAutoParam{
				Name:   arg.Name,
				Init:   init,
				Column: g.dialect.ColumnName(column),
			})
		}
	}

	update := GolangUpdate{
		GolangFunc: MakeGolangFunc(
			params.Table,
			sql,
			GolangArgsFromConditions(params.Conditions),
		),
		Autos:    autos,
		NeedsNow: needs_now,
	}

	if g.dialect.SupportsReturning() {
		return dbx.RenderTemplate(g.tmpl, w, "update", update)
	} else if pk := params.Table.BasicPrimaryKey(); pk != nil {
		by := GolangFieldName(pk)
		update.ReturnBy = &by
		return dbx.RenderTemplate(g.tmpl, w, "update", update)
	} else {
		return dbx.RenderTemplate(g.tmpl, w, "update-no-return", update)
	}
}

func (g *Golang) fieldsFromColumns(columns []*dbx.Column, updatabe_only bool) (
	fields []GolangField) {

	for _, column := range columns {
		if !updatabe_only || column.Updatable && !column.AutoUpdate {
			fields = append(fields, g.fieldFromColumn(column))
		}
	}
	return fields
}

func (g *Golang) fieldFromColumn(column *dbx.Column) GolangField {
	return GolangField{
		Name:   GolangFieldName(column),
		Type:   GolangFieldType(column),
		Tag:    GolangFieldTag(column),
		Column: g.dialect.ColumnName(column),
	}
}

func (g *Golang) structFromTable(table *dbx.Table) GolangStruct {
	return GolangStruct{
		Name:            GolangStructName(table),
		Fields:          g.fieldsFromColumns(table.Columns, false),
		Updatable:       table.Updatable(),
		UpdatableFields: g.fieldsFromColumns(table.Columns, true),
	}
}

func (g *Golang) structsFromTables(tables []*dbx.Table) (structs []GolangStruct) {
	for _, table := range tables {
		structs = append(structs, g.structFromTable(table))
	}
	return structs
}
