package internal

type Column struct {
	Table      *Table
	Name       string
	Type       string
	Unique     bool
	PrimaryKey bool
	Relation   *Column
}

type Table struct {
	Name    string
	Columns []*Column
}

func (t *Table) Column(name string) *Column {
	for _, column := range t.Columns {
		if column.Name == name {
			return column
		}
	}
	return nil
}

func (t *Table) Unique() (columns []*Column) {
	for _, column := range t.Columns {
		if column.PrimaryKey || column.Unique {
			columns = append(columns, column)
		}
	}
	return columns
}

type Relation struct {
	Column        *Column
	ForeignColumn *Column
}

func (m *Table) Relations() (relations []*Relation) {
	for _, column := range m.Columns {
		if column.Relation != nil {
			relations = append(relations, &Relation{
				Column:        column,
				ForeignColumn: column.Relation,
			})
		}
	}
	return relations
}

func (m *Table) RelationChains() (chains [][]*Relation) {
	relations := m.Relations()

	var all_nested [][]*Relation
	for _, relation := range relations {
		all_nested = append(all_nested,
			relation.ForeignColumn.Table.RelationChains()...)
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
	Tables []*Table
}
