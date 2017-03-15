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

package sql

import (
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlcompile"
	. "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlhelpers"
)

func RenderSchema(dialect Dialect, ir_root *ir.Root) string {
	schema := SchemaFromIR(ir_root.Models, dialect)
	sql := SQLFromSchema(schema)
	return sqlgen.Render(dialect, sql, sqlgen.NoFlatten, sqlgen.NoTerminate)
}

func SQLFromSchema(schema *Schema) sqlgen.SQL {
	var stmts []sqlgen.SQL

	for _, table := range schema.Tables {
		var dirs []sqlgen.SQL

		for _, column := range table.Columns {
			dir := Build(Lf("%s %s", column.Name, column.Type))
			if column.NotNull {
				dir.Add(L("NOT NULL"))
			}
			if ref := column.Reference; ref != nil {
				dir.Add(Lf("REFERENCES %s(%s)", ref.Table, ref.Column))
				if ref.OnDelete != "" {
					dir.Add(Lf("ON DELETE %s", ref.OnDelete))
				}
				if ref.OnUpdate != "" {
					dir.Add(Lf("ON UPDATE %s", ref.OnUpdate))
				}
			}
			dirs = append(dirs, dir.SQL())
		}

		if pkey := table.PrimaryKey; len(pkey) > 0 {
			dir := Build(L("PRIMARY KEY ("))
			dir.Add(J(", ", Strings(pkey)...))
			dir.Add(L(")"))
			dirs = append(dirs, dir.SQL())
		}

		for _, unique := range table.Unique {
			dir := Build(L("UNIQUE ("))
			dir.Add(J(", ", Strings(unique)...))
			dir.Add(L(")"))
			dirs = append(dirs, dir.SQL())
		}

		directives := J(",\n\t", dirs...)

		stmt := J("",
			Lf("CREATE TABLE %s (\n\t", table.Name),
			directives,
			Lf("\n);"),
		)

		stmts = append(stmts, stmt)
	}

	for _, index := range schema.Indexes {
		stmt := Build(L("CREATE"))
		if index.Unique {
			stmt.Add(L("UNIQUE"))
		}
		stmt.Add(Lf("INDEX %s ON %s (", index.Name, index.Table))
		stmt.Add(J(", ", Strings(index.Columns)...))
		stmt.Add(L(");"))

		stmts = append(stmts, stmt.SQL())
	}

	return sqlcompile.Compile(J("\n", stmts...))
}

func SchemaFromIR(ir_models []*ir.Model, dialect Dialect) *Schema {
	schema := &Schema{}
	for _, ir_model := range ir_models {
		table := Table{
			Name: ir_model.Table,
		}
		for _, ir_field := range ir_model.PrimaryKey {
			table.PrimaryKey = append(table.PrimaryKey, ir_field.Column)
		}
		for _, ir_unique := range ir_model.Unique {
			var unique []string
			for _, ir_field := range ir_unique {
				unique = append(unique, ir_field.Column)
			}
			table.Unique = append(table.Unique, unique)
		}
		for _, ir_field := range ir_model.Fields {
			column := Column{
				Name:    ir_field.Column,
				Type:    dialect.ColumnType(ir_field),
				NotNull: !ir_field.Nullable,
			}
			if ir_field.Relation != nil {
				column.Reference = &Reference{
					Table:  ir_field.Relation.Field.Model.Table,
					Column: ir_field.Relation.Field.Column,
				}
				switch ir_field.Relation.Kind {
				case consts.SetNull:
					column.Reference.OnDelete = "SET NULL"
					//column.Reference.OnUpdate = "RESTRICT"
				case consts.Cascade:
					column.Reference.OnDelete = "CASCADE"
					//column.Reference.OnUpdate = "RESTRICT"
				case consts.Restrict:
					column.Reference.OnDelete = ""
					//column.Reference.OnUpdate = "RESTRICT"
				default:
					panic(fmt.Sprintf("unhandled relation kind %q",
						ir_field.Relation.Kind))
				}
			}
			table.Columns = append(table.Columns, column)
		}
		schema.Tables = append(schema.Tables, table)
		for _, ir_index := range ir_model.Indexes {
			index := Index{
				Name:   ir_index.Name,
				Table:  ir_index.Model.Table,
				Unique: ir_index.Unique,
			}
			for _, ir_field := range ir_index.Fields {
				index.Columns = append(index.Columns, ir_field.Column)
			}
			schema.Indexes = append(schema.Indexes, index)
		}
	}
	return schema
}

type Schema struct {
	Tables  []Table
	Indexes []Index
}

type Table struct {
	Name       string
	Columns    []Column
	PrimaryKey []string
	Unique     [][]string
}

type Column struct {
	Name      string
	Type      string
	NotNull   bool
	Reference *Reference
}

type Reference struct {
	Table    string
	Column   string
	OnDelete string
	OnUpdate string
}

type Index struct {
	Name    string
	Table   string
	Columns []string
	Unique  bool
}
