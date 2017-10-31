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
	"fmt"
	"os"
	"reflect"
	"text/scanner"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

var positionType = reflect.TypeOf(scanner.Position{})

func redactPos(val interface{}) {
	// it's a ptr to struct
	rv := reflect.ValueOf(val)
	if rv.IsNil() {
		return
	}
	rv = rv.Elem()
	for i := 0; i < rv.NumField(); i++ {
		fi := rv.Field(i)
		if fi.Type() == positionType {
			fi.Set(reflect.Zero(positionType))
			continue
		}
		switch ft := fi.Type(); {
		// field is a *struct
		case ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Struct:
			redactPos(fi.Interface())

		// field is a []*struct
		case ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Ptr &&
			ft.Elem().Elem().Kind() == reflect.Struct:
			for i := 0; i < fi.Len(); i++ {
				redactPos(fi.Index(i).Interface())
			}
		}
	}
}

func assertWhere(tw *testutil.T, data string, exp *ast.Where) {
	scanner, err := NewScanner("<memory>", []byte(data))
	tw.AssertNoError(err)
	node, err := newTupleNode(scanner)
	tw.AssertNoError(err)

	got, err := parseWhere(node)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.GetMessage(err))
		fmt.Fprintln(os.Stderr, errutil.GetContext([]byte(data), err))
		fmt.Fprintln(os.Stderr, errors.GetStack(err))
		tw.FailNow()
	}
	redactPos(got)

	if !reflect.DeepEqual(got, exp) {
		tw.Errorf("got: %s", got)
		tw.Errorf("exp: %s", exp)
		tw.FailNow()
	}
}
