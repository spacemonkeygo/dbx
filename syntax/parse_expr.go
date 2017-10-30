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

func parseExpr(node *tupleNode) (expr *ast.Expr, err error) {
	// expressions are one of the following:
	//  placeholder      : ?
	//  string literal   : "foo"
	//  int literal      : 9
	//  float literal    : 9.9
	//  dotted field ref : model.field
	//  function         : foo(<expr>)

	expr = new(ast.Expr)
	expr.Pos = node.getPos()

	first, err := node.consumeToken(Question, String, Int, Float, Ident)
	if err != nil {
		return nil, err
	}

	switch first.tok {
	case Question:
		expr.Placeholder = placeholderFromToken(first)
		return expr, nil
	case String:
		unquoted, err := strconv.Unquote(first.text)
		if err != nil {
			return nil, errutil.New(first.getPos(),
				"(internal) unable to unquote string token text: %s", err)
		}
		expr.StringLit = stringFromValue(first, unquoted)
		return expr, nil
	case Int, Float:
		expr.NumberLit = stringFromValue(first, first.text)
		return expr, nil
	}

	first.debugAssertToken(Ident)

	if node.consumeIfToken(Dot) == nil {
		switch first.text {
		case "null":
			expr.Null = nullFromToken(first)
			return expr, nil
		case "true", "false":
			expr.BoolLit = boolFromToken(first)
			return expr, nil
		}

		list, err := node.consumeList()
		if err != nil {
			return nil, err
		}
		exprs, err := parseExprs(list)
		if err != nil {
			return nil, err
		}
		expr.FuncCall = funcCallFromTokenAndArgs(first, exprs)
		return expr, nil
	}

	second, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}

	expr.FieldRef = fieldRefFromTokens(first, second)
	return expr, nil
}

func parseExprs(list *listNode) (exprs []*ast.Expr, err error) {

	for {
		tuple, err := list.consumeTupleOrEmpty()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			return exprs, nil
		}
		expr, err := parseExpr(tuple)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}
}
