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

import "gopkg.in/spacemonkeygo/dbx.v1/ast"

func parseModel(node *tupleNode) (*ast.Model, error) {
	model := new(ast.Model)
	model.Pos = node.getPos()

	name_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	model.Name = stringFromToken(name_token)

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"table": func(node *tupleNode) error {
			if model.Table != nil {
				return previouslyDefined(node.getPos(), "model", "table",
					model.Table.Pos)
			}

			name_token, err := node.consumeToken(Ident)
			if err != nil {
				return err
			}
			model.Table = stringFromToken(name_token)

			return nil
		},
		"field": func(node *tupleNode) error {
			field, err := parseField(node)
			if err != nil {
				return err
			}
			model.Fields = append(model.Fields, field)

			return nil
		},
		"key": func(node *tupleNode) error {
			if model.PrimaryKey != nil {
				return previouslyDefined(node.getPos(), "model", "key",
					model.PrimaryKey.Pos)
			}
			primary_key, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			model.PrimaryKey = primary_key
			return nil
		},
		"unique": func(node *tupleNode) error {
			unique, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			model.Unique = append(model.Unique, unique)
			return nil
		},
		"index": func(node *tupleNode) error {
			index, err := parseIndex(node)
			if err != nil {
				return err
			}
			model.Indexes = append(model.Indexes, index)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return model, nil
}
