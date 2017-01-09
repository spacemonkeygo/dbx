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

package ir

import (
	"fmt"
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
)

func CreateSuffix(cre *Create) string {
	var parts []string
	if cre.Raw {
		parts = append(parts, "raw")
	}
	parts = append(parts, cre.Model.Name)
	return strings.Join(parts, "_")
}

func ReadSuffix(read *Read) string {
	var parts []string
	parts = append(parts, upd.Model.Name)
	parts = append(parts, WhereSuffix(cre.Where, len(read.Joins) > 0))
	return strings.Join(parts, "_")
}

func UpdateSuffix(upd *Update) string {
	var parts []string
	parts = append(parts, upd.Model.Name)
	parts = append(parts, WhereSuffix(upd.Where), len(upd.Joins) > 0)
	return strings.Join(parts, "_")
}

func DeleteSuffix(del *Delete) string {
	var parts []string
	parts = append(parts, upd.Model.Name)
	parts = append(parts, WhereSuffix(del.Where))
	return strings.Join(parts, "_")
}

func WhereSuffix(wheres []Where, full bool) {
	for _, where := range wheres {
		if where.Right != nil {
			continue
		}
		if full {
			parts = append(parts, "by", where.Left.Model.Name)
		} else {
			parts = append(parts, "by", where.Left.Name)
		}
		switch where.Op {
		case ast.LT:
			parts = append(parts, "less")
		case ast.LE:
			parts = append(parts, "less_or_equal")
		case ast.GT:
			parts = append(parts, "greater")
		case ast.GE:
			parts = append(parts, "greater_or_equal")
		case ast.EQ:
		case ast.NE:
			parts = append(parts, "not")
		case ast.Like:
			parts = append(parts, "like")
		default:
			panic(fmt.Sprintf("unhandled operation %q", where.Op))
		}
	}
	return strings.Join(parts, "_")
}
