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

func transformField(lookup *lookup, field_entry *fieldEntry) (err error) {
	field := field_entry.field
	ast_field := field_entry.ast

	field.Name = ast_field.Name
	field.Type = ast_field.Type
	field.Column = ast_field.Column
	field.Nullable = ast_field.Nullable
	field.Updatable = ast_field.Updatable
	field.AutoInsert = ast_field.AutoInsert
	field.AutoUpdate = ast_field.AutoUpdate
	field.Length = ast_field.Length

	if ast_field.Relation != nil {
		related, err := lookup.FindField(ast_field.Relation.FieldRef)
		if err != nil {
			return err
		}
		field.Relation = &Relation{
			Field: related,
		}
		field.Type = related.Type.AsLink()
	}

	return nil
}
