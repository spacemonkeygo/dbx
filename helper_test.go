// Copyright (C) 2017 Space Monkey, Inc.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func linedSource(source []byte) string {
	// scan once to find out how many lines
	scanner := bufio.NewScanner(bytes.NewReader(source))
	var lines int
	for scanner.Scan() {
		lines++
	}
	align := 1
	for ; lines > 0; lines = lines / 10 {
		align++
	}

	// now dump with aligned line numbers
	buf := bytes.NewBuffer(make([]byte, 0, len(source)*2))
	format := fmt.Sprintf("%%%dd: %%s\n", align)

	scanner = bufio.NewScanner(bytes.NewReader(source))
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Text()
		fmt.Fprintf(buf, format, i, line)
	}

	return buf.String()
}

type directives struct {
	ds map[string][]string
}

func (d *directives) add(name, value string) {
	if d.ds == nil {
		d.ds = make(map[string][]string)
	}
	d.ds[name] = append(d.ds[name], value)
}

func (d *directives) lookup(name string) (values []string) {
	if d.ds == nil {
		return nil
	}
	return d.ds[name]
}

func (d *directives) has(name string) bool {
	if d.ds == nil {
		return false
	}
	return d.ds[name] != nil
}

func (d *directives) get(name string) string {
	vals := d.lookup(name)
	if len(vals) == 0 {
		return ""
	}
	return vals[len(vals)-1]
}

func loadDirectives(t *testutil.T, source []byte) (d directives) {
	const prefix = "//test:"

	scanner := bufio.NewScanner(bytes.NewReader(source))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		if len(parts) != 2 {
			t.Fatalf("weird directive parsing: %q", line)
		}
		d.add(parts[0][len(prefix):], parts[1])
	}
	return d
}
