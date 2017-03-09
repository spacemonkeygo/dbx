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

func parseOrderBy(node *tupleNode) (*ast.OrderBy, error) {
	order_by := new(ast.OrderBy)
	order_by.Pos = node.getPos()

	err := node.consumeTokenNamed(tokenCases{
		{Ident, "asc"}: func(token *tokenNode) error { return nil },
		{Ident, "desc"}: func(token *tokenNode) error {
			order_by.Descending = boolFromValue(token, true)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	field_refs, err := parseFieldRefs(node, true)
	if err != nil {
		return nil, err
	}
	order_by.Fields = field_refs

	return order_by, nil
}
