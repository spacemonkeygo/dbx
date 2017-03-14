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

package sqlgen

func Equal(a, b SQL) bool { return sqlEqual(a, b) }

func sqlEqual(a, b SQL) bool {
	switch a := a.(type) {
	case Literal:
		if b, ok := b.(Literal); ok {
			return a == b
		}
		return false
	case Literals:
		if b, ok := b.(Literals); ok {
			return a.Join == b.Join && sqlsEqual(a.SQLs, b.SQLs)
		}
		return false
	default:
		panic("unhandled sql type")
	}
}

func sqlsEqual(as, bs []SQL) bool {
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if !sqlEqual(as[i], bs[i]) {
			return false
		}
	}
	return true
}
