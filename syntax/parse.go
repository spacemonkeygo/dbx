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

func Parse(path string, data []byte) (root *ast.Root, err error) {
	scanner, err := NewScanner(path, data)
	if err != nil {
		return nil, err
	}

	return parseRoot(scanner)
}

func parseRoot(scanner *Scanner) (*ast.Root, error) {
	list, err := scanRoot(scanner)
	if err != nil {
		return nil, err
	}

	root := new(ast.Root)

	err = list.consumeAnyTuples(tupleCases{
		"model": func(node *tupleNode) error {
			model, err := parseModel(node)
			if err != nil {
				return err
			}
			root.Models = append(root.Models, model)

			return nil
		},
		"create": func(node *tupleNode) error {
			cre, err := parseCreate(node)
			if err != nil {
				return err
			}
			root.Creates = append(root.Creates, cre)

			return nil
		},
		"read": func(node *tupleNode) error {
			read, err := parseRead(node)
			if err != nil {
				return err
			}
			root.Reads = append(root.Reads, read)

			return nil
		},
		"update": func(node *tupleNode) error {
			upd, err := parseUpdate(node)
			if err != nil {
				return err
			}
			root.Updates = append(root.Updates, upd)

			return nil
		},
		"delete": func(node *tupleNode) error {
			del, err := parseDelete(node)
			if err != nil {
				return err
			}
			root.Deletes = append(root.Deletes, del)

			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return root, nil
}
