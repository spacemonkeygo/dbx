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
	"bytes"
	"path/filepath"
	"text/scanner"
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

func (t Token) String() string {
	switch t {
	case Ident:
		return "Ident"
	case Int:
		return "Int"
	case EOF:
		return "EOF"
	case Colon:
		return "Colon"
	case Dot:
		return "Dot"
	case Comma:
		return "Comma"
	case Equal:
		return "Equal"
	case LeftAngle:
		return "LeftAngle"
	case RightAngle:
		return "RightAngle"
	case Question:
		return "Question"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case Exclamation:
		return "Exclamation"
	case Illegal:
		return "Illegal"
	default:
		return "Unknown"
	}
}

type token struct {
	tok  Token
	pos  scanner.Position
	text string
}

type Scanner struct {
	tokens []token
	pos    int
}

func NewScanner(filename string, data []byte) (*Scanner, error) {
	var s scanner.Scanner
	s.Init(bytes.NewReader(data))
	s.Mode = scanner.ScanInts | scanner.ScanIdents | scanner.ScanComments |
		scanner.SkipComments
	s.Whitespace = 0

	base_filename := filepath.Base(filename)

	var tokens []token

	var tok rune
	for tok != scanner.EOF {
		pos := s.Pos()
		pos.Filename = base_filename
		tok = s.Scan()

		if tok == ' ' || tok == '\r' || tok == '\t' {
			continue
		}

		// insert a comma at newlines and eof unless we already have a comma
		// or we have a list opening
		if tok == '\n' || tok == scanner.EOF {
			if len(tokens) > 0 {
				switch tokens[len(tokens)-1].tok {
				case OpenParen, Comma:
				default:
					tokens = append(tokens, token{
						tok:  Comma,
						pos:  pos,
						text: ",",
					})
				}
			}
			if tok == '\n' {
				continue
			}
		}

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
	// defer func() {
	// 	fmt.Printf("SCAN: %20s %-10s %q\n", pos, token, text)
	// }()

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
