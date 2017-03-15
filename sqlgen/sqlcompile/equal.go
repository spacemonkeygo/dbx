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

package sqlcompile

import "gopkg.in/spacemonkeygo/dbx.v1/sqlgen"

func sqlEqual(a, b sqlgen.SQL) bool {
	switch a := a.(type) {
	case sqlgen.Literal:
		if b, ok := b.(sqlgen.Literal); ok {
			return a == b
		}
		return false

	case sqlgen.Literals:
		if b, ok := b.(sqlgen.Literals); ok {
			return a.Join == b.Join && sqlsEqual(a.SQLs, b.SQLs)
		}
		return false

	case *sqlgen.Condition:
		if b, ok := b.(*sqlgen.Condition); ok {
			return a == b // pointer equality is correct
		}
		return false

	default:
		panic("unhandled sql type")
	}
}

func sqlsEqual(as, bs []sqlgen.SQL) bool {
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if !sqlEqual(as[i], bs[i]) {
			return false
		}
	}
	return true
}

func sqlNormalForm(sql sqlgen.SQL) bool {
	switch sql := sql.(type) {
	case sqlgen.Literal, *sqlgen.Condition:
		return true

	case sqlgen.Literals:
		if sql.Join != "" {
			return false
		}

		// only allow Hole and Literal and not two Literal in a row.

		last := ""

		for _, sql := range sql.SQLs {
			switch sql.(type) {
			case *sqlgen.Condition:
				last = "condition"

			case sqlgen.Literal:
				if last == "literal" {
					return false
				}
				last = "literal"

			case sqlgen.Literals:
				return false

			default:
				panic("unhandled sql type")
			}
		}

		return true

	default:
		panic("unhandled sql type")
	}
}
