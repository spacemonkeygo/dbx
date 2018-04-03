// Copyright (C) 2017 Space Monkey, Inc.

package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestBuild(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	data_dir := filepath.Join("testdata", "build")

	names, err := filepath.Glob(filepath.Join(data_dir, "*.dbx"))
	tw.AssertNoError(err)

	for _, name := range names {
		name := name
		tw.Runp(filepath.Base(name), func(tw *testutil.T) {
			testBuildFile(tw, name)
		})
	}
}

func testBuildFile(t *testutil.T, file string) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatalf("%s\n%s", val, string(debug.Stack()))
		}
	}()

	dir, err := ioutil.TempDir("", "dbx")
	t.AssertNoError(err)
	defer os.RemoveAll(dir)

	dbx_source, err := ioutil.ReadFile(file)
	t.AssertNoError(err)
	t.Context("dbx", linedSource(dbx_source))
	d := loadDirectives(t, dbx_source)

	dialects := []string{"postgres", "sqlite3"}
	if other := d.lookup("dialects"); other != nil {
		dialects = other
		t.Logf("using dialects: %q", dialects)
	}

	type options struct {
		rx       bool
		userdata bool
	}

	runBuild := func(opts options) {
		t.Logf("[%s] generating... %+v", file, opts)
		err = golangCmd("", dialects, "", opts.rx, opts.userdata, file, dir)
		if d.has("fail_gen") {
			t.AssertError(err, d.get("fail_gen"))
			return
		} else {
			t.AssertNoError(err)
		}

		t.Logf("[%s] loading...", file)
		go_file := filepath.Join(dir, filepath.Base(file)+".go")
		go_source, err := ioutil.ReadFile(go_file)
		t.AssertNoError(err)
		t.Context("go", linedSource(go_source))

		t.Logf("[%s] parsing...", file)
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, go_file, go_source, parser.AllErrors)
		t.AssertNoError(err)

		t.Logf("[%s] compiling...", file)
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

	runBuild(options{rx: false, userdata: false})
	runBuild(options{rx: false, userdata: true})
	runBuild(options{rx: true, userdata: false})
	runBuild(options{rx: true, userdata: true})
}
