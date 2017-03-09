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

func parseDelete(node *tupleNode) (*ast.Delete, error) {
	del := new(ast.Delete)
	del.Pos = node.getPos()

	model_ref_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	del.Model = modelRefFromToken(model_ref_token)

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"where": func(node *tupleNode) error {
			where, err := parseWhere(node)
			if err != nil {
				return err
			}
			del.Where = append(del.Where, where)

			return nil
		},
		"join": func(node *tupleNode) error {
			join, err := parseJoin(node)
			if err != nil {
				return err
			}
			del.Joins = append(del.Joins, join)

			return nil
		},
		"suffix": func(node *tupleNode) error {
			if del.Suffix != nil {
				return previouslyDefined(node.getPos(), "delete", "suffix",
					del.Suffix.Pos)
			}

			suffix, err := parseSuffix(node)
			if err != nil {
				return err
			}
			del.Suffix = suffix

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return del, nil
}
