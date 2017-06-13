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

package sql

import (
	"fmt"
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

type sqlite3 struct {
}

func SQLite3() Dialect {
	return &sqlite3{}
}

func (s *sqlite3) Name() string {
	return "sqlite3"
}

func (s *sqlite3) Features() Features {
	return Features{
		Returning:    false,
		NoLimitToken: "-1",
	}
}

func (s *sqlite3) RowId() string {
	return "_rowid_"
}

func (s *sqlite3) ColumnType(field *ir.Field) string {
	switch field.Type {
	case consts.SerialField, consts.Serial64Field,
		consts.IntField, consts.Int64Field,
		consts.UintField, consts.Uint64Field:
		return "INTEGER"
	case consts.FloatField, consts.Float64Field:
		return "REAL"
	case consts.TextField:
		return "TEXT"
	case consts.BoolField:
		return "INTEGER"
	case consts.TimestampField, consts.TimestampUTCField:
		return "TIMESTAMP"
	case consts.BlobField:
		return "BLOB"
	case consts.DateField:
		return "DATE"
	default:
		panic(fmt.Sprintf("unhandled field type %s", field.Type))
	}
}

func (s *sqlite3) Rebind(sql string) string {
	return sql
}

var sqlite3Replacer = strings.NewReplacer(
	`'`, `''`,
)

func (p *sqlite3) EscapeString(s string) string {
	return sqlite3Replacer.Replace(s)
}
