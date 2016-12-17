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
	"bytes"
	"path/filepath"
	"text/scanner"

	"github.com/spacemonkeygo/errors"
)

var (
	Error = errors.NewClass("parser")
)

type Token string

const (
	Ident       Token = "Ident"
	Int         Token = "Int"
	EOF         Token = "EOF"
	Colon       Token = ":"
	Dot         Token = "."
	Comma       Token = ","
	Equal       Token = "="
	LeftAngle   Token = "<"
	RightAngle  Token = ">"
	Question    Token = "?"
	OpenParen   Token = "("
	CloseParen  Token = ")"
	Exclamation Token = "!"
	Illegal     Token = "Illegal"
)

type token struct {
	tok  Token
	pos  scanner.Position
	text string
}

type Scanner struct {
	tokens []token
	pos    int
}

const ws = 1<<'\t' | 1<<' '

func NewScanner(filename string, data []byte) (*Scanner, error) {
	var s scanner.Scanner
	s.Init(bytes.NewReader(data))
	s.Mode = scanner.ScanInts | scanner.ScanIdents | scanner.ScanComments |
		scanner.SkipComments

	var tokens []token

	var tok rune
	for tok != scanner.EOF {
		tok = s.Scan()
		pos := s.Pos()
		pos.Filename = filepath.Base(filename)
		tokens = append(tokens, token{
			tok:  convertToken(tok),
			pos:  pos,
			text: s.TokenText(),
		})
	}

	if s.ErrorCount > 0 {
		return nil, Error.New("%d errors encountered", s.ErrorCount)
	}
	return &Scanner{
		tokens: tokens,
	}, nil
}

func (s *Scanner) Pos() scanner.Position {
	return s.tokens[s.pos].pos
}

func (s *Scanner) Scan() (token Token, pos scanner.Position, text string) {
	return s.scan()
}

func (s *Scanner) ScanWhile(token Token) {
	for s.peek() == token {
		s.scan()
	}
}

func (s *Scanner) ScanTo(token Token) {
	for {
		switch s.peek() {
		case Illegal, EOF:
			return
		case token:
			s.scan()
			return
		}
		s.scan()
	}
}

func (s *Scanner) scan() (token Token, pos scanner.Position, text string) {
	//	defer func() {
	//		fmt.Printf("SCAN: %20s %-10s %q\n", pos, token, text)
	//	}()

	t := s.tokens[s.pos]
	if (s.pos + 1) < len(s.tokens) {
		s.pos++
	}
	return t.tok, t.pos, t.text
}

func (s *Scanner) Peek() (token Token) {
	return s.peek()
}

func (s *Scanner) peek() (token Token) {
	t := s.tokens[s.pos]
	return t.tok
}

func (s *Scanner) ScanIf(token Token) (pos scanner.Position, text string,
	ok bool) {
	if s.peek() == token {
		_, pos, text = s.scan()
		ok = true
	}
	return pos, text, ok
}

func (s *Scanner) ScanExact(token Token) (pos scanner.Position, text string,
	err error) {

	candidate, pos, text := s.scan()
	if candidate != token {
		return pos, "", expectedToken(pos, candidate, token)
	}
	return pos, text, nil
}

func (s *Scanner) ScanOneOf(tokens ...Token) (token Token,
	pos scanner.Position, text string, err error) {

	candidate, pos, text := s.scan()
	for _, token = range tokens {
		if candidate == token {
			return token, pos, text, nil
		}
	}
	return candidate, pos, text, expectedToken(pos, candidate, tokens...)
}

func convertToken(tok rune) Token {
	switch tok {
	case scanner.Ident:
		return Ident
	case scanner.Int:
		return Int
	case scanner.EOF:
		return EOF
	case '!':
		return Exclamation
	case ':':
		return Colon
	case '.':
		return Dot
	case ',':
		return Comma
	case '=':
		return Equal
	case '(':
		return OpenParen
	case ')':
		return CloseParen
	case '<':
		return LeftAngle
	case '>':
		return RightAngle
	case '?':
		return Question
	default:
		return Illegal
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
