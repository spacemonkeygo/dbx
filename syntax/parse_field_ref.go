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
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
)

func parseFieldRefs(node *tupleNode, needs_dot bool) (*ast.FieldRefs, error) {
	refs := new(ast.FieldRefs)
	refs.Pos = node.getPos()

	for {
		ref, err := parseFieldRefOrEmpty(node, needs_dot)
		if err != nil {
			return nil, err
		}
		if ref == nil {
			return refs, nil
		}
		refs.Refs = append(refs.Refs, ref)
	}
}

func parseFieldRefOrEmpty(node *tupleNode, needs_dot bool) (
	*ast.FieldRef, error) {

	first, second, err := node.consumeDottedIdentsOrEmpty()
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, nil
	}
	if second == nil && needs_dot {
		return nil, errutil.New(first.getPos(),
			"field ref must specify a model")
	}
	return fieldRefFromTokens(first, second), nil
}

func parseFieldRef(node *tupleNode, needs_dot bool) (*ast.FieldRef, error) {
	first, second, err := node.consumeDottedIdents()
	if err != nil {
		return nil, err
	}
	if second == nil && needs_dot {
		return nil, errutil.New(first.getPos(),
			"field ref must specify a model")
	}
	return fieldRefFromTokens(first, second), nil
}

func parseRelativeFieldRefs(node *tupleNode) (*ast.RelativeFieldRefs, error) {
	refs := new(ast.RelativeFieldRefs)
	refs.Pos = node.getPos()

	for {
		ref_token, err := node.consumeTokenOrEmpty(Ident)
		if err != nil {
			return nil, err
		}
		if ref_token == nil {
			return refs, nil
		}
		refs.Refs = append(refs.Refs, relativeFieldRefFromToken(ref_token))
	}
}
