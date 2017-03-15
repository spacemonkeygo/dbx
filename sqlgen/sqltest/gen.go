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

package sqltest

import (
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/spacemonkeygo/dbx.v1/sqlgen"
	"gopkg.in/spacemonkeygo/dbx.v1/testutil"
)

type Generator struct {
	rng   *rand.Rand
	conds []*sqlgen.Condition
}

func NewGenerator(tw *testutil.T) *Generator {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	tw.Logf("seed: %d", seed)
	return &Generator{
		rng: rng,
	}
}

func (g *Generator) Gen() (out sqlgen.SQL) { return g.genRecursive(3) }

func (g *Generator) literal() sqlgen.Literal {
	return sqlgen.Literal(fmt.Sprintf("(literal %d)", g.rng.Intn(1000)))
}

func (g *Generator) condition() *sqlgen.Condition {
	if len(g.conds) == 0 || rand.Intn(2) == 0 {
		num := len(g.conds)
		condition := &sqlgen.Condition{
			Name:  fmt.Sprintf("cond%d", num),
			Field: fmt.Sprintf("field%d", num),
			Null:  rand.Intn(2) == 0,
			Equal: rand.Intn(2) == 0,
		}
		g.conds = append(g.conds, condition)
		return condition
	}
	return g.conds[rand.Intn(len(g.conds))]
}

func (g *Generator) literals(depth int) sqlgen.Literals {
	amount := rand.Intn(30)

	sqls := make([]sqlgen.SQL, amount)
	for i := range sqls {
		sqls[i] = g.genRecursive(depth - 1)
	}

	join := fmt.Sprintf("|join %d|", g.rng.Intn(1000))
	if rand.Intn(2) == 0 {
		join = ""
	}

	return sqlgen.Literals{
		Join: join,
		SQLs: sqls,
	}
}

func (g *Generator) genRecursive(depth int) (out sqlgen.SQL) {
	if depth == 0 {
		return g.literal()
	}

	switch g.rng.Intn(10) {
	case 0, 1, 2:
		return g.literal()
	case 3, 4, 5:
		return g.condition()
	default:
		return g.literals(depth)
	}
}
