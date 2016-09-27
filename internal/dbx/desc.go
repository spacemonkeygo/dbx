package dbx

type Column struct {
	Table    *Table
	Name     string
	Type     string
	NotNull  bool
	Relation *Column
}

type Table struct {
	Name       string
	Columns    []*Column
	Unique     [][]*Column
	PrimaryKey []*Column
	Queries    []*Query
}

type Query struct {
	Start []*Column
	Joins []*Relation
	End   []*Column
}

type Relation struct {
	Left  *Column
	Right *Column
}

type Schema struct {
	Tables []*Table
}

func (c *Column) RelationLeft() *Relation {
	return &Relation{
		Left:  c,
		Right: c.Relation,
	}
}

func (c *Column) RelationRight() *Relation {
	return &Relation{
		Left:  c.Relation,
		Right: c,
	}
}

func (c *Column) Insertable() bool {
	if c.Relation != nil {
		return true
	}
	return c.Type != "serial" && c.Type != "serial64"
}

func (t *Table) GetColumn(name string) *Column {
	for _, column := range t.Columns {
		if column.Name == name {
			return column
		}
	}
	return nil
}

func (t *Table) GetColumns(names ...string) (out []*Column) {
	for _, name := range names {
		column := t.GetColumn(name)
		if column == nil {
			return nil
		}
		out = append(out, column)
	}
	return out
}

func (t *Table) InsertableColumns() (out []*Column) {
	for _, column := range t.Columns {
		if !column.Insertable() {
			continue
		}
		out = append(out, column)
	}
	return out
}

// returns true if left is a subset of right
func columnSetSubset(left, right []*Column) bool {
	if len(left) > len(right) {
		return false
	}
lcols:
	for _, lcol := range left {
		for _, rcol := range right {
			if lcol == rcol {
				continue lcols
			}
		}
		return false
	}
	return true
}

func (t *Table) ColumnSetUnique(columns []*Column) bool {
	if columnSetSubset(t.PrimaryKey, columns) {
		return true
	}
	for _, unique := range t.Unique {
		if columnSetSubset(unique, columns) {
			return true
		}
	}
	return false
}
