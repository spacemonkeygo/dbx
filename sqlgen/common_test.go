// Copyright (C) 2019 Space Monkey, Inc.
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

package sqlgen

import (
	"testing"
)

func TestFlattenSQL(t *testing.T) {
	for _, test := range []struct {
		in  string
		exp string
	}{
		{"", ""},
		{" ", ""},
		{" x ", "x"},
		{" x\t\t", "x"},
		{"  x ", "x"},
		{"\t\tx\t\t", "x"},
		{" \tx \t", "x"},
		{"\t x\t ", "x"},
		{"x\tx", "x x"},
		{"  x  ", "x"},
		{" \tx x\t ", "x x"},
		{"  x  x  x  ", "x x x"},
		{"\t\tx\t\tx\t\tx\t\t", "x x x"},
		{"  x  \n\t    x  \n\t   x  ", "x x x"},
	} {
		got := flattenSQL(test.in)
		if got != test.exp {
			t.Logf(" in: %q", test.in)
			t.Logf("got: %q", got)
			t.Logf("exp: %q", test.exp)
			t.Fail()
		}
	}
}

var benchStrings = []string{
	`INSERT INTO example ( alpha, beta, gamma, delta, iota, kappa, lambda ) VALUES ( $1, $2, $3, $4, $5, $6, $7 ) RETURNING example.alpha, example.beta, example.gamma, example.delta, example.iota, example.kappa, example.lambda;`,
	`INSERT INTO example
	 ( alpha, beta, 
	 	gamma, delta, iota,

	 	 kappa, lambda ) VALUES ( $1, $2, $3, $4, 
	 	 	$5, $6, $7 ) RETURNING example.alpha, 
	 	 	example.beta, example.gamma, 
	 	 	example.delta, example.iota, example.kappa, 

	 	 	example.lambda;`,
}

func BenchmarkFlattenSQL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range benchStrings {
			_ = flattenSQL(s)
		}
	}
}
