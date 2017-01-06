// Copyright (C) 2016 Space Monkey, Inc.
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

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
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
		Returning:          false,
		NeedsLimitOnOffset: true,
		NoLimitToken:       "-1",
	}
}

func (s *sqlite3) RowId() string {
	return "_rowid_"
}

func (s *sqlite3) ColumnType(field *ir.Field) string {
	switch field.Type {
	case ast.SerialField, ast.Serial64Field,
		ast.IntField, ast.Int64Field,
		ast.UintField, ast.Uint64Field:
		return "INTEGER"
	case ast.FloatField, ast.Float64Field:
		return "REAL"
	case ast.TextField:
		return "TEXT"
	case ast.BoolField:
		return "INTEGER"
	case ast.TimestampField, ast.TimestampUTCField:
		return "TIMESTAMP"
	case ast.BlobField:
		return "BLOB"
	default:
		panic(fmt.Sprintf("unhandled field type %s", field.Type))
	}
}

func (s *sqlite3) Rebind(sql string) string {
	return sql
}

func (s *sqlite3) ArgumentPrefix() string { return "?" }
