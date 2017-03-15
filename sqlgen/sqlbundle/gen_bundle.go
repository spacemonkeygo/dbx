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

// +build ignore

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const template = `%s
//go:generate go run gen_bundle.go

package sqlbundle

const Source = %q
`

var afterImportRe = regexp.MustCompile(`(?m)^\)$`)

func main() {
	copyright, bundle := loadCopyright(), loadBundle()
	output := []byte(fmt.Sprintf(template, copyright, bundle))

	err := ioutil.WriteFile("bundle.go", output, 0644)
	if err != nil {
		panic(err)
	}
}

func loadCopyright() string {
	fh, err := os.Open("gen_bundle.go")
	if err != nil {
		panic(err)
	}
	defer fh.Close()

	var buf bytes.Buffer
	scanner := bufio.NewScanner(fh)

	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, "//") {
			return buf.String()
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	panic("unreachable")
}

func loadBundle() string {
	source, err := exec.Command("bundle",
		"-dst", "gopkg.in/spacemonkeygo/dbx.v1/sqlgen/sqlbundle",
		"gopkg.in/spacemonkeygo/dbx.v1/sqlgen").Output()
	if err != nil {
		panic(err)
	}

	index := afterImportRe.FindIndex(source)
	if index == nil {
		panic("unable to find package clause")
	}

	return string(bytes.TrimSpace(source[index[1]:]))
}
