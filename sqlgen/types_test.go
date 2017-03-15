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
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestTypes(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("render", testTypesRender)
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

		{in: &Condition{Equal: false, Null: false}, out: "!= ?"},
		{in: &Condition{Equal: false, Null: true}, out: "is not null"},
		{in: &Condition{Equal: true, Null: false}, out: "= ?"},
		{in: &Condition{Equal: true, Null: true}, out: "is null"},
		{
			in:  &Condition{Field: "f", Equal: false, Null: false},
			out: "f != ?",
		},
		{
			in:  &Condition{Field: "f", Equal: false, Null: true},
			out: "f is not null",
		},
		{
			in:  &Condition{Field: "f", Equal: true, Null: false},
			out: "f = ?",
		},
		{
			in:  &Condition{Field: "f", Equal: true, Null: true},
			out: "f is null",
		},

		{in: Literals{}, out: ""},
		{in: Literals{Join: "foo"}, out: ""},
		{in: Literals{Join: "`"}, out: ""},
		{in: Literals{Join: `"`}, out: ""},
		{
			in: Literals{Join: " bar ", SQLs: []SQL{
				Literal("foo baz"),
				Literal("another"),
			}},
			out: "foo baz bar another",
		},
		{
			in: Literals{Join: " bar ", SQLs: []SQL{
				Literal("inside first"),
				Literals{},
				Literal("inside second"),
			}},
			out: "inside first bar  bar inside second",
		},
		{
			in: Literals{Join: " recursive ", SQLs: []SQL{
				Literals{Join: " bif ", SQLs: []SQL{
					Literals{},
					Literal("inside"),
				}},
				Literal("outside"),
			}},
			out: " bif inside recursive outside",
		},
	}
	for i, test := range tests {
		if got := test.in.Render(); got != test.out {
			tw.Errorf("%d: %q != %q", i, got, test.out)
		}
	}
}
