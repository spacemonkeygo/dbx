// Copyright (C) 2017 Space Monkey, Inc.

package testutil

import (
	"strings"
	"testing"
)

func Wrap(t *testing.T) *T {
	return &T{T: t}
}

type T struct {
	*testing.T
}

func (t *T) AssertNoError(err error) {
	if err != nil {
		t.Fatalf("expected no error. got %v", err)
	}
}

func (t *T) AssertError(err error, msg string) {
	if err == nil {
		t.Fatalf("expected an error containing %q. got nil", msg)
	}
	if !strings.Contains(err.Error(), msg) {
		t.Fatalf("expected an error containing %q. got %v", err)
	}
}
