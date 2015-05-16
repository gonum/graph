// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package concrete

import (
	"github.com/gonum/graph"
)

// A simple int alias.
type Node int

func (n Node) ID() int {
	return int(n)
}

// Just a collection of two nodes
type Edge struct {
	F, T graph.Node
}

func (e Edge) From() graph.Node {
	return e.F
}

func (e Edge) To() graph.Node {
	return e.T
}

type WeightedEdge struct {
	graph.Edge
	Weight float64
}

// Undirected graph that can handle an arbitrary number of vertices and edges.
//
// Internally, it uses a map of successors AND predecessors, to speed up some operations (such as
// getting all successors/predecessors). It also speeds up things like adding edges (assuming both
// edges exist).
//
// However, its generality is also its weakness (and partially a flaw in needing to satisfy
// MutableGraph). For most purposes, creating your own graph is probably better. For instance,
// see TileGraph for an example of an immutable 2D grid of tiles that also implements the UndirectedGraph
// interface, but would be more suitable if all you needed was a simple undirected 2D grid.
type UndirectedGraph struct {
	absent float64

	neighbors map[int]map[int]WeightedEdge
	nodeMap   map[int]graph.Node

	// Node add/remove convenience vars
	maxID   int
	freeMap map[int]struct{}
}

func NewUndirectedGraph(absent float64) *UndirectedGraph {
	return &UndirectedGraph{
		absent: absent,

		neighbors: make(map[int]map[int]WeightedEdge),
		nodeMap:   make(map[int]graph.Node),
		maxID:     0,
		freeMap:   make(map[int]struct{}),
	}
}

/* Mutable implementation */

func (g *UndirectedGraph) NewNode() graph.Node {
	if g.maxID != maxInt {
		g.maxID++
		return Node(g.maxID)
	}

	// Implicitly checks if len(g.freeMap) == 0
	for id := range g.freeMap {
		return Node(id)
	}

	// I cannot foresee this ever happening, but just in case, we check.
	if len(g.nodeMap) == maxInt {
		panic("cannot allocate node: graph too large")
	}

	for i := 0; i < maxInt; i++ {
		if _, ok := g.nodeMap[i]; !ok {
			return Node(i)
		}
	}

	// Should not happen.
	panic("cannot allocate node id: no free id found")
}

func (g *UndirectedGraph) AddNode(n graph.Node) {
	g.nodeMap[n.ID()] = n
	g.neighbors[n.ID()] = make(map[int]WeightedEdge)

	delete(g.freeMap, n.ID())
	g.maxID = max(g.maxID, n.ID())
}

func (g *UndirectedGraph) AddUndirectedEdge(e graph.Edge, weight float64) {
	from, to := e.From(), e.To()
	if !g.Has(from) {
		g.AddNode(from)
	}

	if !g.Has(to) {
		g.AddNode(to)
	}

	g.neighbors[from.ID()][to.ID()] = WeightedEdge{Edge: e, Weight: weight}
	g.neighbors[to.ID()][from.ID()] = WeightedEdge{Edge: e, Weight: weight}
}

func (g *UndirectedGraph) RemoveNode(n graph.Node) {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return
	}
	delete(g.nodeMap, n.ID())

	for neigh := range g.neighbors[n.ID()] {
		delete(g.neighbors[neigh], n.ID())
	}
	delete(g.neighbors, n.ID())

	if g.maxID != 0 && n.ID() == g.maxID {
		g.maxID--
	}
	g.freeMap[n.ID()] = struct{}{}
}

func (g *UndirectedGraph) RemoveUndirectedEdge(e graph.Edge) {
	from, to := e.From(), e.To()
	if _, ok := g.nodeMap[from.ID()]; !ok {
		return
	} else if _, ok := g.nodeMap[to.ID()]; !ok {
		return
	}

	delete(g.neighbors[from.ID()], to.ID())
	delete(g.neighbors[to.ID()], from.ID())
}

func (g *UndirectedGraph) EmptyGraph() {
	g.neighbors = make(map[int]map[int]WeightedEdge)
	g.nodeMap = make(map[int]graph.Node)
}

/* UndirectedGraph implementation */

func (g *UndirectedGraph) From(n graph.Node) []graph.Node {
	if !g.Has(n) {
		return nil
	}

	neighbors := make([]graph.Node, len(g.neighbors[n.ID()]))
	i := 0
	for id := range g.neighbors[n.ID()] {
		neighbors[i] = g.nodeMap[id]
		i++
	}

	return neighbors
}

func (g *UndirectedGraph) HasEdge(n, neigh graph.Node) bool {
	_, ok := g.neighbors[n.ID()][neigh.ID()]
	return ok
}

func (g *UndirectedGraph) EdgeBetween(n, neigh graph.Node) graph.Edge {
	// Don't need to check if neigh exists because
	// it's implicit in the neighbors access.
	if !g.Has(n) {
		return nil
	}

	return g.neighbors[n.ID()][neigh.ID()]
}

func (g *UndirectedGraph) Has(n graph.Node) bool {
	_, ok := g.nodeMap[n.ID()]

	return ok
}

func (g *UndirectedGraph) Order() int {
	return len(g.nodeMap)
}

func (g *UndirectedGraph) Nodes() []graph.Node {
	nodes := make([]graph.Node, len(g.nodeMap))
	i := 0
	for _, n := range g.nodeMap {
		nodes[i] = n
		i++
	}

	return nodes
}

func (g *UndirectedGraph) Weight(e graph.Edge) float64 {
	if n, ok := g.neighbors[e.From().ID()]; ok {
		if we, ok := n[e.To().ID()]; ok {
			return we.Weight
		}
	}
	return g.absent
}

func (g *UndirectedGraph) Edges() []graph.Edge {
	m := make(map[WeightedEdge]struct{})
	toReturn := make([]graph.Edge, 0)

	for _, neighs := range g.neighbors {
		for _, we := range neighs {
			if _, ok := m[we]; !ok {
				m[we] = struct{}{}
				toReturn = append(toReturn, we.Edge)
			}
		}
	}

	return toReturn
}

func (g *UndirectedGraph) Degree(n graph.Node) int {
	if _, ok := g.nodeMap[n.ID()]; !ok {
		return 0
	}

	return len(g.neighbors[n.ID()])
}
