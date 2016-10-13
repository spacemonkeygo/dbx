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

package dialect

import (
	"text/template"

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v0/internal/dbx"
)

func SQLite3TableName(table *dbx.Table) string {
	return inflect.Underscore(inflect.Pluralize(table.Name))
}

func SQLite3ColumnName(column *dbx.Column) string {
	return inflect.Underscore(column.Name)
}

func SQLite3ColumnNames(columns []*dbx.Column) (out []string) {
	for _, column := range columns {
		out = append(out, SQLite3ColumnName(column))
	}
	return out
}

func SQLite3ColumnType(column *dbx.Column) string {
	switch column.Type {
	case "text":
		return "text"
	case "int":
		return "integer"
	case "serial":
		return "serial"
	case "int64":
		return "bigint"
	case "serial64":
		return "bigserial"
	case "blob":
		return "bytea"
	case "timestamp":
		return "timestamp with time zone"
	case "bool":
		return "boolean"
	}
	panic("unhandled column type " + "%s")
}

type SQLite3Schema struct {
	Tables []SQLite3Table
}

func SQLite3SchemaFromSchema(schema *dbx.Schema) SQLite3Schema {
	return SQLite3Schema{
		Tables: SQLite3TablesFromTables(schema.Tables),
	}
}

type SQLite3Table struct {
	Name       string
	Columns    []SQLite3Column
	PrimaryKey []string
	Unique     [][]string
}

func SQLite3TableFromTable(table *dbx.Table) SQLite3Table {
	t := SQLite3Table{
		Name:       SQLite3TableName(table),
		Columns:    SQLite3ColumnsFromColumns(table.Columns),
		PrimaryKey: SQLite3ColumnNames(table.PrimaryKey),
	}
	for _, unique := range table.Unique {
		t.Unique = append(t.Unique, SQLite3ColumnNames(unique))
	}
	return t
}

func SQLite3TablesFromTables(tables []*dbx.Table) (out []SQLite3Table) {
	for _, table := range tables {
		out = append(out, SQLite3TableFromTable(table))
	}
	return out
}

type SQLite3Reference struct {
	Table    string
	Column   string
	OnDelete string
}

func SQLite3ReferenceFromRelation(relation *dbx.Relation) *SQLite3Reference {
	if relation == nil {
		return nil
	}
	r := &SQLite3Reference{
		Table:  SQLite3TableName(relation.Column.Table),
		Column: SQLite3ColumnName(relation.Column),
	}

	switch relation.Kind {
	case dbx.HasA:
		r.OnDelete = "RESTRICT"
	case dbx.OwnedBy:
		r.OnDelete = "CASCADE"
	}
	return r
}

type SQLite3Column struct {
	Name      string
	Type      string
	NotNull   bool
	Reference *SQLite3Reference
}

func SQLite3ColumnFromColumn(column *dbx.Column) SQLite3Column {
	return SQLite3Column{
		Name:      SQLite3ColumnName(column),
		Type:      SQLite3ColumnType(column),
		NotNull:   column.NotNull,
		Reference: SQLite3ReferenceFromRelation(column.Relation),
	}
}

func SQLite3ColumnsFromColumns(columns []*dbx.Column) (out []SQLite3Column) {
	for _, column := range columns {
		out = append(out, SQLite3ColumnFromColumn(column))
	}
	return out
}

type SQLite3 struct {
	tmpl *template.Template
}

func NewSQLite3(loader dbx.Loader) (*SQLite3, error) {
	tmpl, err := loader.Load("sqlite3.tmpl")
	if err != nil {
		return nil, err
	}

	return &SQLite3{
		tmpl: tmpl,
	}, nil
}

func (s *SQLite3) Name() string {
	return "sqlite3"
}

func (s *SQLite3) ColumnName(column *dbx.Column) string {
	return SQLite3ColumnName(column)
}

func (s *SQLite3) ListTablesSQL() string {
	return `SELECT tablename FROM pg_tables WHERE schemaname = 'public';`
}

func (s *SQLite3) RenderSchema(schema *dbx.Schema) (
	string, error) {

	return dbx.RenderTemplateString(s.tmpl, "schema",
		SQLite3SchemaFromSchema(schema))
}

type SQLite3ColumnRef struct {
	Table  string
	Column string
}

func SQLite3ColumnRefFromColumn(column *dbx.Column) SQLite3ColumnRef {
	return SQLite3ColumnRef{
		Table:  SQLite3TableName(column.Table),
		Column: SQLite3ColumnName(column),
	}
}

type SQLite3ColumnCmp struct {
	Left     SQLite3ColumnRef
	Operator string
}

func SQLite3ColumnCmpFromColumnCmp(
	params *dbx.ColumnCmpParams) *SQLite3ColumnCmp {

	if params == nil {
		return nil
	}
	return &SQLite3ColumnCmp{
		Left:     SQLite3ColumnRefFromColumn(params.Left),
		Operator: params.Operator,
	}
}

type SQLite3ColumnCmpColumn struct {
	Left     SQLite3ColumnRef
	Right    SQLite3ColumnRef
	Operator string
}

func SQLite3ColumnCmpColumnFromColumnCmpColumn(
	params *dbx.ColumnCmpColumnParams) *SQLite3ColumnCmpColumn {

	if params == nil {
		return nil
	}
	return &SQLite3ColumnCmpColumn{
		Left:     SQLite3ColumnRefFromColumn(params.Left),
		Operator: params.Operator,
	}
}

type SQLite3ColumnIn struct {
	Left SQLite3ColumnRef
	In   SQLite3Select
}

func SQLite3ColumnInFromColumnIn(
	params *dbx.ColumnInParams) *SQLite3ColumnIn {

	if params == nil {
		return nil
	}
	return &SQLite3ColumnIn{
		Left: SQLite3ColumnRefFromColumn(params.Left),
		In:   SQLite3SelectFromSelect(params.In),
	}
}

type SQLite3Condition struct {
	ColumnCmp       *SQLite3ColumnCmp
	ColumnCmpColumn *SQLite3ColumnCmpColumn
	ColumnIn        *SQLite3ColumnIn
}

func SQLite3ConditionFromCondition(
	params *dbx.ConditionParams) SQLite3Condition {
	return SQLite3Condition{
		ColumnCmp:       SQLite3ColumnCmpFromColumnCmp(params.ColumnCmp),
		ColumnCmpColumn: SQLite3ColumnCmpColumnFromColumnCmpColumn(params.ColumnCmpColumn),
		ColumnIn:        SQLite3ColumnInFromColumnIn(params.ColumnIn),
	}
}

func SQLite3ConditionsFromConditions(conditions []*dbx.ConditionParams) (
	out []SQLite3Condition) {

	for _, condition := range conditions {
		out = append(out, SQLite3ConditionFromCondition(condition))
	}
	return out
}

type SQLite3OrderBy struct {
	Column SQLite3ColumnRef
}

type SQLite3Join struct {
	Left  SQLite3ColumnRef
	Right SQLite3ColumnRef
}

func SQLite3JoinFromJoin(params *dbx.JoinParams) SQLite3Join {
	return SQLite3Join{
		Left:  SQLite3ColumnRefFromColumn(params.Left),
		Right: SQLite3ColumnRefFromColumn(params.Right),
	}
}

func SQLite3JoinsFromJoins(joins []*dbx.JoinParams) (out []SQLite3Join) {
	for _, join := range joins {
		out = append(out, SQLite3JoinFromJoin(join))
	}
	return out
}

type SQLite3Select struct {
	Table      string
	What       []string
	LeftJoins  []SQLite3Join
	Conditions []SQLite3Condition
	OrderBy    *SQLite3OrderBy
	Limit      bool
}

func SQLite3SelectFromSelect(params *dbx.SelectParams) SQLite3Select {
	out := SQLite3Select{
		Table:      SQLite3TableName(params.Table),
		What:       SQLite3ColumnNames(params.What),
		LeftJoins:  SQLite3JoinsFromJoins(params.LeftJoins),
		Conditions: SQLite3ConditionsFromConditions(params.Conditions),
	}
	if params.PagedOn != nil {
		out.Conditions = append(out.Conditions, SQLite3ConditionFromCondition(
			&dbx.ConditionParams{
				ColumnCmp: &dbx.ColumnCmpParams{
					Left:     params.PagedOn,
					Operator: ">",
				},
			}))
		out.OrderBy = &SQLite3OrderBy{
			Column: SQLite3ColumnRefFromColumn(params.PagedOn),
		}
		out.Limit = true
	}
	return out
}

func (s *SQLite3) RenderSelect(params *dbx.SelectParams) (
	string, error) {

	return dbx.RenderTemplateString(s.tmpl, "select",
		SQLite3SelectFromSelect(params))
}

func (s *SQLite3) RenderCount(params *dbx.SelectParams) (
	string, error) {

	return dbx.RenderTemplateString(s.tmpl, "count",
		SQLite3SelectFromSelect(params))
}

type SQLite3Delete struct {
	Table      string
	Conditions []SQLite3Condition
}

func SQLite3DeleteFromDelete(params *dbx.DeleteParams) SQLite3Delete {
	return SQLite3Delete{
		Table:      SQLite3TableName(params.Table),
		Conditions: SQLite3ConditionsFromConditions(params.Conditions),
	}
}

func (s *SQLite3) RenderDelete(params *dbx.DeleteParams) (
	string, error) {

	return dbx.RenderTemplateString(s.tmpl, "delete",
		SQLite3DeleteFromDelete(params))
}

type SQLite3Insert struct {
	Table   string
	Columns []string
}

func SQLite3InsertFromInsert(params *dbx.InsertParams) SQLite3Insert {
	return SQLite3Insert{
		Table:   SQLite3TableName(params.Table),
		Columns: SQLite3ColumnNames(params.Columns),
	}
}

func (s *SQLite3) RenderInsert(params *dbx.InsertParams) (
	string, error) {

	return dbx.RenderTemplateString(s.tmpl, "insert",
		SQLite3InsertFromInsert(params))
}

type SQLite3Update struct {
	Table      string
	Conditions []SQLite3Condition
}

func SQLite3UpdateFromUpdate(params *dbx.UpdateParams) SQLite3Update {
	return SQLite3Update{
		Table:      SQLite3TableName(params.Table),
		Conditions: SQLite3ConditionsFromConditions(params.Conditions),
	}
}

func (s *SQLite3) RenderUpdate(params *dbx.UpdateParams) (
	string, error) {

	return dbx.RenderTemplateString(s.tmpl, "update",
		SQLite3UpdateFromUpdate(params))
}

func (s *SQLite3) SupportsReturning() bool {
	return false
}
