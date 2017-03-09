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

func parseIndex(node *tupleNode) (*ast.Index, error) {
	index := new(ast.Index)
	index.Pos = node.getPos()

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"name": func(node *tupleNode) error {
			if index.Name != nil {
				return previouslyDefined(node.getPos(), "index", "name",
					index.Name.Pos)
			}

			name_token, err := node.consumeToken(Ident)
			if err != nil {
				return err
			}
			index.Name = stringFromToken(name_token)

			return nil
		},
		"fields": func(node *tupleNode) error {
			if index.Fields != nil {
				return previouslyDefined(node.getPos(), "index", "fields",
					index.Fields.Pos)
			}

			fields, err := parseRelativeFieldRefs(node)
			if err != nil {
				return err
			}
			index.Fields = fields

			return nil
		},
		"unique": tupleFlagField("index", "unique", &index.Unique),
	})
	if err != nil {
		return nil, err
	}

	return index, nil
}
