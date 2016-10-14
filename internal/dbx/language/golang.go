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

package language

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"text/template"

	"bitbucket.org/pkg/inflect"
	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx"
)

var (
	Error = errors.NewClass("language")
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
	options     *GolangOptions
	header_tmpl *template.Template
	footer_tmpl *template.Template
	tmpl        *template.Template
	funcs       []string
}

func NewGolang(loader dbx.Loader, options *GolangOptions) (*Golang, error) {
	header_tmpl, err := loader.Load("golang.header.tmpl")
	if err != nil {
		return nil, err
	}

	funcs_tmpl, err := loader.Load("golang.funcs.tmpl")
	if err != nil {
		return nil, err
	}

	footer_tmpl, err := loader.Load("golang.footer.tmpl")
	if err != nil {
		return nil, err
	}

	return &Golang{
		tmpl:        funcs_tmpl,
		header_tmpl: header_tmpl,
		footer_tmpl: footer_tmpl,
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

func (g *Golang) RenderHeader(w io.Writer, dialects []dbx.Dialect,
	schema *dbx.Schema) (err error) {

	type headerDialect struct {
		Name      string
		Driver    string
		SchemaSQL string
	}

	type headerParams struct {
		Package        string
		Dialects       []headerDialect
		Structs        []GolangStruct
		StructsReverse []GolangStruct
	}

	params := headerParams{
		Package: g.options.Package,
		Structs: g.structsFromTables(schema.Tables),
	}

	for i := len(params.Structs) - 1; i >= 0; i-- {
		params.StructsReverse = append(params.StructsReverse, params.Structs[i])
	}

	for _, dialect := range dialects {
		schema_sql, err := dialect.RenderSchema(schema)
		if err != nil {
			return err
		}

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

	return dbx.RenderTemplate(g.header_tmpl, w, "", params)
}

func (g *Golang) RenderFooter(w io.Writer) (err error) {
	type footerParams struct {
		Funcs []string
	}

	params := footerParams{
		Funcs: g.funcs,
	}

	return dbx.RenderTemplate(g.footer_tmpl, w, "", params)
}

func GolangArgsFromColumns(columns []*dbx.Column) (out []GolangArg) {
	for _, column := range columns {
		out = append(out, GolangArgFromColumn(column))
	}
	return out
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
	column *dbx.Column
}

type GolangArg struct {
	Name        string
	Type        string
	from_column *dbx.Column
}

type GolangFuncBase struct {
	Struct     string
	FuncSuffix string
	Args       []GolangArg
}

type GolangFunc struct {
	GolangFuncBase
	Dialect string
	SQL     string
}

func MakeGolangFuncBase(table *dbx.Table, args []GolangArg) GolangFuncBase {
	return GolangFuncBase{
		Struct:     GolangStructName(table),
		Args:       args,
		FuncSuffix: GolangFuncSuffix(table, args),
	}
}

func MakeGolangFunc(base GolangFuncBase, dialect, sql string) GolangFunc {
	return GolangFunc{
		GolangFuncBase: base,
		Dialect:        dialect,
		SQL:            sql,
	}
}

type GolangSelect struct {
	GolangFunc
	PagedOn string
}

type GolangInsertBase struct {
	GolangFuncBase
	Inserts []GolangArg
}

type GolangReturnBy struct {
	Pk     string
	Getter *GolangFuncBase
}

type GolangInsert struct {
	GolangFunc
	ReturnBy *GolangReturnBy
	NeedsNow bool
	Inserts  []GolangArg
	Autos    []GolangAutoParam
}

type GolangUpdate struct {
	GolangFunc
	SupportsReturning bool
	NeedsNow          bool
	Autos             []GolangAutoParam
}

func (g *Golang) RenderSelect(w io.Writer, dialects []dbx.Dialect,
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

	base := MakeGolangFuncBase(
		params.Table,
		GolangArgsFromConditions(params.Conditions),
	)

	if err = g.renderBase(g.tmpl, w, tmpl, base); err != nil {
		return err
	}

	data := GolangSelect{
		PagedOn: GolangFieldName(params.PagedOn),
	}

	for _, dialect := range dialects {
		sql, err := dialect.RenderSelect(params)
		if err != nil {
			return err
		}
		data.GolangFunc = MakeGolangFunc(base, dialect.Name(), sql)
		if err = dbx.RenderTemplate(g.tmpl, w, tmpl, data); err != nil {
			return err
		}
	}
	return nil
}

func (g *Golang) RenderCount(w io.Writer, dialects []dbx.Dialect,
	params *dbx.SelectParams) (err error) {

	base := MakeGolangFuncBase(
		params.Table,
		GolangArgsFromConditions(params.Conditions),
	)
	if err = g.renderBase(g.tmpl, w, "count", base); err != nil {
		return err
	}
	if err = g.renderBase(g.tmpl, w, "has", base); err != nil {
		return err
	}

	for _, dialect := range dialects {
		sql, err := dialect.RenderCount(params)
		if err != nil {
			return err
		}
		data := MakeGolangFunc(base, dialect.Name(), sql)
		if err = dbx.RenderTemplate(g.tmpl, w, "count", data); err != nil {
			return err
		}
		if err = dbx.RenderTemplate(g.tmpl, w, "has", data); err != nil {
			return err
		}
	}
	return nil
}

func (g *Golang) RenderDelete(w io.Writer, dialects []dbx.Dialect,
	params *dbx.DeleteParams) (err error) {

	tmpl := "delete"
	if params.Many {
		tmpl = "delete-all"
	}

	base := MakeGolangFuncBase(
		params.Table,
		GolangArgsFromConditions(params.Conditions),
	)

	if err = g.renderBase(g.tmpl, w, tmpl, base); err != nil {
		return err
	}

	for _, dialect := range dialects {
		sql, err := dialect.RenderDelete(params)
		if err != nil {
			return err
		}
		data := MakeGolangFunc(base, dialect.Name(), sql)
		if err = dbx.RenderTemplate(g.tmpl, w, tmpl, data); err != nil {
			return err
		}
	}
	return nil
}

func (g *Golang) RenderInsert(w io.Writer, dialects []dbx.Dialect,
	params *dbx.InsertParams) (err error) {

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
					Column: column.SQLName(),
				})
			} else {
				args = append(args, arg)
			}
		} else {
			args = append(args, arg)
		}
	}

	base := GolangInsertBase{
		GolangFuncBase: MakeGolangFuncBase(params.Table, args),
		Inserts:        all,
	}
	if err = g.renderBase(g.tmpl, w, "insert", base); err != nil {
		return err
	}
	if err = g.renderBase(g.tmpl, w, "raw-insert", base); err != nil {
		return err
	}

	data := GolangInsert{
		Inserts:  all,
		Autos:    autos,
		NeedsNow: needs_now,
	}

	for _, dialect := range dialects {
		if dialect.SupportsReturning() {
			data.ReturnBy = nil
		} else if pk := params.Table.BasicPrimaryKey(); pk != nil {
			data.ReturnBy = &GolangReturnBy{
				Pk: GolangFieldName(pk),
			}
		} else {
			getter := MakeGolangFuncBase(
				params.Table,
				GolangArgsFromColumns(params.Table.PrimaryKey),
			)
			data.ReturnBy = &GolangReturnBy{
				Getter: &getter,
			}
		}

		sql, err := dialect.RenderInsert(params)
		if err != nil {
			return err
		}
		data.GolangFunc = MakeGolangFunc(base.GolangFuncBase,
			dialect.Name(), sql)
		if err = dbx.RenderTemplate(g.tmpl, w, "insert", data); err != nil {
			return err
		}
		if err = dbx.RenderTemplate(g.tmpl, w, "raw-insert", data); err != nil {
			return err
		}
	}
	return nil
}

func (g *Golang) RenderUpdate(w io.Writer, dialects []dbx.Dialect,
	params *dbx.UpdateParams) (err error) {

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
				Column: column.SQLName(),
			})
		}
	}

	base := MakeGolangFuncBase(
		params.Table,
		GolangArgsFromConditions(params.Conditions),
	)
	if err = g.renderBase(g.tmpl, w, "update", base); err != nil {
		return err
	}

	data := GolangUpdate{
		Autos:    autos,
		NeedsNow: needs_now,
	}

	for _, dialect := range dialects {
		data.SupportsReturning = dialect.SupportsReturning()

		sql, err := dialect.RenderUpdate(params)
		if err != nil {
			return err
		}
		data.GolangFunc = MakeGolangFunc(base, dialect.Name(), sql)
		if err = dbx.RenderTemplate(g.tmpl, w, "update", data); err != nil {
			return err
		}
	}
	return nil
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
		Column: column.SQLName(),
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

func (g *Golang) renderBase(tmpl *template.Template, w io.Writer,
	name string, base interface{}) (err error) {

	var buf bytes.Buffer
	err = dbx.RenderTemplate(tmpl, &buf, name+"-func-sig", base)
	if err != nil {
		return err
	}
	g.funcs = append(g.funcs, buf.String())
	return nil
}
