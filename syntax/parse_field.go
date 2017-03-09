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

package syntax

import (
	"strconv"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
)

func parseField(node *tupleNode) (*ast.Field, error) {
	field := new(ast.Field)
	field.Pos = node.getPos()

	field_name_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	field.Name = stringFromToken(field_name_token)

	field.Type, field.Relation, err = parseFieldType(node)
	if err != nil {
		return nil, err
	}

	if field.Relation != nil {
		err = parseRelation(node, field)
		if err != nil {
			return nil, err
		}
		return field, nil
	}

	attributes_list := node.consumeIfList()
	if attributes_list != nil {
		err := attributes_list.consumeAnyTuples(tupleCases{
			"column": func(node *tupleNode) error {
				if field.Column != nil {
					return previouslyDefined(node.getPos(), "field", "column",
						field.Column.Pos)
				}

				name_token, err := node.consumeToken(Ident)
				if err != nil {
					return err
				}
				field.Column = stringFromToken(name_token)

				return nil
			},
			"nullable": tupleFlagField("field", "nullable",
				&field.Nullable),
			"autoinsert": tupleFlagField("field", "autoinsert",
				&field.AutoInsert),
			"autoupdate": tupleFlagField("field", "autoupdate",
				&field.AutoUpdate),
			"updatable": tupleFlagField("field", "updatable",
				&field.Updatable),
			"length": func(node *tupleNode) error {
				if field.Length != nil {
					return previouslyDefined(node.getPos(), "field", "length",
						field.Length.Pos)
				}

				length_token, err := node.consumeToken(Int)
				if err != nil {
					return err
				}
				value, err := strconv.Atoi(length_token.text)
				if err != nil {
					return errutil.New(length_token.getPos(),
						"unable to parse integer %q: %v",
						length_token.text, err)
				}
				field.Length = intFromValue(length_token, value)

				return nil
			},

			// TODO(jeff): do something with these values instead of allowing
			// anything.
			"default":    debugConsume,
			"sqldefault": debugConsume,
		})
		if err != nil {
			return nil, err
		}
	}

	if node.assertEmpty() != nil {
		invalid, _ := node.consume()
		return nil, errutil.New(invalid.getPos(),
			"expected list or end of tuple. got %s", invalid)
	}

	return field, nil
}

func parseFieldType(node *tupleNode) (*ast.FieldType, *ast.FieldRef, error) {
	first, second, err := node.consumeDottedIdents()
	if err != nil {
		return nil, nil, err
	}

	if second == nil {
		first.debugAssertToken(Ident)
		field_type, ok := fieldTypeMapping[first.text]
		if !ok {
			return nil, nil, expectedKeyword(
				first.getPos(), first.text, validFieldTypes()...)
		}
		return fieldTypeFromValue(first, field_type), nil, nil
	}

	return nil, &ast.FieldRef{
		Pos:   node.getPos(),
		Model: stringFromToken(first),
		Field: stringFromToken(second),
	}, nil
}
