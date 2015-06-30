// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamic_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/path"
	"github.com/gonum/graph/path/dynamic"
)

func TestAStarNullHeuristic(t *testing.T) {
	for _, test := range shortestPathTests {
		// Skip zero-weight cycles.
		if strings.HasPrefix(test.name, "zero-weight") {
			continue
		}

		g := test.g()
		for _, e := range test.edges {
			g.SetEdge(e, e.Cost)
		}

		var (
			d *dynamic.DStarLite

			panicked bool
		)
		func() {
			defer func() {
				panicked = recover() != nil
			}()
			d = dynamic.NewDStarLite(test.query.From(), test.query.To(), g.(graph.Graph), path.NullHeuristic, concrete.NewDirectedGraph())
		}()
		if panicked || test.negative {
			if !test.negative {
				t.Errorf("%q: unexpected panic", test.name)
			}
			if !panicked {
				t.Errorf("%q: expected panic for negative edge weight", test.name)
			}
			continue
		}

		p, weight := d.Path()

		if !math.IsInf(weight, 1) && p[0].ID() != test.query.From().ID() {
			t.Fatalf("%q: unexpected from node ID: got:%d want:%d", p[0].ID(), test.query.From().ID())
		}
		if weight != test.weight {
			t.Errorf("%q: unexpected weight from Between: got:%f want:%f",
				test.name, weight, test.weight)
		}

		var got []int
		for _, n := range p {
			got = append(got, n.ID())
		}
		ok := len(got) == 0 && len(test.want) == 0
		for _, sp := range test.want {
			if reflect.DeepEqual(got, sp) {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("%q: unexpected shortest path:\ngot: %v\nwant from:%v",
				test.name, p, test.want)
		}

		// np, weight := pt.To(test.none.To())
		// if pt.From().ID() == test.none.From().ID() && (np != nil || !math.IsInf(weight, 1)) {
		// 	t.Errorf("%q: unexpected path:\ngot: path=%v weight=%f\nwant:path=<nil> weight=+Inf",
		// 		test.name, np, weight)
		// }
	}
}
