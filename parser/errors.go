// Copyright (C) 2016 Space Monkey, Inc.
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

package parser

import (
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/ast"
)

func expectedKeyword(pos scanner.Position, actual string, expected ...string) (
	err error) {

	if len(expected) == 1 {
		return Error.New("%s: expected %q, got %q",
			pos, expected[0], actual)
	} else {
		return Error.New("%s: expected one of %q, got %q",
			pos, expected, actual)
	}
}

func unallowedAttribute(pos scanner.Position, field_type ast.FieldType,
	attr string) (err error) {

	return Error.New("%s: attribute %q not allowed for field type %q",
		pos, attr, field_type)
}

func expectedToken(pos scanner.Position, actual Token, expected ...Token) (
	err error) {

	if len(expected) == 1 {
		return Error.New("%s: expected %q; got %q",
			pos, expected[0], actual)
	} else {
		return Error.New("%s: expected one of %v; got %q",
			pos, expected, actual)
	}
}
