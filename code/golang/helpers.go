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

package golang

import (
	"fmt"
	"strings"

	"bitbucket.org/pkg/inflect"
	"gopkg.in/spacemonkeygo/dbx.v1/ir"
)

func cleanSignature(in string) (out string) {
	return reCollapseSpace.ReplaceAllString(strings.TrimSpace(in), " ")
}

func sliceofFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).SliceOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func paramFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Param)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ",\n"), nil
}

func argFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Arg)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func addrofFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).AddrOf)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func initFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Init)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func initnewFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).InitNew)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, "\n"), nil
}

func zeroFn(intf interface{}) (string, error) {
	vs, err := forVars(intf, (*Var).Zero)
	if err != nil {
		return "", err
	}
	return strings.Join(vs, ", "), nil
}

func flattenFn(intf interface{}) (flattened []*Var, err error) {
	switch obj := intf.(type) {
	case *Var:
		flattened = obj.Flatten()
	case []*Var:
		for _, v := range obj {
			flattened = append(flattened, v.Flatten()...)
		}
	default:
		return nil, Error.New("unsupported type %T", obj)
	}
	return flattened, nil
}

func fieldvalueFn(vars []*Var) string {
	var values []string
	for _, v := range vars {
		values = append(values, fmt.Sprintf("%s.value()", v.Name))
	}
	return strings.Join(values, ", ")
}

func ctxparamFn(intf interface{}) (string, error) {
	param, err := paramFn(intf)
	if err != nil {
		return "", err
	}
	if param == "" {
		return "ctx context.Context", nil
	}
	return "ctx context.Context,\n" + param, nil
}

func ctxargFn(intf interface{}) (string, error) {
	arg, err := argFn(intf)
	if err != nil {
		return "", err
	}
	if arg == "" {
		return "ctx", nil
	}
	return "ctx, " + arg, nil
}

func commaFn(in string) string {
	if in == "" {
		return ""
	}
	return in + ", "
}

func forVars(intf interface{}, fn func(v *Var) string) ([]string, error) {
	var elems []string
	switch obj := intf.(type) {
	case *Var:
		elems = append(elems, fn(obj))
	case []*Var:
		for _, v := range obj {
			elems = append(elems, fn(v))
		}
	default:
		return nil, Error.New("unsupported type %T", obj)
	}
	return elems, nil
}

func structName(m *ir.Model) string {
	return inflect.Camelize(m.Name)
}

func fieldName(f *ir.Field) string {
	return inflect.Camelize(f.Name)
}

func convertSuffix(suffix []string) string {
	parts := make([]string, 0, len(suffix))
	for _, part := range suffix {
		parts = append(parts, inflect.Camelize(part))
	}
	return strings.Join(parts, "_")
}
