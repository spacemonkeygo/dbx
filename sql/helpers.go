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

package sql

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

type renderOp int

const (
	noFlatten renderOp = iota
	noTerminate
)

func render(dialect Dialect, s string, param interface{}, ops ...renderOp) (
	out string) {

	out = mustRender(s, param)

	flatten := true
	terminate := true
	for _, op := range ops {
		switch op {
		case noFlatten:
			flatten = false
		case noTerminate:
			terminate = false
		}
	}

	if flatten {
		out = flattenSQL(out)
	}
	if terminate {
		out = out + ";"
	}

	return dialect.Rebind(out)
}

func mustRender(s string, param interface{}) string {
	var buf bytes.Buffer
	tmpl := template.Must(template.New("").Parse(s))
	err := tmpl.Execute(&buf, param)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

var reSpace = regexp.MustCompile(`\s+`)

func flattenSQL(s string) string {
	return strings.TrimSpace(reSpace.ReplaceAllString(s, " "))
}
