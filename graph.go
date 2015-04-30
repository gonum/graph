// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

// All a node needs to do is identify itself. This allows the user to pass in nodes more
// interesting than an int, but also allow us to reap the benefits of having a map-storable,
// comparable type.
type Node interface {
	ID() int
}

// An unordered group of two nodes
type UndirectedEdge interface {
	Nodes() [2]Node
}

// DirectedEdge specifies an edge that allows the caller to distinguish between
// the front and the back of the node.
type DirectedEdge interface {
	Head() Node
	Tail() Node
}

// Out specifies a graph that can provide a list of outgoing edges
// for an input node. The outgoing edges are expected to have Head().ID()
// == n.ID().
type Out interface {
	Out(n Node) []DirectedEdge
}

// In specifies a graph that can provide a list of incoming edges
// for an input node. The incoming edges are expected to have
// Tail().ID() == n.ID()
type In interface {
	In(n Node) []DirectedEdge
}

// EdgeProvider is a generic method that provides undirected edges coming
// to/from a node. For directed cases, take the union of the In/Out interfaces.
type EdgeProvider interface {
	Edges(n Node) []UndirectedEdge
}

type FiniteGraph interface {
	NodeList() []Node
}

// DirectedCostGraph allows a graph to dictate the cost or weight
// of an arbitrary edge.
type DirectedCostGraph interface {
	Cost(e DirectedEdge) float64
}

// UndirectedCostGraph allows a graph to dictate the cost or weight
// of an arbitrary edge.
type UndirectedCostGraph interface {
	Cost(e UndirectedEdge) float64
}

// HeuristicCostGraph means a graph can estimate
// the distance between two nodes.
type HeuristicCostGraph interface {
	HeuristicCost(n1, n2 Node) float64
}
