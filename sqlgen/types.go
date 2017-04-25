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

import "bytes"

type Literal string

func (Literal) private() {}

func (l Literal) Render() string { return string(l) }

type Literals struct {
	Join string
	SQLs []SQL
}

func (Literals) private() {}

func (l Literals) Render() string {
	var out bytes.Buffer

	first := true
	for _, sql := range l.SQLs {
		if sql == nil {
			continue
		}
		if !first {
			out.WriteString(l.Join)
		}
		first = false
		out.WriteString(sql.Render())
	}

	return out.String()
}

type Condition struct {
	// set at compile/embed time
	Name  string
	Left  string
	Equal bool
	Right string

	// set at runtime
	Null bool
}

func (*Condition) private() {}

func (c *Condition) Render() string {
	// TODO(jeff): maybe check if we can use placeholders instead of the
	// literal null: this would make the templates easier.

	switch {
	case c.Equal && c.Null:
		return c.Left + " is null"
	case c.Equal && !c.Null:
		return c.Left + " = " + c.Right
	case !c.Equal && c.Null:
		return c.Left + " is not null"
	case !c.Equal && !c.Null:
		return c.Left + " != " + c.Right
	default:
		panic("unhandled case")
	}
}

type Hole struct {
	// set at compiile/embed time
	Name string

	// set at runtime
	SQL SQL
}

func (*Hole) private() {}

func (h *Hole) Render() string { return h.SQL.Render() }
