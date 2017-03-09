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
	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func transformField(lookup *lookup, field_entry *fieldEntry) (err error) {
	field := field_entry.field
	ast_field := field_entry.ast

	field.Name = ast_field.Name.Value
	field.Column = ast_field.Column.Get()
	field.Nullable = ast_field.Nullable.Get()
	field.Updatable = ast_field.Updatable.Get()
	field.AutoInsert = ast_field.AutoInsert.Get()
	field.AutoUpdate = ast_field.AutoUpdate.Get()
	field.Length = ast_field.Length.Get()

	if field.AutoUpdate {
		field.Updatable = true
	}

	if ast_field.Relation != nil {
		related, err := lookup.FindField(ast_field.Relation)
		if err != nil {
			return err
		}
		relation_kind := ast_field.RelationKind.Value

		if relation_kind == consts.SetNull && !field.Nullable {
			return errutil.New(ast_field.Pos,
				"setnull relationships must be nullable")
		}

		field.Relation = &ir.Relation{
			Field: related,
			Kind:  relation_kind,
		}
		field.Type = related.Type.AsLink()
	} else {
		field.Type = ast_field.Type.Value
	}

	if ast_field.AutoUpdate != nil && !podFields[field.Type] {
		return errutil.New(ast_field.AutoInsert.Pos,
			"autoinsert must be on plain data type")
	}
	if ast_field.AutoUpdate != nil && !podFields[field.Type] {
		return errutil.New(ast_field.AutoUpdate.Pos,
			"autoupdate must be on plain data type")
	}
	if ast_field.Length != nil && field.Type != consts.TextField {
		return errutil.New(ast_field.Length.Pos,
			"length must be on a text field")
	}

	if field.Column == "" {
		field.Column = field.Name
	}

	return nil
}

var podFields = map[consts.FieldType]bool{
	consts.IntField:          true,
	consts.Int64Field:        true,
	consts.UintField:         true,
	consts.Uint64Field:       true,
	consts.BoolField:         true,
	consts.TextField:         true,
	consts.TimestampField:    true,
	consts.TimestampUTCField: true,
	consts.FloatField:        true,
	consts.Float64Field:      true,
	consts.DateField:         true,
}
