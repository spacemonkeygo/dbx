package dbx

type JoinParams struct {
	Left  *Column
	Right *Column
}

func Joins(joins ...*Join) (out []*JoinParams) {
	for _, join := range joins {
		out = append(out, &JoinParams{
			Left:  join.Left,
			Right: join.Right,
		})
	}

	return out
}

func Where(conditions ...*ConditionParams) []*ConditionParams {
	return conditions
}

type ColumnCmpParams struct {
	Left     *Column
	Operator string
}

type ColumnCmpColumnParams struct {
	Left     *Column
	Right    *Column
	Operator string
}

type ColumnInParams struct {
	Left *Column
	In   *SelectParams
}

type ConditionParams struct {
	ColumnCmp       *ColumnCmpParams
	ColumnCmpColumn *ColumnCmpColumnParams
	ColumnIn        *ColumnInParams
}

func ColumnEquals(left *Column) *ConditionParams {
	return &ConditionParams{
		ColumnCmp: &ColumnCmpParams{
			Left:     left,
			Operator: "=",
		},
	}
}

func ColumnEqualsColumn(left, right *Column) *ConditionParams {
	return &ConditionParams{
		ColumnCmpColumn: &ColumnCmpColumnParams{
			Left:     left,
			Right:    right,
			Operator: "=",
		},
	}
}

func ColumnIn(left *Column, in *SelectParams) *ConditionParams {
	return &ConditionParams{
		ColumnIn: &ColumnInParams{
			Left: left,
			In:   in,
		},
	}
}

type SelectParams struct {
	Many       bool
	What       []*Column
	Table      *Table
	LeftJoins  []*JoinParams
	Conditions []*ConditionParams
	PagedOn    *Column
}

func What(columns ...*Column) []*Column {
	return columns
}

type DeleteParams struct {
	Many       bool
	Table      *Table
	Conditions []*ConditionParams
}

type InsertParams struct {
	Table   *Table
	Columns []*Column
}
