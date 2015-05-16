// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package structure

import (
	"sort"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/internal"
	"github.com/gonum/graph/traverse"
)

// TarjanSCC returns the strongly connected components of the graph g using Tarjan's algorithm.
//
// A strongly connected component of a graph is a set of vertices where it's possible to reach any
// vertex in the set from any other (meaning there's a cycle between them.)
//
// Generally speaking, a directed graph where the number of strongly connected components is equal
// to the number of nodes is acyclic, unless you count reflexive edges as a cycle (which requires
// only a little extra testing.)
//
func TarjanSCC(g graph.Directed) [][]graph.Node {
	nodes := g.Nodes()
	t := tarjan{
		succ: g.From,

		indexTable: make(map[int]int, len(nodes)),
		lowLink:    make(map[int]int, len(nodes)),
		onStack:    make(internal.IntSet, len(nodes)),
	}
	for _, v := range nodes {
		if t.indexTable[v.ID()] == 0 {
			t.strongconnect(v)
		}
	}
	return t.sccs
}

// tarjan implements Tarjan's strongly connected component finding
// algorithm. The implementation is from the pseudocode at
//
// http://en.wikipedia.org/wiki/Tarjan%27s_strongly_connected_components_algorithm?oldid=642744644
//
type tarjan struct {
	succ func(graph.Node) []graph.Node

	index      int
	indexTable map[int]int
	lowLink    map[int]int
	onStack    internal.IntSet

	stack []graph.Node

	sccs [][]graph.Node
}

// strongconnect is the strongconnect function described in the
// wikipedia article.
func (t *tarjan) strongconnect(v graph.Node) {
	vID := v.ID()

	// Set the depth index for v to the smallest unused index.
	t.index++
	t.indexTable[vID] = t.index
	t.lowLink[vID] = t.index
	t.stack = append(t.stack, v)
	t.onStack.Add(vID)

	// Consider successors of v.
	for _, w := range t.succ(v) {
		wID := w.ID()
		if t.indexTable[wID] == 0 {
			// Successor w has not yet been visited; recur on it.
			t.strongconnect(w)
			t.lowLink[vID] = min(t.lowLink[vID], t.lowLink[wID])
		} else if t.onStack.Has(wID) {
			// Successor w is in stack s and hence in the current SCC.
			t.lowLink[vID] = min(t.lowLink[vID], t.indexTable[wID])
		}
	}

	// If v is a root node, pop the stack and generate an SCC.
	if t.lowLink[vID] == t.indexTable[vID] {
		// Start a new strongly connected component.
		var (
			scc []graph.Node
			w   graph.Node
		)
		for {
			w, t.stack = t.stack[len(t.stack)-1], t.stack[:len(t.stack)-1]
			t.onStack.Remove(w.ID())
			// Add w to current strongly connected component.
			scc = append(scc, w)
			if w.ID() == vID {
				break
			}
		}
		// Output the current strongly connected component.
		t.sccs = append(t.sccs, scc)
	}
}

/* Implements minimum-spanning tree algorithms;
puts the resulting minimum spanning tree in the dst graph */

// Generates a minimum spanning tree with sets.
//
// As with other algorithms that use Weight, the order of precedence is
// Argument > Interface > UniformCost.
//
// The destination must be empty (or at least disjoint with the node IDs of the input)
func Prim(dst graph.MutableUndirected, g graph.EdgeList, weight graph.WeightFunc) {
	if weight == nil {
		if g, ok := g.(graph.Weighter); ok {
			weight = g.Weight
		} else {
			weight = graph.UniformCostWeight
		}
	}

	nlist := g.Nodes()
	if len(nlist) == 0 {
		return
	}

	dst.AddNode(nlist[0])
	remainingNodes := make(internal.IntSet)
	for _, node := range nlist[1:] {
		remainingNodes.Add(node.ID())
	}

	edgeList := g.Edges()
	for remainingNodes.Count() != 0 {
		var edges []concrete.WeightedEdge
		for _, edge := range edgeList {
			if (dst.Has(edge.From()) && remainingNodes.Has(edge.To().ID())) ||
				(dst.Has(edge.To()) && remainingNodes.Has(edge.From().ID())) {

				edges = append(edges, concrete.WeightedEdge{Edge: edge, Weight: weight(edge)})
			}
		}

		sort.Sort(byWeight(edges))
		myEdge := edges[0]

		dst.AddUndirectedEdge(myEdge.Edge, myEdge.Weight)
		remainingNodes.Remove(myEdge.Edge.From().ID())
	}

}

// Generates a minimum spanning tree for a graph using discrete.DisjointSet.
//
// As with other algorithms with Cost, the precedence goes Argument > Interface > UniformCost.
//
// The destination must be empty (or at least disjoint with the node IDs of the input)
func Kruskal(dst graph.MutableUndirected, g graph.EdgeList, weight graph.WeightFunc) {
	if weight == nil {
		if g, ok := g.(graph.Weighter); ok {
			weight = g.Weight
		} else {
			weight = graph.UniformCostWeight
		}
	}

	edgeList := g.Edges()
	edges := make([]concrete.WeightedEdge, 0, len(edgeList))
	for _, edge := range edgeList {
		edges = append(edges, concrete.WeightedEdge{Edge: edge, Weight: weight(edge)})
	}

	sort.Sort(byWeight(edges))

	ds := internal.NewDisjointSet()
	for _, node := range g.Nodes() {
		ds.MakeSet(node.ID())
	}

	for _, edge := range edges {
		// The disjoint set doesn't really care for which is head and which is tail so this
		// should work fine without checking both ways
		if s1, s2 := ds.Find(edge.Edge.From().ID()), ds.Find(edge.Edge.To().ID()); s1 != s2 {
			ds.Union(s1, s2)
			dst.AddUndirectedEdge(edge.Edge, edge.Weight)
		}
	}
}

type byWeight []concrete.WeightedEdge

func (e byWeight) Len() int {
	return len(e)
}

func (e byWeight) Less(i, j int) bool {
	return e[i].Weight < e[j].Weight
}

func (e byWeight) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

// VertexOrdering returns the vertex ordering and the k-cores of
// the undirected graph g.
func VertexOrdering(g graph.Graph) (order []graph.Node, cores [][]graph.Node) {
	nodes := g.Nodes()

	// The algorithm used here is essentially as described at
	// http://en.wikipedia.org/w/index.php?title=Degeneracy_%28graph_theory%29&oldid=640308710

	// Initialize an output list L.
	var l []graph.Node

	// Compute a number d_v for each vertex v in G,
	// the number of neighbors of v that are not already in L.
	// Initially, these numbers are just the degrees of the vertices.
	dv := make(map[int]int, len(nodes))
	var (
		maxDegree  int
		neighbours = make(map[int][]graph.Node)
	)
	for _, n := range nodes {
		adj := g.From(n)
		neighbours[n.ID()] = adj
		dv[n.ID()] = len(adj)
		if len(adj) > maxDegree {
			maxDegree = len(adj)
		}
	}

	// Initialize an array D such that D[i] contains a list of the
	// vertices v that are not already in L for which d_v = i.
	d := make([][]graph.Node, maxDegree+1)
	for _, n := range nodes {
		deg := dv[n.ID()]
		d[deg] = append(d[deg], n)
	}

	// Initialize k to 0.
	k := 0
	// Repeat n times:
	s := []int{0}
	for _ = range nodes { // TODO(kortschak): Remove blank assignment when go1.3.3 is no longer supported.
		// Scan the array cells D[0], D[1], ... until
		// finding an i for which D[i] is nonempty.
		var (
			i  int
			di []graph.Node
		)
		for i, di = range d {
			if len(di) != 0 {
				break
			}
		}

		// Set k to max(k,i).
		if i > k {
			k = i
			s = append(s, make([]int, k-len(s)+1)...)
		}

		// Select a vertex v from D[i]. Add v to the
		// beginning of L and remove it from D[i].
		var v graph.Node
		v, d[i] = di[len(di)-1], di[:len(di)-1]
		l = append(l, v)
		s[k]++
		delete(dv, v.ID())

		// For each neighbor w of v not already in L,
		// subtract one from d_w and move w to the
		// cell of D corresponding to the new value of d_w.
		for _, w := range neighbours[v.ID()] {
			dw, ok := dv[w.ID()]
			if !ok {
				continue
			}
			for i, n := range d[dw] {
				if n.ID() == w.ID() {
					d[dw][i], d[dw] = d[dw][len(d[dw])-1], d[dw][:len(d[dw])-1]
					dw--
					d[dw] = append(d[dw], w)
					break
				}
			}
			dv[w.ID()] = dw
		}
	}

	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	cores = make([][]graph.Node, len(s))
	offset := len(l)
	for i, n := range s {
		cores[i] = l[offset-n : offset]
		offset -= n
	}
	return l, cores
}

// BronKerbosch returns the set of maximal cliques of the undirected graph g.
func BronKerbosch(g graph.Graph) [][]graph.Node {
	nodes := g.Nodes()

	// The algorithm used here is essentially BronKerbosch3 as described at
	// http://en.wikipedia.org/w/index.php?title=Bron%E2%80%93Kerbosch_algorithm&oldid=656805858

	p := make(internal.Set, len(nodes))
	for _, n := range nodes {
		p.Add(n)
	}
	x := make(internal.Set)
	var bk bronKerbosch
	order, _ := VertexOrdering(g)
	for _, v := range order {
		neighbours := g.From(v)
		nv := make(internal.Set, len(neighbours))
		for _, n := range neighbours {
			nv.Add(n)
		}
		bk.maximalCliquePivot(g, []graph.Node{v}, make(internal.Set).Intersect(p, nv), make(internal.Set).Intersect(x, nv))
		p.Remove(v)
		x.Add(v)
	}
	return bk
}

type bronKerbosch [][]graph.Node

func (bk *bronKerbosch) maximalCliquePivot(g graph.Graph, r []graph.Node, p, x internal.Set) {
	if len(p) == 0 && len(x) == 0 {
		*bk = append(*bk, r)
		return
	}

	neighbours := bk.choosePivotFrom(g, p, x)
	nu := make(internal.Set, len(neighbours))
	for _, n := range neighbours {
		nu.Add(n)
	}
	for _, v := range p {
		if nu.Has(v) {
			continue
		}
		neighbours := g.From(v)
		nv := make(internal.Set, len(neighbours))
		for _, n := range neighbours {
			nv.Add(n)
		}

		var found bool
		for _, n := range r {
			if n.ID() == v.ID() {
				found = true
				break
			}
		}
		var sr []graph.Node
		if !found {
			sr = append(r, v)
		}

		bk.maximalCliquePivot(g, sr, make(internal.Set).Intersect(p, nv), make(internal.Set).Intersect(x, nv))
		p.Remove(v)
		x.Add(v)
	}
}

func (*bronKerbosch) choosePivotFrom(g graph.Graph, p, x internal.Set) (neighbors []graph.Node) {
	// TODO(kortschak): Investigate the impact of pivot choice that maximises
	// |p ⋂ neighbours(u)| as a function of input size. Until then, leave as
	// compile time option.
	if !tomitaTanakaTakahashi {
		for _, n := range p {
			return g.From(n)
		}
		for _, n := range x {
			return g.From(n)
		}
		panic("bronKerbosch: empty set")
	}

	var (
		max   = -1
		pivot graph.Node
	)
	maxNeighbors := func(s internal.Set) {
	outer:
		for _, u := range s {
			nb := g.From(u)
			c := len(nb)
			if c <= max {
				continue
			}
			for n := range nb {
				if _, ok := p[n]; ok {
					continue
				}
				c--
				if c <= max {
					continue outer
				}
			}
			max = c
			pivot = u
			neighbors = nb
		}
	}
	maxNeighbors(p)
	maxNeighbors(x)
	if pivot == nil {
		panic("bronKerbosch: empty set")
	}
	return neighbors
}

// ConnectedComponents returns the connected components of the graph g. All
// edges are treated as undirected.
func ConnectedComponents(g graph.Undirected) [][]graph.Node {
	var (
		w  traverse.DepthFirst
		c  []graph.Node
		cc [][]graph.Node
	)
	during := func(n graph.Node) {
		c = append(c, n)
	}
	after := func() {
		cc = append(cc, []graph.Node(nil))
		cc[len(cc)-1] = append(cc[len(cc)-1], c...)
		c = c[:0]
	}
	w.WalkAll(g, nil, after, during)

	return cc
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
