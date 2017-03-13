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
	"go/parser"
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestTypes(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Run("render", testTypesRender)
	tw.Run("embed golang", testTypesEmbedGolang)
}

func testTypesRender(tw *testutil.T) {
	type renderTestCase struct {
		in  SQL
		out string
	}

	tests := []renderTestCase{
		{in: Literal(""), out: ""},
		{in: Literal("foo bar sql"), out: "foo bar sql"},
		{in: Literal("`"), out: "`"},
		{in: Literal(`"`), out: "\""},
		{in: Literals{}, out: ""},
		{in: Literals{Join: "foo"}, out: ""},
		{in: Literals{Join: "`"}, out: ""},
		{in: Literals{Join: `"`}, out: ""},
		{
			in: Literals{Join: "bar", SQLs: []SQL{
				Literal("foo baz"),
				Literal("another"),
			}},
			out: "foo baz bar another",
		},
		{
			in: Literals{Join: "bar", SQLs: []SQL{
				Literal("inside first"),
				Literals{},
				Literal("inside second"),
			}},
			out: "inside first bar  bar inside second",
		},
		{
			in: Literals{Join: "recursive", SQLs: []SQL{
				Literals{Join: "bif", SQLs: []SQL{
					Literals{},
					Literal("inside"),
				}},
				Literal("outside"),
			}},
			out: " bif inside recursive outside",
		},
	}
	for i, test := range tests {
		if got := test.in.render(SQLite3()); got != test.out {
			tw.Errorf("%d: %q != %q", i, got, test.out)
		}
	}
}

func testTypesEmbedGolang(tw *testutil.T) {
	tests := []SQL{
		Literal(""),
		Literal("foo bar sql"),
		Literal("`"),
		Literal(`"`),

		// zero value
		Literals{},

		// no sqls
		Literals{Join: "foo"},
		Literals{Join: "`"},
		Literals{Join: `"`},

		// simple sqls
		Literals{Join: "bar", SQLs: []SQL{
			Literal("foo baz"),
			Literal("another"),
		}},

		// hard sqls
		Literals{Join: "bar", SQLs: []SQL{
			Literals{},
		}},
		Literals{Join: "recursive", SQLs: []SQL{
			Literals{Join: "bif", SQLs: []SQL{
				Literals{},
			}},
		}},
	}
	for i, test := range tests {
		emb := test.embedGolang()
		if _, err := parser.ParseExpr(emb); err != nil {
			tw.Errorf("%d: %s but got error: %v", i, emb, err)
		}
	}
}
