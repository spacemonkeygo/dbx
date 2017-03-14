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

package sqlgen

import (
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

type generator struct {
	rng *rand.Rand
}

func newGenerator(tw *testutil.T) *generator {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	tw.Logf("seed: %d", seed)
	return &generator{
		rng: rng,
	}
}

func (g *generator) gen() (out SQL) { return g.genRecursive(10) }

func (g *generator) literal() Literal {
	return Literal(fmt.Sprintf("(literal %d)", g.rng.Intn(1000)))
}

func (g *generator) literals(depth int) Literals {
	amount := rand.Intn(10)

	sqls := make([]SQL, amount)
	for i := range sqls {
		sqls[i] = g.genRecursive(depth - 1)
	}

	return Literals{
		Join: fmt.Sprintf("|join %d|", g.rng.Intn(1000)),
		SQLs: sqls,
	}
}

func (g *generator) genRecursive(depth int) (out SQL) {
	if depth == 0 {
		return g.literal()
	}

	switch g.rng.Intn(10) {
	case 0, 1, 2, 3, 4, 5, 6:
		return g.literal()
	default:
		return g.literals(depth)
	}
}
