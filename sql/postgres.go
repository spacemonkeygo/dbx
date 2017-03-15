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
	"strconv"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

type postgres struct {
}

func Postgres() Dialect {
	return &postgres{}
}

func (p *postgres) Name() string {
	return "postgres"
}

func (p *postgres) Features() Features {
	return Features{
		Returning:           true,
		PositionalArguments: true,
		NoLimitToken:        "ALL",
	}
}

func (p *postgres) RowId() string {
	return ""
}

func (p *postgres) ColumnType(field *ir.Field) string {
	switch field.Type {
	case consts.SerialField:
		return "serial"
	case consts.Serial64Field:
		return "bigserial"
	case consts.IntField:
		return "integer"
	case consts.Int64Field:
		return "bigint"
	case consts.UintField:
		return "integer"
	case consts.Uint64Field:
		return "bigint"
	case consts.FloatField:
		return "real"
	case consts.Float64Field:
		return "double precision"
	case consts.TextField:
		if field.Length > 0 {
			return fmt.Sprintf("varchar(%d)", field.Length)
		}
		return "text"
	case consts.BoolField:
		return "boolean"
	case consts.TimestampField:
		return "timestamp with time zone"
	case consts.TimestampUTCField:
		return "timestamp"
	case consts.BlobField:
		return "bytea"
	case consts.DateField:
		return "date"
	default:
		panic(fmt.Sprintf("unhandled field type %s", field.Type))
	}
}

func (p *postgres) Rebind(sql string) string {
	out := make([]byte, 0, len(sql)+10)

	j := 1
	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		if ch != '?' {
			out = append(out, ch)
			continue
		}

		out = append(out, '$')
		out = append(out, strconv.Itoa(j)...)
		j++
	}

	return string(out)
}

func (p *postgres) ArgumentPrefix() string { return "$" }
func (p *postgres) ExecOnOpen() []string   { return nil }
