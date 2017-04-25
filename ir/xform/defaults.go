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

package xform

import (
	"fmt"
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func DefaultIndexName(i *ir.Index) string {
	parts := []string{i.Model.Table}
	for _, field := range i.Fields {
		parts = append(parts, field.Column)
	}
	if i.Unique {
		parts = append(parts, "unique")
	}
	parts = append(parts, "index")
	return strings.Join(parts, "_")
}

func DefaultCreateSuffix(cre *ir.Create) []string {
	var parts []string
	parts = append(parts, cre.Model.Name)
	return parts
}

func DefaultReadSuffix(read *ir.Read) []string {
	var parts []string
	for _, selectable := range read.Selectables {
		switch obj := selectable.(type) {
		case *ir.Model:
			parts = append(parts, obj.Name)
		case *ir.Field:
			parts = append(parts, obj.Model.Name)
			parts = append(parts, obj.Name)
		default:
			panic(fmt.Sprintf("unhandled selectable %T", selectable))
		}
	}
	full := len(read.Joins) > 0
	parts = append(parts, whereSuffix(read.Where, full)...)
	if read.OrderBy != nil {
		parts = append(parts, "order_by")
		if read.OrderBy.Descending {
			parts = append(parts, "desc")
		} else {
			parts = append(parts, "asc")
		}
		for _, field := range read.OrderBy.Fields {
			if full {
				parts = append(parts, field.Model.Name)
			}
			parts = append(parts, field.Name)
		}
	}
	return parts
}

func DefaultUpdateSuffix(upd *ir.Update) []string {
	var parts []string
	parts = append(parts, upd.Model.Name)
	parts = append(parts, whereSuffix(upd.Where, len(upd.Joins) > 0)...)
	return parts
}

func DefaultDeleteSuffix(del *ir.Delete) []string {
	var parts []string
	parts = append(parts, del.Model.Name)
	parts = append(parts, whereSuffix(del.Where, len(del.Joins) > 0)...)
	return parts
}

func whereSuffix(wheres []*ir.Where, full bool) (parts []string) {
	if len(wheres) == 0 {
		return nil
	}
	parts = append(parts, "by")
	for i, where := range wheres {
		if i > 0 {
			parts = append(parts, "and")
		}

		left := exprSuffix(where.Left, full)
		right := exprSuffix(where.Right, full)

		parts = append(parts, left...)
		if len(right) > 0 || where.Op != consts.EQ {
			op := where.Op.Suffix()
			nulloperand := where.Left.Null || where.Right.Null
			switch where.Op {
			case consts.EQ:
				if nulloperand {
					parts = append(parts, "is")
				} else {
					parts = append(parts, op)
				}
			case consts.NE:
				if nulloperand {
					parts = append(parts, "is not")
				} else {
					parts = append(parts, op)
				}
			default:
				parts = append(parts, op)
			}
		}
		if len(right) > 0 {
			parts = append(parts, right...)
		}

	}
	return parts
}

func exprSuffix(expr *ir.Expr, full bool) (parts []string) {
	switch {
	case expr.Null:
		parts = []string{"null"}
	case expr.StringLit != nil:
		parts = []string{"literal"}
	case expr.NumberLit != nil:
		parts = []string{"literal"}
	case expr.Placeholder:
	case expr.Field != nil:
		if full {
			parts = append(parts, expr.Field.Model.Name)
		}
		parts = append(parts, expr.Field.Name)
	case expr.FuncCall != nil:
		parts = append(parts, expr.FuncCall.Name)
		for i, arg := range expr.FuncCall.Args {
			arg_suffix := exprSuffix(arg, full)
			if len(arg_suffix) == 0 {
				continue
			}
			if i != 0 {
				parts = append(parts, "and")
			}
			parts = append(parts, arg_suffix...)
		}
	}
	return parts
}
