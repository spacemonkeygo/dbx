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

package syntax

import (
	"fmt"
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

func errorAt(n node, format string, args ...interface{}) error {
	return Error.New("%s: %s", n.getPos(), fmt.Sprintf(format, args...))
}

func previouslyDefined(n node, kind, field string,
	where scanner.Position) error {

	return errorAt(n, "%s already defined on %s. previous definition at %s",
		field, kind, where)
}

func flagField(kind, field string, val **ast.Bool) func(*tupleNode) error {
	return func(node *tupleNode) error {
		if *val != nil {
			return previouslyDefined(node, kind, field, (*val).Pos)
		}

		*val = boolFromValue(node, true)
		return nil
	}
}

func tokenFlagField(kind, field string, val **ast.Bool) func(*tokenNode) error {
	return func(node *tokenNode) error {
		if *val != nil {
			return previouslyDefined(node, kind, field, (*val).Pos)
		}

		*val = boolFromValue(node, true)
		return nil
	}
}
