// Copyright (C) 2017 Space Monkey, Inc.

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"testing"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

func TestRun(t *testing.T) {
	tw := testutil.Wrap(t)
	tw.Parallel()

	data_dir := filepath.Join("testdata", "run")

	names, err := filepath.Glob(filepath.Join(data_dir, "*.dbx"))
	tw.AssertNoError(err)

	for _, name := range names {
		name := name
		tw.Runp(filepath.Base(name), func(tw *testutil.T) {
			testRunFile(tw, name)
		})
	}
}

func testRunFile(t *testutil.T, dbx_file string) {
	defer func() {
		if val := recover(); val != nil {
			t.Fatalf("%s\n%s", val, string(debug.Stack()))
		}
	}()

	dbx_source, err := ioutil.ReadFile(dbx_file)
	t.AssertNoError(err)
	t.Context("dbx", linedSource(dbx_source))
	d := loadDirectives(t, dbx_source)

	dir, err := ioutil.TempDir("", "dbx")
	t.AssertNoError(err)
	defer os.RemoveAll(dir)

	t.Logf("[%s] generating... {rx:%t, userdata:%t}", dbx_file,
		d.has("rx"), d.has("userdata"))
	err = golangCmd("main", []string{"sqlite3"}, "",
		d.has("rx"), d.has("userdata"), dbx_file, dir)
	if d.has("fail_gen") {
		t.AssertError(err, d.get("fail_gen"))
		return
	} else {
		t.AssertNoError(err)
	}

	ext := filepath.Ext(dbx_file)
	go_file := dbx_file[:len(dbx_file)-len(ext)] + ".go"
	go_source, err := ioutil.ReadFile(go_file)
	t.AssertNoError(err)
	t.Context("go", linedSource(go_source))

	t.Logf("[%s] copying go source...", dbx_file)
	t.AssertNoError(ioutil.WriteFile(
		filepath.Join(dir, filepath.Base(go_file)), go_source, 0644))

	t.Logf("[%s] running output...", dbx_file)
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	t.AssertNoError(err)

	var stdout, stderr bytes.Buffer

	cmd := exec.Command("go", append([]string{"run"}, files...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	t.Context("stdout", stdout.String())
	t.Context("stderr", stderr.String())

	if d.has("fail") {
		t.AssertError(err, "")
		t.AssertContains(stderr.String(), d.get("fail"))
	} else {
		t.AssertNoError(err)
	}
}
