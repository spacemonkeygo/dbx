package internal

import (
	"text/template"
)

type LeftJoinParams struct {
	Left  *Column
	Right *Column
}

func LeftJoin(relation *Relation) *LeftJoinParams {
	return &LeftJoinParams{
		Left:  relation.Column,
		Right: relation.ForeignColumn,
	}
}

func Where(conditions ...*ConditionParams) *WhereParams {
	return &WhereParams{Conditions: conditions}
}

type ConditionParams struct {
	Left     *Column
	Right    *Column
	Operator string
}

func EqualsQ(left *Column) *ConditionParams {
	return Equals(left, nil)
}

func Equals(left, right *Column) *ConditionParams {
	return &ConditionParams{
		Left:     left,
		Right:    right,
		Operator: "=",
	}
}

type WhereParams struct {
	Conditions []*ConditionParams
	PagingOn   *Column
}

type SelectParams struct {
	Many      bool
	Table     *Table
	LeftJoins []*LeftJoinParams
	Where     *WhereParams
}

type DeleteParams struct {
	Many      bool
	Table     *Table
	LeftJoins []*LeftJoinParams
	Where     *WhereParams
}

type SQL struct {
	tmpl *template.Template
}

func NewSQL(loader Loader, name string) (*SQL, error) {
	data, err := loader.Load(name)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("").Funcs(globalFuncs).Parse(string(data))
	if err != nil {
		return nil, err
	}

	return &SQL{
		tmpl: tmpl,
	}, nil
}

func (s *SQL) RenderSelect(params *SelectParams) (
	string, error) {

	return RenderTemplate(s.tmpl, "select", params)
}

func (s *SQL) RenderDelete(params *DeleteParams) (
	string, error) {

	return RenderTemplate(s.tmpl, "delete", params)
}
