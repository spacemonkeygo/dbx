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

func Compile(sql SQL) SQL { return sqlCompile(sql) }

func sqlCompile(sql SQL) (out SQL) {
	switch sql := sql.(type) {
	case Literal: // a literal has nothing to do
		return sql
	case Literals:
		// if there are no SQLs, we just have an empty string so hoist to a
		// literal type
		if len(sql.SQLs) == 0 {
			return Literal("")
		}

		// if there is one sql, we can just return the compiled form of that.
		if len(sql.SQLs) == 1 {
			return sqlCompile(sql.SQLs[0])
		}

		// more than one sql. constant fold the inner sql's and then recompile
		// if that was able to do any work.
		folded := sqlConstantFold(sql.SQLs, sql.Join)
		if sqlsEqual(sql.SQLs, folded) {
			return sql
		}

		return sqlCompile(Literals{
			Join: sql.Join,
			SQLs: folded,
		})
	default:
		panic("unhandled sql type")
	}
}

func sqlConstantFold(sqls []SQL, join string) (out []SQL) {
	buf := Literals{Join: join}
	for _, sql := range sqls {
		sql = sqlCompile(sql)

		lit, ok := sql.(Literal)
		if ok {
			buf.SQLs = append(buf.SQLs, lit)
			continue
		}

		if len(buf.SQLs) > 0 {
			out = append(out, Literal(buf.render()))
			buf.SQLs = buf.SQLs[:0]
		}
		out = append(out, sql)
	}

	if len(buf.SQLs) > 0 {
		out = append(out, Literal(buf.render()))
	}

	return out
}
