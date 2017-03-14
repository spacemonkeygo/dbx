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
	rng   *rand.Rand
	holes []*Hole
}

func newGenerator(tw *testutil.T) *generator {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	tw.Logf("seed: %d", seed)
	return &generator{
		rng: rng,
	}
}

func (g *generator) gen() (out SQL) { return g.genRecursive(3) }

func (g *generator) literal() Literal {
	return Literal(fmt.Sprintf("(literal %d)", g.rng.Intn(1000)))
}

func (g *generator) hole() *Hole {
	if len(g.holes) == 0 || rand.Intn(2) == 0 {
		num := len(g.holes)
		hole := &Hole{Name: fmt.Sprintf("(hole %d)", num)}
		hole.Fill(Literal(fmt.Sprintf("(filled %d)", num)))
		g.holes = append(g.holes, hole)
		return hole
	}
	return g.holes[rand.Intn(len(g.holes))]
}

func (g *generator) literals(depth int) Literals {
	amount := rand.Intn(30)

	sqls := make([]SQL, amount)
	for i := range sqls {
		sqls[i] = g.genRecursive(depth - 1)
	}

	join := fmt.Sprintf("|join %d|", g.rng.Intn(1000))
	if rand.Intn(2) == 0 {
		join = ""
	}

	return Literals{
		Join: join,
		SQLs: sqls,
	}
}

func (g *generator) genRecursive(depth int) (out SQL) {
	if depth == 0 {
		return g.literal()
	}

	switch g.rng.Intn(10) {
	case 0, 1, 2:
		return g.literal()
	case 3, 4, 5:
		return g.hole()
	default:
		return g.literals(depth)
	}
}
