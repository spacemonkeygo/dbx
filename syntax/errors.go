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
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
)

func expectedKeyword(pos scanner.Position, actual string, expected ...string) (
	err error) {

	if len(expected) == 1 {
		return errutil.New(pos, "expected %q, got %q", expected[0], actual)
	} else {
		return errutil.New(pos, "expected one of %q, got %q", expected, actual)
	}
}

func expectedToken(pos scanner.Position, actual Token, expected ...Token) (
	err error) {

	if len(expected) == 1 {
		return errutil.New(pos, "expected %q; got %q", expected[0], actual)
	} else {
		return errutil.New(pos, "expected one of %v; got %q", expected, actual)
	}
}

func previouslyDefined(pos scanner.Position, kind, field string,
	where scanner.Position) error {

	return errutil.New(pos,
		"%s already defined on %s. previous definition at %s",
		field, kind, where)
}
