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

func TestCompile(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()
	tw.Runp("fuzz render", testCompileFuzzRender)
	tw.Runp("idempotent", testCompileIdempotent)
}

func testCompileFuzzRender(tw *testutil.T) {
	g := newGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.gen()
		compiled := sqlCompile(sql)
		exp := sql.render()
		got := compiled.render()

		if exp != got {
			tw.Logf("sql:      %#v", sql)
			tw.Logf("compiled: %#v", compiled)
			tw.Logf("exp:      %q", exp)
			tw.Logf("got:      %q", got)
			tw.Error()
		}
	}
}

func testCompileIdempotent(tw *testutil.T) {
	g := newGenerator(tw)

	for i := 0; i < 1000; i++ {
		sql := g.gen()
		first := sqlCompile(sql)
		second := sqlCompile(first)

		if !sqlEqual(first, second) {
			tw.Logf("sql:    %#v", sql)
			tw.Logf("first:  %#v", first)
			tw.Logf("second: %#v", second)
			tw.Error()
		}
	}
}
