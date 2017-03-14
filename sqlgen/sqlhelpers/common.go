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

package sqlhelpers

import (
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
)

// ls is the basic primitive for constructing larger SQLs. The first argument
// may be nil, and the result is a Literals.
func ls(sql sqlgen.SQL, join string, sqls ...sqlgen.SQL) sqlgen.SQL {
	var joined []sqlgen.SQL
	if sql != nil {
		joined = append(joined, sql)
	}
	joined = append(joined, sqls...)

	return sqlgen.Literals{Join: join, SQLs: joined}
}

// P is a placeholder literal
const P = sqlgen.Literal("?")

// L constructs a Literal
func L(sql string) sqlgen.SQL {
	return sqlgen.Literal(sql)
}

// Lf constructs a literal from a format string
func Lf(sqlf string, args ...interface{}) sqlgen.SQL {
	return sqlgen.Literal(fmt.Sprintf(sqlf, args...))
}

// J constructs a SQL by joining the given sqls with the string.
func J(join string, sqls ...sqlgen.SQL) sqlgen.SQL {
	return ls(nil, join, sqls...)
}

// Strings turns a slice of strings into a slice of Literal.
func Strings(parts []string) (out []sqlgen.SQL) {
	for _, part := range parts {
		out = append(out, sqlgen.Literal(part))
	}
	return out
}

// Placeholders returns a slice of placeholder literals of the right size.
func Placeholders(n int) (out []sqlgen.SQL) {
	for i := 0; i < n; i++ {
		out = append(out, P)
	}
	return out
}

//
// Builder constructs larger SQL statements by joining in pieces with spaces.
//

type Builder struct {
	sql sqlgen.SQL
}

func Build(sql sqlgen.SQL) *Builder {
	return &Builder{
		sql: sql,
	}
}

func (b *Builder) Add(sqls ...sqlgen.SQL) {
	b.sql = ls(b.sql, " ", sqls...)
}

func (b *Builder) SQL() sqlgen.SQL {
	return b.sql
}
