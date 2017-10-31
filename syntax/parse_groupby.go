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

func parseGroupBy(node *tupleNode) (*ast.GroupBy, error) {
	group_by := new(ast.GroupBy)
	group_by.Pos = node.getPos()

	field_refs, err := parseFieldRefs(node, true)
	if err != nil {
		return nil, err
	}
	group_by.Fields = field_refs

	return group_by, nil
}
