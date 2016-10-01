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

package dbx

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

func deserial(t string) string {
	switch t {
	case "serial":
		return "int"
	case "serial64":
		return "int64"
	default:
		return t
	}
}

func LoadSchema(path string) (schema *Schema, err error) {
	type YRelation struct {
		OwnedBy  string `yaml:"owned_by"`
		HasA     string `yaml:"has_a"`
		Name     string `yaml:"name"`
		Nullable bool   `yaml:"nullable"`
	}

	type YColumn struct {
		Name       string `yaml:"name"`
		Type       string `yaml:"type"`
		Nullable   bool   `yaml:"nullable"`
		AutoInsert bool   `yaml:"auto_insert"`
	}

	type YTable struct {
		Name       string      `yaml:"name"`
		Columns    []YColumn   `yaml:"columns"`
		Relations  []YRelation `yaml:"relations"`
		Unique     [][]string  `yaml:"unique"`
		PrimaryKey []string    `yaml:"primary_key"`
	}

	type YSchema struct {
		Tables  []YTable `yaml:"tables"`
		Queries [][][]string
	}

	yaml_bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var yschema YSchema
	err = yaml.Unmarshal(yaml_bytes, &yschema)
	if err != nil {
		return nil, err
	}

	splitdot := func(s string) (table, column string) {
		rsplit := strings.SplitN(s, ".", 2)
		if len(rsplit) < 2 {
			return rsplit[0], ""
		}
		return rsplit[0], rsplit[1]
	}

	tables := map[string]*Table{}
	// create tables and columns
	for _, ytable := range yschema.Tables {
		table := &Table{Name: ytable.Name}
		for _, ycolumn := range ytable.Columns {
			if table.GetColumn(ycolumn.Name) != nil {
				return nil, fmt.Errorf("%s.%s already defined",
					ytable.Name, ycolumn.Name)
			}
			table.Columns = append(table.Columns, &Column{
				Table:      table,
				Name:       ycolumn.Name,
				Type:       ycolumn.Type,
				NotNull:    !ycolumn.Nullable,
				AutoInsert: ycolumn.AutoInsert,
			})
		}
		tables[ytable.Name] = table
	}

	// resolve relations
	for _, ytable := range yschema.Tables {
		table := tables[ytable.Name]
		for _, yrelation := range ytable.Relations {
			var rtable string
			var rcolumn string
			var kind RelationKind
			switch {
			case yrelation.OwnedBy == "" && yrelation.HasA == "":
				return nil, fmt.Errorf(
					"empty relationship specified on %s relation",
					ytable.Name)
			case yrelation.OwnedBy != "" && yrelation.HasA != "":
				return nil, fmt.Errorf(
					"both has_a and owned_by specified on %s relation",
					ytable.Name)
			case yrelation.OwnedBy != "":
				kind = OwnedBy
				rtable, rcolumn = splitdot(yrelation.OwnedBy)
			case yrelation.HasA != "":
				kind = HasA
				rtable, rcolumn = splitdot(yrelation.HasA)
			}

			name := yrelation.Name
			if name == "" {
				name = fmt.Sprintf("%s_%s", rtable, rcolumn)
			}

			ftable := tables[rtable]
			if ftable == nil {
				return nil, fmt.Errorf("no table %s in relation %s.%s",
					rtable, ytable.Name, name)
			}
			fcolumn := ftable.GetColumn(rcolumn)
			if fcolumn == nil {
				return nil, fmt.Errorf("no column %s.%s in relation %s.%s",
					rcolumn, rtable,
					ytable.Name, name)
			}
			table.Columns = append(table.Columns, &Column{
				Table:   table,
				Name:    name,
				Type:    deserial(fcolumn.Type),
				NotNull: !yrelation.Nullable,
				Relation: &Relation{
					Column: fcolumn,
					Kind:   kind,
				},
			})
		}
	}

	// Primary key and unique's
	for _, ytable := range yschema.Tables {
		table := tables[ytable.Name]
		table.PrimaryKey = table.GetColumns(ytable.PrimaryKey...)
		if table.PrimaryKey == nil {
			return nil, fmt.Errorf(
				"table %s missing columns in primary key %s",
				ytable.Name, ytable.PrimaryKey)
		}
		for _, yunique := range ytable.Unique {
			unique := table.GetColumns(yunique...)
			if unique == nil {
				return nil, fmt.Errorf(
					"table %s missing columns in unique %s",
					ytable.Name, yunique)
			}
			table.Unique = append(table.Unique, unique)
		}
		tables[ytable.Name] = table
	}

	resolvedot := func(s string, need_column bool) (*Table, *Column, error) {
		t, c := splitdot(s)
		if t == "" {
			return nil, nil, fmt.Errorf("invalid dotname syntax %q", s)
		}
		table := tables[t]
		if table == nil {
			return nil, nil, fmt.Errorf("no such table %q in dotname %q", t, s)
		}
		if c == "" {
			if need_column {
				return nil, nil, fmt.Errorf(
					"missing required column in dotname %q", s)
			}
			return table, nil, nil
		}
		column := table.GetColumn(c)
		if column == nil {
			return nil, nil, fmt.Errorf("no such column in dotname %q", c, s)
		}
		return table, column, nil
	}

	schema = &Schema{}

	// create queries
	for _, yquery := range yschema.Queries {
		var ystarts []string
		var yjoins []string
		var yends []string

		switch len(yquery) {
		case 0:
			continue
		case 3:
			yends = yquery[2]
			fallthrough
		case 2:
			yjoins = yquery[1]
			fallthrough
		case 1:
			ystarts = yquery[0]
		default:
			return nil, fmt.Errorf("query %q has too many parts", yquery)
		}

		query := &Query{}
		var curtable *Table

		for _, ystart := range ystarts {
			table, column, err := resolvedot(ystart, false)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid dotname on query %+v: %v", yquery, err)
			}
			if curtable == nil {
				query.Table = table
				curtable = table
			} else if table != curtable {
				return nil, fmt.Errorf(
					"expected table %q on %q; got %q",
					curtable.Name, yquery, table.Name)
			}
			if column != nil {
				query.Start = append(query.Start, column)
			}
		}

		for _, yjoin := range yjoins {
			_, jcolumn, err := resolvedot(yjoin, true)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid dotname on join %+v: %v", yquery, err)
			}

			if curtable == nil {
				query.Table = jcolumn.Table
				curtable = jcolumn.Table
			}

			var relation *Join
			if jcolumn.Table == curtable {
				relation = jcolumn.RelationLeft()
			} else {
				relation = jcolumn.RelationRight()
			}
			if relation == nil {
				return nil, fmt.Errorf("missing relation on join %s",
					yjoin)
			}
			if relation.Left.Table != curtable {
				return nil, fmt.Errorf("incomplete table chain on %s",
					yquery)
			}
			curtable = relation.Right.Table
			query.Joins = append(query.Joins, relation)
		}

		for _, yend := range yends {
			table, column, err := resolvedot(yend, true)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid dotname on query %+v: %v", yquery, err)
			}
			if curtable != nil && table != curtable {
				return nil, fmt.Errorf("incomplete table chain on %s",
					yquery)
			}
			query.End = append(query.End, column)
		}

		schema.Queries = append(schema.Queries, query)
	}

	// In order to create tables in dependency order and also to reduce
	// generated code churn between runs, order the tables first by depth and
	// then alphabetically.
	for _, table := range tables {
		schema.Tables = append(schema.Tables, table)
	}
	sort.Sort(sortTable(schema.Tables))

	return schema, nil
}

type sortTable []*Table

func (by sortTable) Len() int {
	return len(by)
}

func (by sortTable) Swap(a, b int) {
	by[a], by[b] = by[b], by[a]
}

func (by sortTable) Less(a, b int) bool {
	adepth := by[a].Depth()
	bdepth := by[b].Depth()
	if adepth < bdepth {
		return true
	}
	if adepth > bdepth {
		return false
	}
	return by[a].Name < by[b].Name
}
