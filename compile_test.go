// Copyright (C) 2017 Space Monkey, Inc.

package main

import (
	"bufio"
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestCompilation(t *testing.T) {
	tw := testutil.Wrap(t)
	fileinfos, err := ioutil.ReadDir("testdata")
	tw.AssertNoError(err)

	for _, fileinfo := range fileinfos {
		tw.Run(fileinfo.Name(), func(t *testing.T) {
			testFile(testutil.Wrap(t),
				filepath.Join("testdata", fileinfo.Name()))
		})
	}
}

func testFile(t *testutil.T, file string) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatal(val)
		}
	}()

	dir, err := ioutil.TempDir("", "dbx")
	t.AssertNoError(err)
	defer os.RemoveAll(dir)

	dbx_source, err := ioutil.ReadFile(file)
	t.AssertNoError(err)

	d := loadDirectives(t, dbx_source)

	dialects := []string{"postgres", "sqlite3"}
	if other := d.lookup("dialects"); other != nil {
		dialects = other
	}

	err = golangCmd("", dialects, "", file, dir)
	t.AssertNoError(err)

	go_source, err := ioutil.ReadFile(
		filepath.Join(dir, filepath.Base(file)+".go"))
	t.AssertNoError(err)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, go_source, parser.AllErrors)
	t.AssertNoError(err)

	config := types.Config{
		Importer: importer.Default(),
	}
	_, err = config.Check(dir, fset, []*ast.File{f}, nil)

	if d.has("fail") {
		t.AssertError(err, d.get("fail"))
	} else {
		t.AssertNoError(err)
	}
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
		parts := strings.SplitN(line, " ", 1)
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
