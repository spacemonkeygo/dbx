// Copyright (C) 2017 Space Monkey, Inc.

package testutil

import (
	"sort"
	"strings"
	"testing"
)

func Wrap(t *testing.T) *T {
	return &T{
		T: t,

		context: make(map[string]string),
	}
}

type T struct {
	*testing.T

	context map[string]string
}

func (t *T) Context(name string, val string) {
	t.context[name] = val
}

func (t *T) dumpContext() {
	var keys []string
	for key := range t.context {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		t.Logf("%s:\n%s", key, t.context[key])
	}
}

func (t *T) AssertNoError(err error) {
	if err != nil {
		t.dumpContext()
		t.Fatalf("expected no error. got %v", err)
	}
}

func (t *T) AssertError(err error, msg string) {
	if err == nil {
		t.dumpContext()
		t.Fatalf("expected an error containing %q. got nil", msg)
	}
	if !strings.Contains(err.Error(), msg) {
		t.dumpContext()
		t.Fatalf("expected an error containing %q. got %v", msg, err)
	}
}

func (t *T) Run(name string, f func(*T)) bool {
	return t.T.Run(name, func(t *testing.T) { f(Wrap(t)) })
}

func (t *T) Runp(name string, f func(*T)) bool {
	return t.T.Run(name, func(t *testing.T) { t.Parallel(); f(Wrap(t)) })
}
