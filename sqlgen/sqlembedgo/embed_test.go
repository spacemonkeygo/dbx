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
	"go/parser"
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlcompile"
	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqltest"
	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestGolang(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("basic types", testGolangBasicTypes)
	tw.Runp("fuzz", testGolangFuzz)
}

func testGolangBasicTypes(tw *testutil.T) {
	tests := []sqlgen.SQL{
		sqlgen.Literal(""),
		sqlgen.Literal("foo bar sql"),
		sqlgen.Literal("`"),
		sqlgen.Literal(`"`),

		&sqlgen.Condition{Name: "foo"},

		sqlgen.Literals{},
		sqlgen.Literals{Join: "foo"},
		sqlgen.Literals{Join: "`"},
		sqlgen.Literals{Join: `"`},
		sqlgen.Literals{Join: "bar", SQLs: []sqlgen.SQL{
			sqlgen.Literal("foo baz"),
			sqlgen.Literal("another"),
			&sqlgen.Condition{Name: "foo"},
		}},
	}
	for i, test := range tests {
		info := Embed("prefix_", test)
		if _, err := parser.ParseExpr(info.Expression); err != nil {
			tw.Errorf("%d: %+v but got error: %v", i, info, err)
		}
	}
}

func testGolangFuzz(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()
		compiled := sqlcompile.Compile(sql)
		info := Embed("prefix_", compiled)

		if _, err := parser.ParseExpr(info.Expression); err != nil {
			tw.Logf("sql:      %#v", sql)
			tw.Logf("compiled: %#v", compiled)
			tw.Logf("info:     %+v", info)
			tw.Logf("err:      %v", err)
			tw.Error()
		}
	}
}
