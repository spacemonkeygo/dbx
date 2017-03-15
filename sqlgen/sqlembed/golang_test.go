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

		&sqlgen.Hole{},
		&sqlgen.Hole{Name: "`"},
		&sqlgen.Hole{Name: `"`},

		// zero value
		sqlgen.Literals{},

		// no sqls
		sqlgen.Literals{Join: "foo"},
		sqlgen.Literals{Join: "`"},
		sqlgen.Literals{Join: `"`},

		// simple sqls
		sqlgen.Literals{Join: "bar", SQLs: []sqlgen.SQL{
			sqlgen.Literal("foo baz"),
			sqlgen.Literal("another"),
		}},

		// hard sqls
		sqlgen.Literals{Join: "bar", SQLs: []sqlgen.SQL{
			sqlgen.Literals{},
		}},
		sqlgen.Literals{Join: "recursive", SQLs: []sqlgen.SQL{
			sqlgen.Literals{Join: "bif", SQLs: []sqlgen.SQL{
				sqlgen.Literals{},
			}},
		}},
	}
	for i, test := range tests {
		emb := Golang("prefix_", test)
		if _, err := parser.ParseExpr(emb); err != nil {
			tw.Errorf("%d: %s but got error: %v", i, emb, err)
		}
	}
}

func testGolangFuzz(tw *testutil.T) {
	g := sqltest.NewGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.Gen()
		emb := Golang("prefix_", sql)

		if _, err := parser.ParseExpr(emb); err != nil {
			tw.Logf("sql: %#v", sql)
			tw.Logf("emb: %s", emb)
			tw.Logf("err: %v", err)
			tw.Error()
		}
	}
}
