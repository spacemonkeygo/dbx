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
	"regexp"
	"strings"
)

type SQL interface {
	Render() string

	private()
}

type Dialect interface {
	Rebind(sql string) string
}

type RenderOp int

const (
	NoFlatten RenderOp = iota
	NoTerminate
)

func Render(dialect Dialect, sql SQL, ops ...RenderOp) string {
	out := sql.Render()

	flatten := true
	terminate := true
	for _, op := range ops {
		switch op {
		case NoFlatten:
			flatten = false
		case NoTerminate:
			terminate = false
		}
	}

	if flatten {
		out = flattenSQL(out)
	}
	if terminate {
		out += ";"
	}

	return dialect.Rebind(out)
}

var reSpace = regexp.MustCompile(`\s+`)

func flattenSQL(s string) string {
	return strings.TrimSpace(reSpace.ReplaceAllString(s, " "))
}
