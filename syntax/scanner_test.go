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
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestScanDoubleList(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	scanner, err := NewScanner("", []byte("((foo))"))
	tw.AssertNoError(err)

	got, err := newTupleNode(scanner)
	tw.AssertNoError(err)

	exp := &tupleNode{
		value: []node{&listNode{
			value: []node{&tupleNode{
				value: []node{&listNode{
					value: []node{&tupleNode{
						value: []node{&tokenNode{
							tok:  Ident,
							text: "foo",
						}},
					}},
				}},
			}},
		}},
	}

	if !got.equal(exp) {
		tw.Errorf("got: %+v", got)
		tw.Errorf("exp: %+v", exp)
		tw.FailNow()
	}
}
