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

package golang

import (
	"strings"

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func cleanSignature(in string) (out string) {
	return reCollapseSpace.ReplaceAllString(strings.TrimSpace(in), " ")
}

func asParam(intf interface{}) (string, error) {
	return forVars(intf, (*Var).Param)
}

func asArg(intf interface{}) (string, error) {
	return forVars(intf, (*Var).Arg)
}

func asZero(intf interface{}) (string, error) {
	return forVars(intf, (*Var).Zero)
}

func forVars(intf interface{}, fn func(v *Var) string) (string, error) {
	var elems []string
	switch obj := intf.(type) {
	case Var:
		return fn(&obj), nil
	case *Var:
		return fn(obj), nil
	case []Var:
		for _, v := range obj {
			elems = append(elems, fn(&v))
		}
	case []*Var:
		for _, v := range obj {
			elems = append(elems, fn(v))
		}
	default:
		return "", Error.New("unsupported type %T", obj)
	}
	return strings.Join(elems, ", "), nil
}

func structName(m *ir.Model) string {
	return inflect.Camelize(m.Name)
}

func fieldName(f *ir.Field) string {
	return inflect.Camelize(f.Name)
}
