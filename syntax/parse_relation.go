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
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/consts"
)

func parseRelation(node *tupleNode, field *ast.Field) error {
	err := node.consumeTokenNamed(tokenCases{
		{Ident, "setnull"}: func(token *tokenNode) error {
			field.RelationKind = relationKindFromValue(token, consts.SetNull)
			return nil
		},
		{Ident, "cascade"}: func(token *tokenNode) error {
			field.RelationKind = relationKindFromValue(token, consts.Cascade)
			return nil
		},
		{Ident, "restrict"}: func(token *tokenNode) error {
			field.RelationKind = relationKindFromValue(token, consts.Restrict)
			return nil
		},
	})
	if err != nil {
		return err
	}

	attributes_list := node.consumeIfList()
	if attributes_list != nil {
		err := attributes_list.consumeAnyTuples(tupleCases{
			"column": func(node *tupleNode) error {
				if field.Column != nil {
					return previouslyDefined(node.getPos(), "relation", "column",
						field.Column.Pos)
				}

				name_token, err := node.consumeToken(Ident)
				if err != nil {
					return err
				}
				field.Column = stringFromToken(name_token)

				return nil
			},
			"nullable": tupleFlagField("relation", "nullable",
				&field.Nullable),
			"updatable": tupleFlagField("relation", "updatable",
				&field.Updatable),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
