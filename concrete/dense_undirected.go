// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concrete

import (
	"github.com/gonum/graph"
	"github.com/gonum/matrix/mat64"
)

// UndirectedDenseGraph represents a graph such that all IDs are in a contiguous
// block from 0 to n-1.
type UndirectedDenseGraph struct {
	mat *mat64.SymDense
}

// NewUndirectedDenseGraph creates an undirected dense graph with n nodes.
// If passable is true all nodes will have an edge with unit cost, otherwise
// every node will start unconnected (cost of +Inf).
func NewUndirectedDenseGraph(n int, passable bool) *UndirectedDenseGraph {
	mat := make([]float64, n*n)
	v := 1.
	if !passable {
		v = inf
	}
	for i := range mat {
		mat[i] = v
	}
	return &UndirectedDenseGraph{mat64.NewSymDense(n, mat)}
}

func (g *UndirectedDenseGraph) Has(n graph.Node) bool {
	id := n.ID()
	r := g.mat.Symmetric()
	return 0 <= id && id < r
}

func (g *UndirectedDenseGraph) Order() int {
	r, _ := g.mat.Dims()
	return r
}

func (g *UndirectedDenseGraph) Nodes() []graph.Node {
	r := g.mat.Symmetric()
	nodes := make([]graph.Node, r)
	for i := 0; i < r; i++ {
		nodes[i] = Node(i)
	}
	return nodes
}

func (g *UndirectedDenseGraph) Degree(n graph.Node) int {
	id := n.ID()
	var deg int
	if g.mat.At(id, id) != inf {
		deg = 1
	}
	r := g.mat.Symmetric()
	for i := 0; i < r; i++ {
		if g.mat.At(id, i) != inf {
			deg++
		}
	}
	return deg
}

func (g *UndirectedDenseGraph) From(n graph.Node) []graph.Node {
	var neighbors []graph.Node
	id := n.ID()
	r := g.mat.Symmetric()
	for i := 0; i < r; i++ {
		if g.mat.At(id, i) != inf {
			neighbors = append(neighbors, Node(i))
		}
	}
	return neighbors
}

func (g *UndirectedDenseGraph) HasEdge(n, neighbor graph.Node) bool {
	return g.mat.At(n.ID(), neighbor.ID()) != inf
}

func (g *UndirectedDenseGraph) EdgeBetween(n, neighbor graph.Node) graph.Edge {
	if g.HasEdge(n, neighbor) {
		return Edge{n, neighbor}
	}
	return nil
}

func (g *UndirectedDenseGraph) Cost(e graph.Edge) float64 {
	return g.mat.At(e.Head().ID(), e.Tail().ID())
}

func (g *UndirectedDenseGraph) SetEdgeCost(e graph.Edge, cost float64) {
	g.mat.SetSym(e.Head().ID(), e.Tail().ID(), cost)
}

func (g *UndirectedDenseGraph) RemoveEdge(e graph.Edge) {
	g.mat.SetSym(e.Head().ID(), e.Tail().ID(), inf)
}

func (g *UndirectedDenseGraph) Matrix() *mat64.SymDense {
	// Prevent alteration of dimentions of the returned matrix.
	m := *g.mat
	return &m
}

func (g *UndirectedDenseGraph) Crunch() {}
