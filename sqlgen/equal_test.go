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

func TestEqual(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("fuzz identity", testEqualFuzzIdentity)
	tw.Runp("normal form", testEqualNormalForm)
}

func testEqualFuzzIdentity(tw *testutil.T) {
	g := newGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.gen()

		if !sqlEqual(sql, sql) {
			tw.Logf("sql: %#v", sql)
			tw.Error()
		}
	}
}

func testEqualNormalForm(tw *testutil.T) {
	type normalFormTestCase struct {
		in     SQL
		normal bool
	}

	tests := []normalFormTestCase{
		{in: Literal(""), normal: true},
		{in: new(Hole), normal: true},
		{in: Literals{}, normal: true},
		{in: Literals{Join: "foo"}, normal: false},
		{
			in: Literals{Join: "", SQLs: []SQL{
				Literal("foo baz"),
			}},
			normal: true,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				Literal("foo baz"),
				Literal("bif bar"),
			}},
			normal: false,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				new(Hole),
				Literal("foo baz"),
			}},
			normal: true,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				Literal("bif bar"),
				new(Hole),
			}},
			normal: true,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				Literal("foo baz"),
				new(Hole),
				Literal("bif bar"),
			}},
			normal: true,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				new(Hole),
				new(Hole),
				Literal("foo baz"),
			}},
			normal: true,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				Literal("bif bar"),
				new(Hole),
				new(Hole),
			}},
			normal: true,
		},
		{
			in: Literals{Join: "", SQLs: []SQL{
				Literal("foo baz"),
				new(Hole),
				new(Hole),
				Literal("bif bar"),
			}},
			normal: true,
		},
	}
	for i, test := range tests {
		if got := sqlNormalForm(test.in); got != test.normal {
			tw.Errorf("%d: got:%v != exp:%v. sql:%#v",
				i, got, test.normal, test.in)
		}
	}
}
