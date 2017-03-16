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

package sqlembedgo

import (
	"bytes"
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlbundle"
)

type Condition struct {
	Name       string
	Expression string
}

type Hole struct {
	Name       string
	Expression string
}

type Info struct {
	Expression string
	Conditions []Condition
	Holes      []Hole
}

func Embed(prefix string, sql sqlgen.SQL) Info {
	switch sql := sql.(type) {
	case sqlgen.Literal:
		return Info{
			Expression: golangLiteral(sql),
			Conditions: nil,
		}

	case sqlgen.Literals:
		return golangLiterals(prefix, sql)

	case *sqlgen.Condition:
		cond := golangCondition(prefix, sql)
		return Info{
			Expression: cond.Name,
			Conditions: []Condition{cond},
		}

	case *sqlgen.Hole:
		hole := golangHole(prefix, sql)
		return Info{
			Expression: hole.Name,
			Holes:      []Hole{hole},
		}

	default:
		panic("unhandled sql type")
	}
}

func golangLiteral(sql sqlgen.Literal) string {
	const format = "%[1]sLiteral(%[2]q)"

	return fmt.Sprintf(format, sqlbundle.Prefix, string(sql))
}

func golangLiterals(prefix string, sql sqlgen.Literals) (info Info) {
	const format = "%[1]sLiterals{Join:%[2]q,SQLs:[]%[1]sSQL{"

	var conds []Condition
	var holes []Hole
	var expr bytes.Buffer
	fmt.Fprintf(&expr, format, sqlbundle.Prefix, sql.Join)

	first := true
	for _, sql := range sql.SQLs {
		if !first {
			expr.WriteString(",")
		}
		first = false

		switch sql := sql.(type) {
		case sqlgen.Literal:
			expr.WriteString(golangLiteral(sql))

		case *sqlgen.Condition:
			cond := golangCondition(prefix, sql)
			expr.WriteString(cond.Name)

			// TODO(jeff): dedupe based on name?
			conds = append(conds, cond)

		case *sqlgen.Hole:
			hole := golangHole(prefix, sql)
			expr.WriteString(hole.Name)

			// TODO(jeff): dedupe based on name?
			holes = append(holes, hole)

		case sqlgen.Literals:
			panic("sql not in normal form")

		default:
			panic("unhandled sql type")
		}
	}
	expr.WriteString("}}")

	return Info{
		Expression: expr.String(),
		Conditions: conds,
		Holes:      holes,
	}
}

func golangCondition(prefix string, sql *sqlgen.Condition) Condition {
	const format = "&%[1]sCondition{Field:%q, Equal:%t}"

	return Condition{
		Name:       prefix + sql.Name,
		Expression: fmt.Sprintf(format, sqlbundle.Prefix, sql.Field, sql.Equal),
	}
}

func golangHole(prefix string, sql *sqlgen.Hole) Hole {
	const format = "&%[1]sHole{}"

	// TODO(jeff): embed what the hole is filled with? no use case yet.

	return Hole{
		Name:       prefix + sql.Name,
		Expression: fmt.Sprintf(format, sqlbundle.Prefix),
	}
}
