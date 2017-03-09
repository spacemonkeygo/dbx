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
	"sort"
	"strings"
	"text/scanner"

	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
)

type node interface {
	nodeType() string
	getPos() scanner.Position
}

type listNode struct {
	pos     scanner.Position
	end_pos scanner.Position
	value   []node
}

type tupleNode struct {
	pos   scanner.Position
	value []node
}

type tokenNode struct {
	pos  scanner.Position
	tok  Token
	text string
}

func (listNode) nodeType() string  { return "list" }
func (tupleNode) nodeType() string { return "tuple" }
func (tokenNode) nodeType() string { return "token" }

func (l listNode) getPos() scanner.Position  { return l.pos }
func (t tupleNode) getPos() scanner.Position { return t.pos }
func (t tokenNode) getPos() scanner.Position { return t.pos }

func (l listNode) String() string {
	return fmt.Sprintf("(%s)", stringNodes(l.value, ", "))
}

func (t tupleNode) String() string {
	return fmt.Sprintf("%s", stringNodes(t.value, " "))
}

func (t tokenNode) String() string {
	return fmt.Sprintf("%q", t.text)
}

func stringNodes(values []node, join string) string {
	var strs []string
	for _, val := range values {
		strs = append(strs, fmt.Sprint(val))
	}
	return strings.Join(strs, join)
}

func isList(n node) bool {
	_, ok := n.(*listNode)
	return ok
}

func expectList(n node) (*listNode, error) {
	list, ok := n.(*listNode)
	if !ok {
		return nil, errutil.New(n.getPos(), "expected a list. got a %s: %s",
			n.nodeType(), n)
	}
	return list, nil
}

func expectTuple(n node) (*tupleNode, error) {
	tuple, ok := n.(*tupleNode)
	if !ok {
		return nil, errutil.New(n.getPos(), "expected a tuple. got a %s: %s",
			n.nodeType(), n)
	}
	return tuple, nil
}

func expectToken(n node) (*tokenNode, error) {
	token, ok := n.(*tokenNode)
	if !ok {
		return nil, errutil.New(n.getPos(), "expected a token. got a %s: %s",
			n.nodeType(), n)
	}
	return token, nil
}

func scanRoot(scanner *Scanner) (*listNode, error) {
	l := &listNode{}

	for {
		switch tok := scanner.Peek(); tok {
		case EOF:
			scanner.Scan()
			return l, nil
		case Ident:
			tuple, err := newTupleNode(scanner)
			if err != nil {
				return nil, err
			}
			l.value = append(l.value, tuple)
		default:
			return nil, expectedToken(scanner.Pos(), tok, EOF, Ident)
		}
	}
}

func newListNode(scanner *Scanner) (*listNode, error) {
	l := &listNode{
		pos: scanner.Pos(),
	}

	_, _, err := scanner.ScanExact(OpenParen)
	if err != nil {
		return nil, err
	}

	for {
		switch tok := scanner.Peek(); tok {
		case CloseParen:
			_, l.end_pos, _ = scanner.Scan()
			return l, nil
		default:
			tuple, err := newTupleNode(scanner)
			if err != nil {
				return nil, err
			}
			l.value = append(l.value, tuple)
		}
	}
}

func newTupleNode(scanner *Scanner) (*tupleNode, error) {
	t := &tupleNode{
		pos: scanner.Pos(),
	}

	for {
		switch tok := scanner.Peek(); tok {
		case Comma:
			scanner.Scan()
			return t, nil
		// we don't require trailing commas for tuple nodes in a list
		case CloseParen:
			return t, nil
		case OpenParen:
			list, err := newListNode(scanner)
			if err != nil {
				return nil, err
			}
			t.value = append(t.value, list)
		default:
			token, err := newTokenNode(scanner)
			if err != nil {
				return nil, err
			}
			t.value = append(t.value, token)
		}
	}
}

func newTokenNode(scanner *Scanner) (*tokenNode, error) {
	t := &tokenNode{}

	switch tok := scanner.Peek(); tok {
	case EOF:
		return nil, errutil.New(scanner.Pos(), "unexpected EOF")
	case Illegal:
		_, pos, text := scanner.Scan()
		return nil, errutil.New(pos, "illegal token: %q", text)
	default:
		_, pos, text := scanner.Scan()
		t.tok = tok
		t.text = text
		t.pos = pos
		return t, nil
	}
}

//
// list nodes
//

func (l *listNode) consume() (n node, err error) {
	if len(l.value) == 0 {
		return nil, errutil.New(l.getPos(), "expected a node. found nothing")
	}
	n, l.value = l.value[0], l.value[1:]
	// fmt.Printf("consumed list entry: %s\n", n)
	return n, nil
}

func (l *listNode) consumeTuple() (*tupleNode, error) {
	if len(l.value) == 0 {
		return nil, errutil.New(l.getPos(), "expected a tuple. found nothing")
	}
	node, err := l.consume()
	if err != nil {
		return nil, err
	}
	return expectTuple(node)
}

func (l *listNode) consumeTupleOrEmpty() (*tupleNode, error) {
	if len(l.value) == 0 {
		return nil, nil
	}
	node, err := l.consume()
	if err != nil {
		return nil, err
	}
	return expectTuple(node)
}

type tupleCases map[string]func(*tupleNode) error

func (t tupleCases) idents() []string {
	out := make([]string, 0, len(t))
	for key := range t {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (l *listNode) consumeAnyTuples(cases tupleCases) error {
	for len(l.value) > 0 {
		if err := l.consumeAnyTuple(cases); err != nil {
			return err
		}
	}
	return nil
}

func (l *listNode) consumeAnyTuple(cases tupleCases) error {
	tuple, err := l.consumeTuple()
	if err != nil {
		return err
	}
	token, err := tuple.consumeToken(Ident)
	if err != nil {
		return err
	}
	for text, cont := range cases {
		if token.text == text {
			err := cont(tuple)
			if err != nil {
				return err
			}
			return tuple.assertEmpty()
		}
	}
	return expectedKeyword(token.pos, token.text, cases.idents()...)
}

//
// tuple nodes
//

func (t *tupleNode) consume() (n node, err error) {
	if len(t.value) == 0 {
		return nil, errutil.New(t.getPos(), "expected a node. found nothing")
	}
	n, t.value = t.value[0], t.value[1:]
	// fmt.Printf("consumed tuple entry: %s\n", n)
	return n, nil
}

func (t *tupleNode) consumeToken(kinds ...Token) (*tokenNode, error) {
	if len(t.value) == 0 {
		return nil, errutil.New(t.getPos(), "expected a token. found nothing")
	}
	node, err := t.consume()
	if err != nil {
		return nil, err
	}
	token, err := expectToken(node)
	if err != nil {
		return nil, err
	}
	if err := token.assertKind(kinds...); err != nil {
		return nil, err
	}
	return token, nil
}

func (t *tupleNode) consumeTokens(kinds ...Token) ([]*tokenNode, error) {
	var out []*tokenNode
	for len(t.value) > 0 {
		token, err := t.consumeToken(kinds...)
		if err != nil {
			return nil, err
		}
		out = append(out, token)
	}
	return out, nil
}

func (t *tupleNode) consumeIfToken(kinds ...Token) *tokenNode {
	if len(t.value) == 0 {
		return nil
	}
	token, err := expectToken(t.value[0])
	if err != nil {
		return nil
	}
	if err := token.assertKind(kinds...); err != nil {
		return nil
	}
	t.consume()
	return token
}

func (t *tupleNode) consumeTokenOrEmpty(kinds ...Token) (*tokenNode, error) {
	if len(t.value) == 0 {
		return nil, nil
	}
	return t.consumeToken(kinds...)
}

func (t *tupleNode) consumeList() (*listNode, error) {
	if len(t.value) == 0 {
		return nil, errutil.New(t.getPos(), "expected a list. found nothing")
	}
	node, err := t.consume()
	if err != nil {
		return nil, err
	}
	return expectList(node)
}

func (t *tupleNode) consumeIfList() *listNode {
	if len(t.value) == 0 {
		return nil
	}
	list, err := expectList(t.value[0])
	if err != nil {
		return nil
	}
	t.consume()
	return list
}

func (t *tupleNode) consumeListOrEmpty() (*listNode, error) {
	if len(t.value) == 0 {
		return nil, nil
	}
	return t.consumeList()
}

type tokenCase struct {
	tok  Token
	text string
}

func (t Token) tokenCase() tokenCase {
	return tokenCase{
		tok:  t,
		text: string(t),
	}
}

type tokenCases map[tokenCase]func(*tokenNode) error

func (t tokenCases) idents() []string {
	out := make([]string, 0, len(t))
	for c := range t {
		out = append(out, c.text)
	}
	sort.Strings(out)
	return out
}

func (t *tupleNode) consumeTokensNamed(cases tokenCases) error {
	for len(t.value) > 0 {
		if err := t.consumeTokenNamed(cases); err != nil {
			return err
		}
	}
	return nil
}

func (t *tupleNode) consumeTokenNamed(cases tokenCases) error {
	token, err := t.consumeToken()
	if err != nil {
		return err
	}
	for c, cont := range cases {
		if token.tok == c.tok && token.text == c.text {
			return cont(token)
		}
	}
	return expectedKeyword(token.pos, token.text, cases.idents()...)
}

func (t *tupleNode) consumeTokensNamedUntilList(cases tokenCases) error {
	if len(t.value) == 0 {
		return errutil.New(t.getPos(), "expected a token. found nothing")
	}
	for {
		if err := t.consumeTokenNamed(cases); err != nil {
			return err
		}
		if len(t.value) == 0 {
			return errutil.New(t.getPos(), "expected a list. found nothing")
		}
		if isList(t.value[0]) {
			break
		}
	}
	return nil
}

func (t *tupleNode) consumeDottedIdents() (
	first, second *tokenNode, err error) {

	first, err = t.consumeToken(Ident)
	if err != nil {
		return nil, nil, err
	}

	if t.consumeIfToken(Dot) == nil {
		return first, nil, nil
	}

	second, err = t.consumeToken(Ident)
	if err != nil {
		return nil, nil, err
	}

	return first, second, nil
}

func (t *tupleNode) consumeDottedIdentsOrEmpty() (
	first, second *tokenNode, err error) {

	if len(t.value) == 0 {
		return nil, nil, nil
	}
	return t.consumeDottedIdents()
}

func (t *tupleNode) assertEmpty() error {
	if len(t.value) > 0 {
		return errutil.New(t.value[0].getPos(),
			"expected end of tuple. got a %s: %s",
			t.value[0].nodeType(), t.value[0])
	}
	return nil
}

//
// token nodes
//

func (t *tokenNode) assertKind(kinds ...Token) error {
	if len(kinds) == 0 {
		return nil
	}
	for _, kind := range kinds {
		if t.tok == kind {
			return nil
		}
	}
	return expectedToken(t.pos, t.tok, kinds...)
}

func (t *tokenNode) debugAssertToken(kind Token) {
	if t.tok != kind {
		panic(fmt.Sprintf("internal error: token is a %s, not %s",
			t.tok, kind))
	}
}
