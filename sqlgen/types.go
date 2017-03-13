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

package sqlgen

import (
	"bytes"
	"fmt"
)

type Literal string

func (l Literal) render(dialect Dialect) string { return string(l) }

func (l Literal) embedGolang() string {
	const format = prefix + "Literal(%q)"
	return fmt.Sprintf(format, string(l))
}

type Literals struct {
	Join string
	SQLs []SQL
}

func (l Literals) render(dialect Dialect) string {
	var out bytes.Buffer

	first := true
	for _, sql := range l.SQLs {
		if sql == nil {
			continue
		}
		if !first {
			out.WriteString(" ")
			out.WriteString(l.Join)
			out.WriteString(" ")
		}
		first = false
		out.WriteString(sql.render(dialect))
	}

	return out.String()
}

func (l Literals) embedGolang() string {
	var out bytes.Buffer

	const format = prefix + "Literals{Join:%q,SQLs:[]" + prefix + "SQL{"

	fmt.Fprintf(&out, format, l.Join)
	first := true
	for _, sql := range l.SQLs {
		if !first {
			out.WriteString(",")
		}
		first = false
		out.WriteString(sql.embedGolang())
	}
	out.WriteString("}}")

	return out.String()
}
