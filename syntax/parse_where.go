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
	"fmt"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/consts"
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
)

func parseWhere(node *tupleNode) (where *ast.Where, err error) {
	where = new(ast.Where)
	where.Pos = node.getPos()

	// either "and", "or", or "".
	compound_kind := ""
	var children []*ast.Where

	for {
		child := new(ast.Where)
		child.Pos = node.getPos()

		// if the first element of the tuple is a list, it should consist of a
		// single tuple and the result of parsing that tuple gives us an
		// *ast.Where that we should use.
		list := node.consumeIfList()
		if list != nil {
			node, err := list.consumeTuple()
			if err != nil {
				return nil, err
			}
			child, err = parseWhere(node)
			if err != nil {
				return nil, err
			}
			if err := node.assertEmpty(); err != nil {
				return nil, err
			}
		} else if len(node.value) > 0 {
			child.Clause, err = parseWhereClause(node)
			if err != nil {
				return nil, err
			}
		}

		// add the child to the children
		children = append(children, child)

		// if there are no more compound statements, we're done
		if len(node.value) == 0 {
			// if there was only one child, return the child itself as the
			// parsed where.
			if len(children) == 1 {
				return child, nil
			}

			// otherwise, append the child to the children, set the children on
			// the appropriate field of the where and return.
			switch compound_kind {
			case "and":
				where.And = children
			case "or":
				where.Or = children
			default:
				panic(fmt.Sprintf(
					"internal: compound_kind = %q", compound_kind))
			}
			return where, nil
		}

		// consume the compound joining token and ensure that it stays the same
		var token *tokenNode
		storeToken := func(consumed *tokenNode) error {
			token = consumed
			return nil
		}
		err = node.consumeTokenNamed(tokenCases{
			{Ident, "and"}: storeToken,
			{Ident, "or"}:  storeToken,
		})
		if err != nil {
			return nil, err
		}
		if compound_kind != "" && compound_kind != token.text {
			return nil, errutil.New(token.pos,
				"distinct compound statements must be grouped. "+
					"e.g. `(a and b) or c` instead of `a and b or c`")
		}
		compound_kind = token.text
	}
}

func parseWhereClause(node *tupleNode) (clause *ast.WhereClause, err error) {
	clause = new(ast.WhereClause)
	clause.Pos = node.getPos()

	clause.Left, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	err = node.consumeTokenNamed(tokenCases{
		Exclamation.tokenCase(): func(token *tokenNode) error {
			_, err := node.consumeToken(Equal)
			if err != nil {
				return err
			}
			clause.Op = operatorFromValue(token, consts.NE)
			return nil
		},
		{Ident, "like"}: func(token *tokenNode) error {
			clause.Op = operatorFromValue(token, consts.Like)
			return nil
		},
		Equal.tokenCase(): func(token *tokenNode) error {
			clause.Op = operatorFromValue(token, consts.EQ)
			return nil
		},
		LeftAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				clause.Op = operatorFromValue(token, consts.LE)
			} else {
				clause.Op = operatorFromValue(token, consts.LT)
			}
			return nil
		},
		RightAngle.tokenCase(): func(token *tokenNode) error {
			if node.consumeIfToken(Equal) != nil {
				clause.Op = operatorFromValue(token, consts.GE)
			} else {
				clause.Op = operatorFromValue(token, consts.GT)
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	clause.Right, err = parseExpr(node)
	if err != nil {
		return nil, err
	}

	return clause, nil
}
