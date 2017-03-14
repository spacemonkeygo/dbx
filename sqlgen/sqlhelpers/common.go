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

const Param = sqlgen.Literal("?")

func Append(begin sqlgen.SQL, suffix ...sqlgen.SQL) sqlgen.SQL {
	var joined []sqlgen.SQL
	if begin != nil {
		joined = append(joined, begin)
	}
	joined = append(joined, suffix...)

	return sqlgen.Literals{
		Join: " ",
		SQLs: joined,
	}
}

func L(sql string) sqlgen.SQL {
	return sqlgen.Literal(sql)
}

func Lf(sqlf string, args ...interface{}) sqlgen.SQL {
	return sqlgen.Literal(fmt.Sprintf(sqlf, args...))
}

func Ls(with string, sqls ...sqlgen.SQL) sqlgen.SQL {
	return sqlgen.Literals{Join: with, SQLs: sqls}
}
