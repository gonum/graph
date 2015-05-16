// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/internal"
	"github.com/gonum/graph/path"
)

func TestSimpleAStar(t *testing.T) {
	tg, err := internal.NewTileGraphFrom("" +
		"▀  ▀\n" +
		"▀▀ ▀\n" +
		"▀▀ ▀\n" +
		"▀▀ ▀",
	)
	if err != nil {
		t.Fatalf("Couldn't generate tilegraph: %v", err)
	}

	p, cost, _ := path.AStar(concrete.Node(1), concrete.Node(14), tg, nil, nil)
	if math.Abs(cost-4) > 1e-5 {
		t.Errorf("A* reports incorrect cost for simple tilegraph search")
	}

	if p == nil {
		t.Fatalf("A* fails to find path for for simple tilegraph search")
	} else {
		correctPath := []int{1, 2, 6, 10, 14}
		if len(p) != len(correctPath) {
			t.Fatalf("Astar returns wrong length path for simple tilegraph search")
		}
		for i, node := range p {
			if node.ID() != correctPath[i] {
				t.Errorf("Astar returns wrong path at step", i, "got:", node, "actual:", correctPath[i])
			}
		}
	}
}

func TestBiggerAStar(t *testing.T) {
	tg := internal.NewTileGraph(3, 3, true)

	p, cost, _ := path.AStar(concrete.Node(0), concrete.Node(8), tg, nil, nil)

	if math.Abs(cost-4) > 1e-5 || !path.IsPath(p, tg) {
		t.Error("Non-optimal or impossible path found for 3x3 grid")
	}

	tg = internal.NewTileGraph(1000, 1000, true)
	p, cost, _ = path.AStar(concrete.Node(0), concrete.Node(999*1000+999), tg, nil, nil)
	if !path.IsPath(p, tg) || cost != 1998 {
		t.Error("Non-optimal or impossible path found for 100x100 grid; cost:", cost, "path:\n"+tg.PathString(p))
	}
}

func TestObstructedAStar(t *testing.T) {
	tg := internal.NewTileGraph(10, 10, true)

	// Creates a partial "wall" down the middle row with a gap down the left side
	tg.SetPassability(4, 1, false)
	tg.SetPassability(4, 2, false)
	tg.SetPassability(4, 3, false)
	tg.SetPassability(4, 4, false)
	tg.SetPassability(4, 5, false)
	tg.SetPassability(4, 6, false)
	tg.SetPassability(4, 7, false)
	tg.SetPassability(4, 8, false)
	tg.SetPassability(4, 9, false)

	rows, cols := tg.Dimensions()
	p, cost1, expanded := path.AStar(concrete.Node(5), tg.CoordsToNode(rows-1, cols-1), tg, nil, nil)

	if !path.IsPath(p, tg) {
		t.Error("Path doesn't exist in obstructed graph")
	}

	ManhattanHeuristic := func(n1, n2 graph.Node) float64 {
		id1, id2 := n1.ID(), n2.ID()
		r1, c1 := tg.IDToCoords(id1)
		r2, c2 := tg.IDToCoords(id2)

		return math.Abs(float64(r1)-float64(r2)) + math.Abs(float64(c1)-float64(c2))
	}

	p, cost2, expanded2 := path.AStar(concrete.Node(5), tg.CoordsToNode(rows-1, cols-1), tg, nil, ManhattanHeuristic)
	if !path.IsPath(p, tg) {
		t.Error("Path doesn't exist when using heuristic on obstructed graph")
	}

	if math.Abs(cost1-cost2) > 1e-5 {
		t.Error("Cost when using admissible heuristic isn't approximately equal to cost without it")
	}

	if expanded2 > expanded {
		t.Error("Using admissible, consistent heuristic expanded more nodes than null heuristic (possible, but unlikely -- suggests an error somewhere)")
	}

}

func TestNoPathAStar(t *testing.T) {
	tg := internal.NewTileGraph(5, 5, true)

	// Creates a "wall" down the middle row
	tg.SetPassability(2, 0, false)
	tg.SetPassability(2, 1, false)
	tg.SetPassability(2, 2, false)
	tg.SetPassability(2, 3, false)
	tg.SetPassability(2, 4, false)

	rows, _ := tg.Dimensions()
	p, _, _ := path.AStar(tg.CoordsToNode(0, 2), tg.CoordsToNode(rows-1, 2), tg, nil, nil)

	if len(p) > 0 { // Note that a nil slice will return len of 0, this won't panic
		t.Error("A* finds path where none exists")
	}
}

func TestSmallAStar(t *testing.T) {
	gg := newSmallUndirected()
	heur := newSmallHeuristic()
	if ok, edge, goal := monotonic(gg, heur); !ok {
		t.Fatalf("non-monotonic heuristic.  edge: %v goal: %v", edge, goal)
	}
	for _, start := range gg.Nodes() {
		// get reference paths by Dijkstra
		dPaths, dCosts := path.Dijkstra(start, gg, nil)
		// assert that AStar finds each path
		for goalID, dPath := range dPaths {
			exp := fmt.Sprintln(dPath, dCosts[goalID])
			aPath, aCost, _ := path.AStar(start, concrete.Node(goalID), gg, nil, heur)
			got := fmt.Sprintln(aPath, aCost)
			if got != exp {
				t.Error("expected", exp, "got", got)
			}
		}
	}
}

func TestIsPath(t *testing.T) {
	dg := concrete.NewDirectedGraph(math.Inf(1))
	if !path.IsPath(nil, dg) {
		t.Error("IsPath returns false on nil path")
	}
	p := []graph.Node{concrete.Node(0)}
	if path.IsPath(p, dg) {
		t.Error("IsPath returns true on nonexistant node")
	}
	dg.AddNode(p[0])
	if !path.IsPath(p, dg) {
		t.Error("IsPath returns false on single-length path with existing node")
	}
	p = append(p, concrete.Node(1))
	dg.AddNode(p[1])
	if path.IsPath(p, dg) {
		t.Error("IsPath returns true on bad path of length 2")
	}
	dg.AddDirectedEdge(concrete.Edge{p[0], p[1]}, 1)
	if !path.IsPath(p, dg) {
		t.Error("IsPath returns false on correct path of length 2")
	}
	p[0], p[1] = p[1], p[0]
	if path.IsPath(p, dg) {
		t.Error("IsPath erroneously returns true for a reverse path")
	}
	p = []graph.Node{p[1], p[0], concrete.Node(2)}
	dg.AddDirectedEdge(concrete.Edge{p[1], p[2]}, 1)
	if !path.IsPath(p, dg) {
		t.Error("IsPath does not find a correct path for path > 2 nodes")
	}
	ug := concrete.NewUndirectedGraph(math.Inf(1))
	ug.AddUndirectedEdge(concrete.Edge{p[1], p[0]}, 1)
	ug.AddUndirectedEdge(concrete.Edge{p[1], p[2]}, 1)
	if !path.IsPath(p, ug) {
		t.Error("IsPath does not correctly account for undirected behavior")
	}
}

func ExampleBreadthFirstSearch() {
	g := concrete.NewDirectedGraph(math.Inf(1))
	var n0, n1, n2, n3 concrete.Node = 0, 1, 2, 3
	g.AddDirectedEdge(concrete.Edge{n0, n1}, 1)
	g.AddDirectedEdge(concrete.Edge{n0, n2}, 1)
	g.AddDirectedEdge(concrete.Edge{n2, n3}, 1)
	p, v := path.BreadthFirstSearch(n0, n3, g)
	fmt.Println("path:", p)
	fmt.Println("nodes visited:", v)
	// Output:
	// path: [0 2 3]
	// nodes visited: 4
}

func newSmallUndirected() *concrete.UndirectedGraph {
	eds := []struct{ n1, n2, edgeCost int }{
		{1, 2, 7},
		{1, 3, 9},
		{1, 6, 14},
		{2, 3, 10},
		{2, 4, 15},
		{3, 4, 11},
		{3, 6, 2},
		{4, 5, 7},
		{5, 6, 9},
	}
	g := concrete.NewUndirectedGraph(math.Inf(1))
	for n := concrete.Node(1); n <= 6; n++ {
		g.AddNode(n)
	}
	for _, ed := range eds {
		e := concrete.Edge{
			concrete.Node(ed.n1),
			concrete.Node(ed.n2),
		}
		g.AddUndirectedEdge(e, float64(ed.edgeCost))
	}
	return g
}

func newSmallHeuristic() func(n1, n2 graph.Node) float64 {
	nds := []struct{ id, x, y int }{
		{1, 0, 6},
		{2, 1, 0},
		{3, 8, 7},
		{4, 16, 0},
		{5, 17, 6},
		{6, 9, 8},
	}
	return func(n1, n2 graph.Node) float64 {
		i1 := n1.ID() - 1
		i2 := n2.ID() - 1
		dx := nds[i2].x - nds[i1].x
		dy := nds[i2].y - nds[i1].y
		return math.Hypot(float64(dx), float64(dy))
	}
}

type weightedEdgeList interface {
	graph.Weighter
	graph.EdgeList
}

func monotonic(g weightedEdgeList, heur func(n1, n2 graph.Node) float64) (bool, graph.Edge, graph.Node) {
	for _, goal := range g.Nodes() {
		for _, edge := range g.Edges() {
			from := edge.From()
			to := edge.To()
			if heur(from, goal) > g.Weight(edge)+heur(to, goal) {
				return false, edge, goal
			}
		}
	}
	return true, nil, nil
}

// Test for correct result on a small graph easily solvable by hand
func TestDijkstraSmall(t *testing.T) {
	g := newSmallUndirected()
	paths, lens := path.Dijkstra(concrete.Node(1), g, nil)
	s := fmt.Sprintln(len(paths), len(lens))
	for i := 1; i <= 6; i++ {
		s += fmt.Sprintln(paths[i], lens[i])
	}
	if s != `6 6
[1] 0
[1 2] 7
[1 3] 9
[1 3 4] 20
[1 3 6 5] 20
[1 3 6] 11
` {
		t.Fatal(s)
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
