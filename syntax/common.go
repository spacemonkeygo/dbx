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
	"sort"

	"github.com/spacemonkeygo/errors"
	"gopkg.in/spacemonkeygo/dbx.v1/ast"
	"gopkg.in/spacemonkeygo/dbx.v1/consts"
)

var (
	Error = errors.NewClass("syntax")
)

func debugConsume(node *tupleNode) error {
	for {
		_, err := node.consume()
		if err != nil {
			return nil
		}
	}
}

func tupleFlagField(kind, field string, val **ast.Bool) func(*tupleNode) error {
	return func(node *tupleNode) error {
		if *val != nil {
			return previouslyDefined(node.getPos(), kind, field, (*val).Pos)
		}

		*val = boolFromValue(node, true)
		return nil
	}
}

func tokenFlagField(kind, field string, val **ast.Bool) func(*tokenNode) error {
	return func(node *tokenNode) error {
		if *val != nil {
			return previouslyDefined(node.getPos(), kind, field, (*val).Pos)
		}

		*val = boolFromValue(node, true)
		return nil
	}
}

var fieldTypeMapping = map[string]consts.FieldType{
	"serial":     consts.SerialField,
	"serial64":   consts.Serial64Field,
	"int":        consts.IntField,
	"int64":      consts.Int64Field,
	"uint":       consts.UintField,
	"uint64":     consts.Uint64Field,
	"bool":       consts.BoolField,
	"text":       consts.TextField,
	"timestamp":  consts.TimestampField,
	"utimestamp": consts.TimestampUTCField,
	"float":      consts.FloatField,
	"float64":    consts.Float64Field,
	"blob":       consts.BlobField,
	"date":       consts.DateField,
}

func validFieldTypes() []string {
	out := make([]string, 0, len(fieldTypeMapping))
	for typ := range fieldTypeMapping {
		out = append(out, typ)
	}
	sort.Strings(out)
	return out
}

func modelRefFromToken(node *tokenNode) *ast.ModelRef {
	node.debugAssertToken(Ident)
	return &ast.ModelRef{
		Pos:   node.getPos(),
		Model: stringFromToken(node),
	}
}

func stringFromToken(node *tokenNode) *ast.String {
	node.debugAssertToken(Ident)
	return stringFromValue(node, node.text)
}

func stringsFromTokens(nodes []*tokenNode) (out []*ast.String) {
	for _, node := range nodes {
		out = append(out, stringFromToken(node))
	}
	return out
}

func fieldRefFromTokens(first, second *tokenNode) *ast.FieldRef {
	ref := &ast.FieldRef{
		Pos:   first.getPos(),
		Model: stringFromToken(first),
	}

	if second != nil {
		ref.Field = stringFromToken(second)
	}

	return ref
}

func relativeFieldRefFromToken(node *tokenNode) *ast.RelativeFieldRef {
	return &ast.RelativeFieldRef{
		Pos:   node.getPos(),
		Field: stringFromToken(node),
	}
}

func stringFromValue(n node, val string) *ast.String {
	return &ast.String{
		Pos:   n.getPos(),
		Value: val,
	}
}

func boolFromValue(n node, val bool) *ast.Bool {
	return &ast.Bool{
		Pos:   n.getPos(),
		Value: val,
	}
}

func intFromValue(n node, val int) *ast.Int {
	return &ast.Int{
		Pos:   n.getPos(),
		Value: val,
	}
}

func fieldTypeFromValue(n node, val consts.FieldType) *ast.FieldType {
	return &ast.FieldType{
		Pos:   n.getPos(),
		Value: val,
	}
}

func relationKindFromValue(n node, val consts.RelationKind) *ast.RelationKind {
	return &ast.RelationKind{
		Pos:   n.getPos(),
		Value: val,
	}
}

func operatorFromValue(n node, val consts.Operator) *ast.Operator {
	return &ast.Operator{
		Pos:   n.getPos(),
		Value: val,
	}
}

func nullFromToken(token *tokenNode) *ast.Null {
	return &ast.Null{
		Pos: token.getPos(),
	}
}

func placeholderFromToken(token *tokenNode) *ast.Placeholder {
	return &ast.Placeholder{
		Pos: token.getPos(),
	}
}

func funcCallFromTokenAndArgs(name *tokenNode, args []*ast.Expr) *ast.FuncCall {
	return &ast.FuncCall{
		Pos:  name.getPos(),
		Name: stringFromToken(name),
		Args: args,
	}
}
