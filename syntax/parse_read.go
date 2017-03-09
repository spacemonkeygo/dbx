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

func parseRead(node *tupleNode) (*ast.Read, error) {
	read := new(ast.Read)
	read.Pos = node.getPos()

	view, err := parseView(node)
	if err != nil {
		return nil, err
	}
	read.View = view

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"select": func(node *tupleNode) error {
			if read.Select != nil {
				return previouslyDefined(node.getPos(), "read", "select",
					read.Select.Pos)
			}

			refs, err := parseFieldRefs(node, false)
			if err != nil {
				return err
			}
			read.Select = refs

			return nil
		},
		"where": func(node *tupleNode) error {
			where, err := parseWhere(node)
			if err != nil {
				return err
			}
			read.Where = append(read.Where, where)

			return nil
		},
		"join": func(node *tupleNode) error {
			join, err := parseJoin(node)
			if err != nil {
				return err
			}
			read.Joins = append(read.Joins, join)

			return nil
		},
		"orderby": func(node *tupleNode) error {
			if read.OrderBy != nil {
				return previouslyDefined(node.getPos(), "read", "orderby",
					read.OrderBy.Pos)
			}

			order_by, err := parseOrderBy(node)
			if err != nil {
				return err
			}
			read.OrderBy = order_by

			return nil
		},
		"suffix": func(node *tupleNode) error {
			if read.Suffix != nil {
				return previouslyDefined(node.getPos(), "read", "suffix",
					read.Suffix.Pos)
			}

			suffix, err := parseSuffix(node)
			if err != nil {
				return err
			}
			read.Suffix = suffix

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return read, nil
}
