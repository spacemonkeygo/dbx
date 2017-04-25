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

func parseWhere(node *tupleNode) (where *ast.Where, err error) {
	where = new(ast.Where)
	where.Pos = node.getPos()

	where.Left, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	err = node.consumeTokenNamed(tokenCases{
		Exclamation.tokenCase(): func(token *tokenNode) error {
			_, err := node.consumeToken(Equal)
			if err != nil {
				return err
			}
			where.Op = operatorFromValue(token, consts.NE)
			return nil
		},
		{Ident, "like"}: func(token *tokenNode) error {
			where.Op = operatorFromValue(token, consts.Like)
			return nil
		},
		Equal.tokenCase(): func(token *tokenNode) error {
			where.Op = operatorFromValue(token, consts.EQ)
			return nil
		},
		LeftAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				where.Op = operatorFromValue(token, consts.LE)
			} else {
				where.Op = operatorFromValue(token, consts.LT)
			}
			return nil
		},
		RightAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				where.Op = operatorFromValue(token, consts.GE)
			} else {
				where.Op = operatorFromValue(token, consts.GT)
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	where.Right, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	return where, nil
}
