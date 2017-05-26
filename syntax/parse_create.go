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

func parseCreate(node *tupleNode) (*ast.Create, error) {
	cre := new(ast.Create)
	cre.Pos = node.getPos()

	model_ref_token, err := node.consumeToken(Ident)
	if err != nil {
		return nil, err
	}
	cre.Model = modelRefFromToken(model_ref_token)

	list_token, err := node.consumeList()
	if err != nil {
		return nil, err
	}

	err = list_token.consumeAnyTuples(tupleCases{
		"raw":      tupleFlagField("create", "raw", &cre.Raw),
		"noreturn": tupleFlagField("create", "noreturn", &cre.NoReturn),
		"suffix": func(node *tupleNode) error {
			if cre.Suffix != nil {
				return previouslyDefined(node.getPos(), "create", "suffix",
					cre.Suffix.Pos)
			}

			suffix, err := parseSuffix(node)
			if err != nil {
				return err
			}
			cre.Suffix = suffix

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return cre, nil
}
