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
	"reflect"
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

var positionType = reflect.TypeOf(scanner.Position{})

func redactPos(val interface{}) {
	rv := reflect.ValueOf(val).Elem() // ptr to struct
	for i := 0; i < rv.NumField(); i++ {
		fi := rv.Field(i)
		if fi.Type() == positionType {
			fi.Set(reflect.Zero(positionType))
			continue
		}
		if fi.Kind() == reflect.Ptr && fi.Elem().Kind() == reflect.Struct {
			redactPos(fi.Interface())
		}
	}
}

func assertWhere(tw *testutil.T, data string, exp *ast.Where) {
	scanner, err := NewScanner("", []byte(data))
	tw.AssertNoError(err)
	list, err := scanRoot(scanner)
	tw.AssertNoError(err)
	tuple, err := list.consumeTuple()
	tw.AssertNoError(err)

	got, err := parseWhere(tuple)
	tw.AssertNoError(err)
	redactPos(got)

	if !reflect.DeepEqual(got, exp) {
		tw.Errorf("got: %#v", got)
		tw.Errorf("exp: %#v", exp)
		tw.FailNow()
	}
}
