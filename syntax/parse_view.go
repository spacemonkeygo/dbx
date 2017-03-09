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

func parseView(node *tupleNode) (*ast.View, error) {
	view := new(ast.View)
	view.Pos = node.getPos()

	err := node.consumeTokensNamedUntilList(tokenCases{
		{Ident, "all"}:   tokenFlagField("view", "all", &view.All),
		{Ident, "paged"}: tokenFlagField("view", "paged", &view.Paged),
		{Ident, "count"}: tokenFlagField("view", "count", &view.Count),
		{Ident, "has"}:   tokenFlagField("view", "has", &view.Has),
		{Ident, "limitoffset"}: tokenFlagField("view", "limitoffset",
			&view.LimitOffset),
		{Ident, "scalar"}: tokenFlagField("view", "scalar", &view.Scalar),
		{Ident, "one"}:    tokenFlagField("view", "one", &view.One),
		{Ident, "first"}:  tokenFlagField("view", "first", &view.First),
	})
	if err != nil {
		return nil, err
	}

	return view, nil
}
