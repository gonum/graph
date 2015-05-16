// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package structure

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"testing"

	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/internal"
)

type interval struct{ start, end int }

var tarjanTests = []struct {
	g []set

	ambiguousOrder []interval
	want           [][]int
}{
	{
		g: []set{
			0: linksTo(1),
			1: linksTo(2, 7),
			2: linksTo(3, 6),
			3: linksTo(4),
			4: linksTo(2, 5),
			6: linksTo(3, 5),
			7: linksTo(0, 6),
		},

		want: [][]int{
			{5},
			{2, 3, 4, 6},
			{0, 1, 7},
		},
	},
	{
		g: []set{
			0: linksTo(1, 2, 3),
			1: linksTo(2),
			2: linksTo(3),
			3: linksTo(1),
		},

		want: [][]int{
			{1, 2, 3},
			{0},
		},
	},
	{
		g: []set{
			0: linksTo(1),
			1: linksTo(0, 2),
			2: linksTo(1),
		},

		want: [][]int{
			{0, 1, 2},
		},
	},
	{
		g: []set{
			0: linksTo(1),
			1: linksTo(2, 3),
			2: linksTo(4, 5),
			3: linksTo(4, 5),
			4: linksTo(6),
			5: nil,
			6: nil,
		},

		// Node pairs (2, 3) and (4, 5) are not
		// relatively orderable within each pair.
		ambiguousOrder: []interval{
			{0, 3}, // This includes node 6 since it only needs to be before 4 in topo sort.
			{3, 5},
		},
		want: [][]int{
			{6}, {5}, {4}, {3}, {2}, {1}, {0},
		},
	},
	{
		g: []set{
			0: linksTo(1),
			1: linksTo(2, 3, 4),
			2: linksTo(0, 3),
			3: linksTo(4),
			4: linksTo(3),
		},

		// SCCs are not relatively ordable.
		ambiguousOrder: []interval{
			{0, 2},
		},
		want: [][]int{
			{0, 1, 2},
			{3, 4},
		},
	},
}

func TestTarjanSCC(t *testing.T) {
	for i, test := range tarjanTests {
		g := concrete.NewDirectedGraph(math.Inf(1))
		for u, e := range test.g {
			if !g.Has(concrete.Node(u)) {
				g.AddNode(concrete.Node(u))
			}
			for v := range e {
				if !g.Has(concrete.Node(v)) {
					g.AddNode(concrete.Node(v))
				}
				g.AddDirectedEdge(concrete.Edge{F: concrete.Node(u), T: concrete.Node(v)}, 0)
			}
		}
		gotSCCs := TarjanSCC(g)
		// tarjan.strongconnect does range iteration over maps,
		// so sort SCC members to ensure consistent ordering.
		gotIDs := make([][]int, len(gotSCCs))
		for i, scc := range gotSCCs {
			gotIDs[i] = make([]int, len(scc))
			for j, id := range scc {
				gotIDs[i][j] = id.ID()
			}
			sort.Ints(gotIDs[i])
		}
		for _, iv := range test.ambiguousOrder {
			sort.Sort(internal.BySliceValues(test.want[iv.start:iv.end]))
			sort.Sort(internal.BySliceValues(gotIDs[iv.start:iv.end]))
		}
		if !reflect.DeepEqual(gotIDs, test.want) {
			t.Errorf("unexpected Tarjan scc result for %d:\n\tgot:%v\n\twant:%v", i, gotIDs, test.want)
		}
	}
}

// batageljZaversnikGraph is the example graph from
// figure 1 of http://arxiv.org/abs/cs/0310049v1
var batageljZaversnikGraph = []set{
	0: nil,

	1: linksTo(2, 3),
	2: linksTo(4),
	3: linksTo(4),
	4: linksTo(5),
	5: nil,

	6:  linksTo(7, 8, 14),
	7:  linksTo(8, 11, 12, 14),
	8:  linksTo(14),
	9:  linksTo(11),
	10: linksTo(11),
	11: linksTo(12),
	12: linksTo(18),
	13: linksTo(14, 15),
	14: linksTo(15, 17),
	15: linksTo(16, 17),
	16: nil,
	17: linksTo(18, 19, 20),
	18: linksTo(19, 20),
	19: linksTo(20),
	20: nil,
}

var vOrderTests = []struct {
	g        []set
	wantCore [][]int
	wantK    int
}{
	{
		g: []set{
			0: linksTo(1, 2, 4, 6),
			1: linksTo(2, 4, 6),
			2: linksTo(3, 6),
			3: linksTo(4, 5),
			4: linksTo(6),
			5: nil,
			6: nil,
		},
		wantCore: [][]int{
			{},
			{5},
			{3},
			{0, 1, 2, 4, 6},
		},
		wantK: 3,
	},
	{
		g: batageljZaversnikGraph,
		wantCore: [][]int{
			{0},
			{5, 9, 10, 16},
			{1, 2, 3, 4, 11, 12, 13, 15},
			{6, 7, 8, 14, 17, 18, 19, 20},
		},
		wantK: 3,
	},
}

func TestVertexOrdering(t *testing.T) {
	for i, test := range vOrderTests {
		g := concrete.NewUndirectedGraph(math.Inf(1))
		for u, e := range test.g {
			if !g.Has(concrete.Node(u)) {
				g.AddNode(concrete.Node(u))
			}
			for v := range e {
				if !g.Has(concrete.Node(v)) {
					g.AddNode(concrete.Node(v))
				}
				g.AddUndirectedEdge(concrete.Edge{F: concrete.Node(u), T: concrete.Node(v)}, 0)
			}
		}
		order, core := VertexOrdering(g)
		if len(core)-1 != test.wantK {
			t.Errorf("unexpected value of k for test %d: got: %d want: %d", i, len(core)-1, test.wantK)
		}
		var offset int
		for k, want := range test.wantCore {
			sort.Ints(want)
			got := make([]int, len(want))
			for j, n := range order[len(order)-len(want)-offset : len(order)-offset] {
				got[j] = n.ID()
			}
			sort.Ints(got)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected %d-core for test %d:\ngot: %v\nwant:%v", got, test.wantCore)
			}

			for j, n := range core[k] {
				got[j] = n.ID()
			}
			sort.Ints(got)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected %d-core for test %d:\ngot: %v\nwant:%v", got, test.wantCore)
			}
			offset += len(want)
		}
	}
}

var bronKerboschTests = []struct {
	g    []set
	want [][]int
}{
	{
		// This is the example given in the Bron-Kerbosch article on wikipedia (renumbered).
		// http://en.wikipedia.org/w/index.php?title=Bron%E2%80%93Kerbosch_algorithm&oldid=656805858
		g: []set{
			0: linksTo(1, 4),
			1: linksTo(2, 4),
			2: linksTo(3),
			3: linksTo(4, 5),
			4: nil,
			5: nil,
		},
		want: [][]int{
			{0, 1, 4},
			{1, 2},
			{2, 3},
			{3, 4},
			{3, 5},
		},
	},
	{
		g: batageljZaversnikGraph,
		want: [][]int{
			{0},
			{1, 2},
			{1, 3},
			{2, 4},
			{3, 4},
			{4, 5},
			{6, 7, 8, 14},
			{7, 11, 12},
			{9, 11},
			{10, 11},
			{12, 18},
			{13, 14, 15},
			{14, 15, 17},
			{15, 16},
			{17, 18, 19, 20},
		},
	},
}

func TestBronKerbosch(t *testing.T) {
	for i, test := range bronKerboschTests {
		g := concrete.NewUndirectedGraph(math.Inf(1))
		for u, e := range test.g {
			if !g.Has(concrete.Node(u)) {
				g.AddNode(concrete.Node(u))
			}
			for v := range e {
				if !g.Has(concrete.Node(v)) {
					g.AddNode(concrete.Node(v))
				}
				g.AddUndirectedEdge(concrete.Edge{F: concrete.Node(u), T: concrete.Node(v)}, 0)
			}
		}
		cliques := BronKerbosch(g)
		got := make([][]int, len(cliques))
		for j, c := range cliques {
			ids := make([]int, len(c))
			for k, n := range c {
				ids[k] = n.ID()
			}
			sort.Ints(ids)
			got[j] = ids
		}
		sort.Sort(internal.BySliceValues(got))
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("unexpected cliques for test %d:\ngot: %v\nwant:%v", i, got, test.want)
		}
	}
}

var connectedComponentTests = []struct {
	g    []set
	want [][]int
}{
	{
		g: batageljZaversnikGraph,
		want: [][]int{
			{0},
			{1, 2, 3, 4, 5},
			{6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
	},
}

func TestConnectedComponents(t *testing.T) {
	for i, test := range connectedComponentTests {
		g := concrete.NewUndirectedGraph(math.Inf(1))

		for u, e := range test.g {
			if !g.Has(concrete.Node(u)) {
				g.AddNode(concrete.Node(u))
			}
			for v := range e {
				if !g.Has(concrete.Node(v)) {
					g.AddNode(concrete.Node(v))
				}
				g.AddUndirectedEdge(concrete.Edge{F: concrete.Node(u), T: concrete.Node(v)}, 0)
			}
		}
		cc := ConnectedComponents(g)
		got := make([][]int, len(cc))
		for j, c := range cc {
			ids := make([]int, len(c))
			for k, n := range c {
				ids[k] = n.ID()
			}
			sort.Ints(ids)
			got[j] = ids
		}
		sort.Sort(internal.BySliceValues(got))
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("unexpected connected components for test %d %T:\ngot: %v\nwant:%v", i, g, got, test.want)
		}
		fmt.Println()
	}
}

// set is an integer set.
type set map[int]struct{}

func linksTo(i ...int) set {
	if len(i) == 0 {
		return nil
	}
	s := make(set)
	for _, v := range i {
		s[v] = struct{}{}
	}
	return s
}
