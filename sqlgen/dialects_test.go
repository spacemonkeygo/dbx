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

type rebindTestCase struct {
	in  string
	out string
}

func testDialectsRebind(tw *testutil.T, dialect Dialect,
	tests []rebindTestCase) {

	for i, test := range tests {
		if got := dialect.Rebind(test.in); got != test.out {
			tw.Errorf("%d: %q != %q", i, got, test.out)
		}
	}
}

func TestDialects(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Run("postgres", testDialectsPostgres)
	tw.Run("sqlite3", testDialectsSQLite3)
}

func testDialectsPostgres(tw *testutil.T) {
	tw.Run("rebind", testDialectsPostgresRebind)
}

func testDialectsPostgresRebind(tw *testutil.T) {
	testDialectsRebind(tw, Postgres(), []rebindTestCase{
		{in: "", out: ""},
		{in: "? foo bar ? baz", out: "$1 foo bar $2 baz"},
		{in: "? ? ?", out: "$1 $2 $3"},
	})
}

func testDialectsSQLite3(tw *testutil.T) {
	tw.Run("rebind", testDialectsSQLite3Rebind)
}

func testDialectsSQLite3Rebind(tw *testutil.T) {
	testDialectsRebind(tw, SQLite3(), []rebindTestCase{
		{in: "", out: ""},
		{in: "? foo bar ? baz", out: "? foo bar ? baz"},
		{in: "? ? ?", out: "? ? ?"},
	})
}
