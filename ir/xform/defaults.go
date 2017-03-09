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
		if where.Right != nil {
			continue
		}
		if i > 0 {
			parts = append(parts, "and")
		}
		if full {
			parts = append(parts, where.Left.Model.Name)
		}
		parts = append(parts, where.Left.Name)
		if suffix := where.Op.Suffix(); suffix != "" {
			parts = append(parts, suffix)
		}
	}
	return parts
}
