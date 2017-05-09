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

package xform

import (
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func transformJoins(lookup *lookup, ast_joins []*ast.Join) (
	models map[string]scanner.Position, joins []*ir.Join, err error) {

	models = make(map[string]scanner.Position)

	for _, ast_join := range ast_joins {
		left, err := lookup.FindField(ast_join.Left)
		if err != nil {
			return nil, nil, err
		}
		right, err := lookup.FindField(ast_join.Right)
		if err != nil {
			return nil, nil, err
		}

		joins = append(joins, &ir.Join{
			Type:  ast_join.Type.Get(),
			Left:  left,
			Right: right,
		})

		models[ast_join.Left.Model.Value] = ast_join.Left.Pos
		models[ast_join.Right.Model.Value] = ast_join.Right.Pos
	}

	return models, joins, nil
}
