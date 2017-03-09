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
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/errutil"
)

func Format(path string, data []byte) (formatted []byte, err error) {
	scanner, err := NewScanner(path, data)
	if err != nil {
		return nil, err
	}

	root, err := scanRoot(scanner)
	if err != nil {
		return nil, err
	}

	formatted, err = formatRoot(root)
	if err != nil {
		return nil, err
	}

	return formatted, nil
}

func formatRoot(node *listNode) (formatted []byte, err error) {
	groups, err := tupleGroups(node)
	if err != nil {
		return nil, err
	}
	return formatTupleGroups(-1, groups)
}

func tupleGroups(node *listNode) (groups [][]*tupleNode, err error) {
	var group []*tupleNode
	for len(node.value) > 0 {

		tuple, err := node.consumeTuple()
		if err != nil {
			return nil, err
		}

		if len(group) == 0 {
			group = append(group, tuple)
			continue
		}

		last_line := group[len(group)-1].pos.Line
		if tuple.pos.Line-last_line < 2 {
			group = append(group, tuple)
			continue
		}

		// start a new group
		groups = append(groups, group)
		group = []*tupleNode{tuple}
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	return groups, nil
}

func formatTupleGroups(indent int, groups [][]*tupleNode) (
	formatted []byte, err error) {

	for _, group := range groups {
		formatted_group, err := formatTupleGroup(indent+1, group)
		if err != nil {
			return nil, err
		}
		formatted = append(formatted, formatted_group...)
	}
	return formatted, nil
}

func formatTupleGroup(indent int, group []*tupleNode) (
	formatted []byte, err error) {

	var wordss [][]string
	var lists []*listNode
	for _, tuple := range group {
		words, list, err := stringifyTuple(tuple)
		if err != nil {
			return nil, err
		}
		wordss = append(wordss, words)
		lists = append(lists, list)
	}

	alignWords(wordss)

	var line []byte
	addLine := func() {
		line = bytes.TrimRight(line, " ")
		formatted = append(formatted, strings.Repeat("\t", indent)...)
		formatted = append(formatted, line...)
		formatted = append(formatted, '\n')
		line = line[:0]
	}

	for i, words := range wordss {
		list := lists[i]

		line = append(line, strings.Join(words, " ")...)

		if list == nil {
			addLine()
			continue
		}

		if !isListMultiLine(list) {
			formatted_line, err := formatSingleLineList(list)
			if err != nil {
				return nil, err
			}
			line = append(line, ' ')
			line = append(line, formatted_line...)
			addLine()
			continue
		}

		line = append(line, " ("...)
		addLine()

		groups, err := tupleGroups(list)
		if err != nil {
			return nil, err
		}
		formatted_groups, err := formatTupleGroups(indent, groups)
		if err != nil {
			return nil, err
		}
		formatted = append(formatted, formatted_groups...)
		formatted = formatted[:len(formatted)-1] // strip off a \n
		formatted = append(formatted, strings.Repeat("\t", indent)...)
		formatted = append(formatted, ")\n"...)
	}

	return append(formatted, '\n'), nil
}

func stringifyTuple(tuple *tupleNode) (words []string, list *listNode,
	err error) {

	// oh man there has to be a better way :)
	var word string
	var operator bool
	var dot bool

	for i, token := range tuple.value {
		switch node := token.(type) {
		case *tokenNode:
			switch node.tok {
			case Ident:
				if dot {
					word += node.text
					dot = false
					operator = false
					continue
				}

				operator = false
				if word != "" {
					words = append(words, word)
				}
				word = node.text

			case Dot:
				word += "."
				dot = true

			case Question:
				dot = false
				operator = false
				if word != "" {
					words = append(words, word)
				}
				word = node.text

			case Exclamation, Equal, LeftAngle, RightAngle:
				dot = false

				if operator {
					word += node.text
					continue
				}

				if word != "" {
					words = append(words, word)
				}

				word = node.text
				operator = true
			}

		case *listNode:
			if i != len(tuple.value)-1 {
				invalid := tuple.value[i+1]
				return nil, nil, errutil.New(invalid.getPos(),
					"expected end of tuple. got a %s: %s",
					invalid.nodeType(), invalid)
			}

			if word != "" {
				words = append(words, word)
			}
			return words, node, nil
		}
	}

	if word != "" {
		words = append(words, word)
	}
	return words, nil, nil
}

func alignWords(wordss [][]string) {
	max_len := 0
	for _, words := range wordss {
		if len(words) > max_len {
			max_len = len(words)
		}
	}

	for i := 0; i < max_len; i++ {
		max_word_len := 0
		for _, words := range wordss {
			if i < len(words) && len(words[i]) > max_word_len {
				max_word_len = len(words[i])
			}
		}
		for j, words := range wordss {
			if i >= len(words) {
				words = append(words, strings.Repeat(" ", max_word_len))
				wordss[j] = words
				continue
			}
			words[i] += strings.Repeat(" ", max_word_len-len(words[i]))
		}
	}
}

func isListMultiLine(list *listNode) bool {
	return list.pos.Line != list.end_pos.Line
}

func formatSingleLineList(list *listNode) (formatted []byte, err error) {
	formatted = append(formatted, "( "...)
	first := true
	for {
		tuple, err := list.consumeTupleOrEmpty()
		if err != nil {
			return nil, err
		}
		if tuple == nil {
			if formatted[len(formatted)-1] != ' ' {
				formatted = append(formatted, ' ')
			}
			formatted = append(formatted, ')')
			return formatted, nil
		}
		words, tuple_list, err := stringifyTuple(tuple)
		if err != nil {
			return nil, err
		}
		var list_part []byte
		if tuple_list != nil {
			list_part, err = formatSingleLineList(tuple_list)
			if err != nil {
				return nil, err
			}
		}
		if !first {
			formatted = append(formatted, ", "...)
		}
		formatted = append(formatted, strings.Join(words, " ")...)
		if len(list_part) > 0 {
			formatted = append(formatted, ' ')
			formatted = append(formatted, list_part...)
		}
		first = false
	}
}
