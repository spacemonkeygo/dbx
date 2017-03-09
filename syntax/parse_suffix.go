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

func parseSuffix(node *tupleNode) (*ast.Suffix, error) {
	suf := new(ast.Suffix)
	suf.Pos = node.getPos()

	tokens, err := node.consumeTokens(Ident)
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		suf.Parts = append(suf.Parts, stringFromToken(token))
	}

	return suf, nil
}
