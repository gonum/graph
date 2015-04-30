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

// InOut means that a graph can generate predecessor and
// successors, the InOut function has no guarantees about which is head
// or tail.
//
// Note that this interface technically embeds EdgeProvider, but doesn't
// explicitly do so for semantic reasons.
type InOut interface {
	In
	Out
	InOut(n Node) []DirectedEdge
	// Provides the UNION of edges from InOut, but returns the edges
	// in unspecified order.
	Edges(n Node) []UndirectedEdge
	Degree(n Node) uint
}

// EdgeProvider is a generic method that provides undirected edges coming
// to/from a node. For directed cases, use the InOut interface.
type EdgeProvider interface {
	Edges(n Node) []UndirectedEdge
	Degree(n Node) uint
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

// NodeVerifier tests whether a node
// is in the graph (or could be in
// the graph for infinite cases).
type NodeVerifier interface {
	NodeExists(n Node) bool
}

// SuccessorVerifier tests whether the node n2
// is a successor of n1. (That is, whether there
// exists some edge such edge.Head() == n1 and
// edge.Tail() == n2).
//
// Keep in mind that you automatically have
// a predecessor verifier by implementing this
// simply by reordering your nodes.
type SuccessorVerifier interface {
	IsSuccessor(n1, n2 Node) bool
}

// NeighborVerifier tests whether the nodes
// n1 and n2 are the two ends of some edge in the graph.
// This is agnostic to whether the edge is undirected or directed.
type NeighborVerifier interface {
	IsNeighbor(n1, n2 Node) bool
}

// NodeSuccessorVerifier can test if a node is in
// the graph and whether one node is a successor of another.
type NodeSuccessorVerifier interface {
	NodeVerifier
	SuccessorVerifier
}

// NodeNeighborVerifier can test if a node in in a graph,
// as well as whether two nodes are adjacent. This is essentially
// a general Verifier for an undirected graph.
type NodeNeighborVerifier interface {
	NodeVerifier
	NeighborVerifier
}

// Verifier can verify any node property of a graph.
type Verifier interface {
	NodeVerifier
	NeighborVerifier
	SuccessorVerifier
}

// FiniteForwardGraph has a finite number of nodes
// for which it can provide successors.
type FiniteForwardGraph interface {
	Out
	FiniteGraph
}

// FiniteBackwardGraph has a finite number of nodes
// for which it can provide predecessors.
type FiniteBackwardGraph interface {
	In
	FiniteGraph
}

// FiniteDirectedGraph has a finite number of nodes
// for which it can provide successors or predecessors.
type FiniteDirectedGraph interface {
	InOut
	FiniteGraph
}

// FiniteUndirectedGraph has a finite number of nodes
// for which it can provide its neighbors.
type FiniteUndirectedGraph interface {
	EdgeProvider
	FiniteGraph
}
