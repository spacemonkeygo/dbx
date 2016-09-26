package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"strings"
)

type Field struct {
	Model      *Model
	Name       string
	Unique     bool
	PrimaryKey bool
	Relation   *Field
}

func (f *Field) Param() string {
	return f.Model.Name + "_" + f.Name
}

type Model struct {
	Name   string
	Fields []*Field
}

func (m *Model) LookupField(name string) *Field {
	for _, field := range m.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (m *Model) Indexed() (fields []*Field) {
	for _, field := range m.Fields {
		if field.PrimaryKey || field.Unique {
			fields = append(fields, field)
		}
	}
	return fields
}

type Relation struct {
	Field        *Field
	ForeignField *Field
}

func (m *Model) Relations() (relations []*Relation) {
	for _, field := range m.Fields {
		if field.Relation != nil {
			relations = append(relations, &Relation{
				Field:        field,
				ForeignField: field.Relation,
			})
		}
	}
	return relations
}

func (m *Model) RelationChains() (chains [][]*Relation) {
	relations := m.Relations()

	var all_nested [][]*Relation
	for _, relation := range relations {
		all_nested = append(all_nested,
			relation.ForeignField.Model.RelationChains()...)
	}

	for _, relation := range relations {
		chain := []*Relation{relation}
		if len(all_nested) > 0 {
			for _, nested := range all_nested {
				chains = append(chains, append(chain, nested...))
			}
		} else {
			chains = append(chains, chain)
		}
	}
	return chains
}

type Schema struct {
	Models []*Model
}

func main() {
	project := &Model{Name: "project"}
	project.Fields = append(project.Fields, &Field{
		Model:      project,
		Name:       "pk",
		PrimaryKey: true,
	}, &Field{
		Model:  project,
		Name:   "id",
		Unique: true,
	})

	bookie := &Model{Name: "bookie"}
	bookie.Fields = append(bookie.Fields, &Field{
		Model:      bookie,
		Name:       "pk",
		PrimaryKey: true,
	}, &Field{
		Model:  bookie,
		Name:   "id",
		Unique: true,
	}, &Field{
		Model:    bookie,
		Name:     "project_id",
		Relation: project.LookupField("id"),
	})

	billing_key := &Model{Name: "billing_key"}
	billing_key.Fields = append(billing_key.Fields, &Field{
		Model:      billing_key,
		Name:       "pk",
		PrimaryKey: true,
	}, &Field{
		Model:  billing_key,
		Name:   "id",
		Unique: true,
	}, &Field{
		Model:    billing_key,
		Name:     "bookie_id",
		Relation: bookie.LookupField("id"),
	})

	schema := &Schema{
		Models: []*Model{
			project,
			bookie,
			billing_key,
		},
	}

	w := NewStatementWriter(os.Stdout)
	if err := w.WriteStatements(schema); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type StatementWriter struct {
	w   io.Writer
	err error
}

func NewStatementWriter(w io.Writer) *StatementWriter {
	return &StatementWriter{
		w: w,
	}
}

func (sw *StatementWriter) WriteStatements(schema *Schema) (err error) {
	for _, model := range schema.Models {
		sw.writeModelStatements(model)
	}
	return sw.err
}

func (sw *StatementWriter) writeModelStatements(model *Model) {
	indexed := model.Indexed()

	fmt.Println("// SELECT ALL")
	sw.render(SelectAll(model))
	for _, field := range indexed {
		fmt.Println("// SELECT ONE")
		sw.render(SelectOne(field))
	}

	for _, chain := range model.RelationChains() {
		fmt.Println("// JOIN SELECT ALL")
		for _, foreign_by := range chain[len(chain)-1].ForeignField.Model.Indexed() {
			sw.render(JoinSelectAll(foreign_by, chain...))
			for _, by := range indexed {
				fmt.Println("// JOIN SELECT ONE")
				sw.render(JoinSelectOne(by, foreign_by, chain...))
			}
		}
	}
}

type Statement struct {
	Verb    Verb
	Clauses []Clause
}

func NewStatement(all bool, verb Verb, clauses ...Clause) *Statement {
	return &Statement{
		Verb:    verb,
		Clauses: clauses,
	}
}

func (s *Statement) Name() string {
	names := []string{s.Verb.Name()}
	for _, clause := range s.Clauses {
		names = append(names, clause.Name())
	}

	return strings.Join(names, "_")
}

func PascalCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(s, "_", " ", -1)), " ", "", -1)
}

func (s *Statement) Params() (params []*Field) {
	for _, clause := range s.Clauses {
		params = append(params, clause.Params()...)
	}
	return params
}

var funcTmpl = template.Must(template.New("func").Parse(`
func {{ .Name }}({{ range $i, $e := .Params }}{{if $i}},{{end}}{{ $e.Param }}{{end}}) {
	const stmt=` + "`{{ .SQL }}`" + `
	_, err = db.Exec(stmt, {{ range $i, $e := .Params }}{{if $i}},{{end}}{{ $e.Name }}{{end}})
	if err != nil {
		return wrapError(err)
	}
	return nil
}`))

func (s *Statement) RenderFunc() (string, error) {
	rendered_sql, err := s.RenderSQL()
	if err != nil {
		return "", nil
	}

	names := []string{s.Verb.Name()}
	for _, clause := range s.Clauses {
		names = append(names, clause.Name())
	}

	name := PascalCase(strings.Join(names, ","))
	return evalTemplate(funcTmpl, map[string]interface{}{
		"Name":   name,
		"Params": s.Params(),
		"SQL":    rendered_sql,
	})
}

func (s *Statement) RenderSQL() (string, error) {
	var rendered []string
	if r, err := s.Verb.RenderSQL(); err != nil {
		return "", err
	} else {
		rendered = append(rendered, r)
	}

	for _, clause := range s.Clauses {
		r, err := clause.RenderSQL()
		if err != nil {
			return "", err
		}
		rendered = append(rendered, r)
	}
	return strings.Join(rendered, "\n\t\t"), nil
}

func SelectAll(model *Model) *Statement {
	return NewStatement(true, Select(model))
}

func SelectOne(field *Field, relations ...*Relation) *Statement {
	return NewStatement(false, Select(field.Model), Where(ColumnEqual(field)))
}

func JoinSelectAll(foreign_by *Field, relations ...*Relation) *Statement {
	return joinSelect(nil, foreign_by, relations...)
}

func JoinSelectOne(by, foreign_by *Field, relations ...*Relation) *Statement {
	return joinSelect(by, foreign_by, relations...)
}

func joinSelect(by, foreign_by *Field, relations ...*Relation) *Statement {
	var clauses []Clause

	// left join
	for _, relation := range relations {
		clauses = append(clauses, LeftJoin(relation))
	}

	// where
	foreign_equal := ColumnEqual(foreign_by)
	if by != nil {
		clauses = append(clauses, Where(foreign_equal, ColumnEqual(by)))
	} else {
		clauses = append(clauses, Where(foreign_equal))
	}

	return NewStatement(by == nil, Select(relations[0].Field.Model), clauses...)
}

type Renderer interface {
	RenderSQL() (string, error)
	RenderFunc() (string, error)
}

func (sw *StatementWriter) render(renderer Renderer) {
	rendered, err := renderer.RenderFunc()
	sw.setError(err)
	sw.printf("%s;\n", rendered)
}

func (sw *StatementWriter) printf(format string, args ...interface{}) {
	if sw.err != nil {
		return
	}
	_, err := fmt.Fprintf(sw.w, format, args...)
	sw.setError(err)
}

func joinClauseName(clauses []Clause) string {
	var names []string
	for _, clause := range clauses {
		names = append(names, clause.Name())
	}
	return strings.Join(names, "_")
}

func (sw *StatementWriter) setError(err error) {
	if err != nil && sw.err == nil {
		sw.err = err
	}
}

type Verb interface {
	Name() string
	RenderSQL() (string, error)
}

type Clause interface {
	Name() string
	Params() []*Field
	RenderSQL() (string, error)
}

var selectTmpl = template.Must(template.New("select").Parse(
	`SELECT {{ .Table }}.* FROM {{ .Table }}`))

type NoParams struct{}

func (NoParams) Params() []*Field { return nil }

type SelectVerb struct {
	NoParams
	Model *Model
}

func Select(model *Model) *SelectVerb {
	return &SelectVerb{
		Model: model,
	}
}

func (c *SelectVerb) Name() string {
	return "get_" + c.Model.Name
}

func (c *SelectVerb) RenderSQL() (string, error) {
	return evalTemplate(selectTmpl, map[string]interface{}{
		"Table": c.Model.Name,
	})
}

var whereTmpl = template.Must(template.New("where").Parse(
	`WHERE {{ range $i, $e := .Conditions }}{{ if $i }} AND {{ end }}{{ $e.RenderSQL }}{{ end }}`))

type WhereClause struct {
	Conditions []Clause
}

func Where(conditions ...Clause) *WhereClause {
	return &WhereClause{
		Conditions: conditions,
	}
}

func (c *WhereClause) Name() string {
	var conds []string
	for _, cond := range c.Conditions {
		conds = append(conds, cond.Name())
	}

	return "where_" + strings.Join(conds, "_and_")
}

func (c *WhereClause) Params() (params []*Field) {
	for _, cond := range c.Conditions {
		params = append(params, cond.Params()...)
	}
	return params
}

func (c *WhereClause) RenderSQL() (string, error) {
	return evalTemplate(whereTmpl, map[string]interface{}{
		"Conditions": c.Conditions,
	})
}

var columnEqualTmpl = template.Must(template.New("column_equal").Parse(
	`{{ .Table }}.{{ .Column }} = ?`))

type ColumnEqualClause struct {
	Field *Field
}

func ColumnEqual(field *Field) *ColumnEqualClause {
	return &ColumnEqualClause{
		Field: field,
	}
}

func (c *ColumnEqualClause) Params() []*Field {
	return []*Field{c.Field}
}

func (c *ColumnEqualClause) Name() string {
	return c.Field.Model.Name + "_" + c.Field.Name + "_eq_q"
}

func (c *ColumnEqualClause) RenderSQL() (string, error) {
	return evalTemplate(columnEqualTmpl, map[string]interface{}{
		"Table":  c.Field.Model.Name,
		"Column": c.Field.Name,
	})
}

var leftJoinTmpl = template.Must(template.New("left_join").Parse(
	`LEFT JOIN {{ .ForeignTable }} ON {{ .Table }}.{{ .Column }} = {{ .ForeignTable }}.{{ .ForeignColumn }}`))

type LeftJoinClause struct {
	NoParams
	Relation *Relation
}

func LeftJoin(relation *Relation) *LeftJoinClause {
	return &LeftJoinClause{
		Relation: relation,
	}
}

func (c *LeftJoinClause) Name() string {
	return "left_join_" + c.Relation.ForeignField.Model.Name + "_on_" +
		c.Relation.Field.Model.Name + "_" + c.Relation.Field.Name + "_eq_" +
		c.Relation.ForeignField.Name
}

func (c *LeftJoinClause) RenderSQL() (string, error) {
	return evalTemplate(leftJoinTmpl, map[string]interface{}{
		"Table":         c.Relation.Field.Model.Name,
		"Column":        c.Relation.Field.Name,
		"ForeignTable":  c.Relation.ForeignField.Model.Name,
		"ForeignColumn": c.Relation.ForeignField.Name,
	})
}

func evalTemplate(tmpl *template.Template, data interface{}) (string, error) {
	var w bytes.Buffer
	err := tmpl.Execute(&w, data)
	return w.String(), err
}
