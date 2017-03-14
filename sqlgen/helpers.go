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

import "fmt"

const Param = Literal("?")

func Append(begin SQL, suffix ...SQL) SQL {
	if blits, ok := begin.(Literals); ok && blits.Join == " " {
		blits.SQLs = append(blits.SQLs, suffix...)
		return blits
	}

	var joined []SQL
	if begin != nil {
		joined = append(joined, begin)
	}
	joined = append(joined, suffix...)

	return Literals{
		Join: " ",
		SQLs: joined,
	}
}

func L(sql string) SQL {
	return Literal(sql)
}

func Lf(sqlf string, args ...interface{}) SQL {
	return Literal(fmt.Sprintf(sqlf, args...))
}

func Join(with string, sqls ...SQL) SQL {
	return Literals{Join: with, SQLs: sqls}
}
