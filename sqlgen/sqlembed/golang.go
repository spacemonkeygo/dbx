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

package sqlembed

import (
	"bytes"
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
)

func Golang(prefix string, sql sqlgen.SQL) string {
	switch sql := sql.(type) {
	case sqlgen.Literal:
		return golangLiteral(prefix, sql)

	case sqlgen.Literals:
		return golangLiterals(prefix, sql)

	case *sqlgen.Hole:
		return golangHole(prefix, sql)

	default:
		panic("unhandled sql type")
	}
}

func golangLiteral(prefix string, sql sqlgen.Literal) string {
	const format = "%sLiteral(%q)"
	return fmt.Sprintf(format, prefix, string(sql))
}

func golangLiterals(prefix string, sql sqlgen.Literals) string {
	const format = "%[1]sLiterals{Join:%[2]q,SQLs:[]%[1]sSQL{"

	var out bytes.Buffer
	fmt.Fprintf(&out, format, prefix, sql.Join)

	first := true
	for _, sql := range sql.SQLs {
		if !first {
			out.WriteString(",")
		}
		first = false
		out.WriteString(Golang(prefix, sql))
	}
	out.WriteString("}}")

	return out.String()
}

func golangHole(prefix string, sql *sqlgen.Hole) string {
	const format = "&%sHole{Name:%q}"
	return fmt.Sprintf(format, prefix, sql.Name)
}
