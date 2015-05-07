// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graph

import "math"

// All a node needs to do is identify itself. This allows the user to pass in nodes more
// interesting than an int, but also allow us to reap the benefits of having a map-storable,
// comparable type.
type Node interface {
	ID() int
}

// Allows edges to do something more interesting that just be a group of nodes. While the methods
// are called From and To, they are not considered directed unless the given interface specifies
// otherwise.
type Edge interface {
	From() Node
	To() Node
}

// A Graph implements the behavior of a graph.
type Graph interface {
	// Order returns the order of the graph.
	Order() int

	// Has returns whether the graph has the given node.
	Has(Node) bool

	// Nodes returns all nodes in the graph.
	Nodes() []Node

	// From returns all nodes that can be reached from
	// the given node.
	From(Node) []Node

	// HasEdge returns an edge exists between node u and v.
	HasEdge(u, v Node) bool
}

// A Graph implements the behavior of an undirected graph.
//
// All methods in Graph are implicitly undirected. Graph algorithms that care about directionality
// will intelligently choose the Directed behavior if that interface is also implemented,
// even if the function itself only takes in a Graph (or a super-interface of graph).
type Undirected interface {
	Graph

	// EdgeBetween returns the edge between node u and v such that
	// From is one argument and To is the other. If no
	// such edge exists, this function returns nil.
	EdgeBetween(u, v Node) Edge
}

// Directed graphs are characterized by having separable heads and tails in their edges.
// That is, if node1 goes to node2, that does not necessarily imply that node2 goes to node1.
//
// While it's possible for a directed graph to have fully reciprocal edges (i.e. the graph is
// symmetric) -- it is not required to be. The graph is also required to implement Graph
// because in many cases it can be useful to know all neighbors regardless of direction.
type Directed interface {
	Graph

	// To gives the nodes connected by INBOUND edges.
	// If the graph is an undirected graph, this set is equal to From.
	To(Node) []Node

	// EdgeFromTo returns an edge between node u and v such that
	// From is one argument and To is the other. If no
	// such edge exists, this function returns nil.
	EdgeFromTo(u, v Node) Edge
}

// Returns all undirected edges in the graph
type EdgeLister interface {
	Edges() []Edge
}

type EdgeList interface {
	Graph
	EdgeLister
}

// Returns all directed edges in the graph.
type DirectedEdgeLister interface {
	DirectedEdges() []Edge
}

type DirectedEdgeList interface {
	Graph
	DirectedEdgeLister
}

// A Graph that implements Coster has an actual cost between adjacent nodes, also known as a
// weighted graph. If a graph implements coster and a function needs to read cost (e.g. A*),
// this function will take precedence over the Uniform Cost function (all weights are 1) if "nil"
// is passed in for the function argument.
//
// If the argument is nil, or the edge is invalid for some reason, this should return math.Inf(1)
type Coster interface {
	Cost(Edge) float64
}

type CostGraph interface {
	Coster
	Graph
}

type CostDirected interface {
	Coster
	Directed
}

// A graph that implements HeuristicCoster implements a heuristic between any two given nodes.
// Like Coster, if a graph implements this and a function needs a heuristic cost (e.g. A*), this
// function will take precedence over the Null Heuristic (always returns 0) if "nil" is passed in
// for the function argument. If HeuristicCost is not intended to be used, it can be implemented as
// the null heuristic (always returns 0).
type HeuristicCoster interface {
	// HeuristicCost returns a heuristic cost between any two nodes.
	HeuristicCost(n1, n2 Node) float64
}

// A Mutable is a graph that can have arbitrary nodes and edges added or removed.
//
// Anything implementing Mutable is required to store the actual argument. So if AddNode(myNode) is
// called and later a user calls on the graph graph.NodeList(), the node added by AddNode must be
// an the exact node, not a new node with the same ID.
//
// In any case where conflict is possible (e.g. adding two nodes with the same ID), the later
// call always supercedes the earlier one.
//
// Functions will generally expect one of Mutable or MutableDirected and not Mutable
// itself. That said, any function that takes Mutable[x], the destination mutable should
// always be a different graph than the source.
type Mutable interface {
	// NewNode returns a node with a unique arbitrary ID.
	NewNode() Node

	// Adds a node to the graph. If this is called multiple times for the same ID, the newer node
	// overwrites the old one.
	AddNode(Node)

	// RemoveNode removes a node from the graph, as well as any edges
	// attached to it. If no such node exists, this is a no-op, not an error.
	RemoveNode(Node)
}

// Mutable is an interface ensuring the implementation of the ability to construct
// an arbitrary undirected graph. It is very important to note that any implementation
// of Mutable absolutely cannot safely implement the Directed interface.
//
// A Mutable is required to store any Edge argument in the same way Mutable must
// store a Node argument -- any retrieval call is required to return the exact supplied edge.
// This is what makes it incompatible with Directed.
//
// The reasoning is this: if you call AddUndirectedEdge(Edge{from,to}); you are required
// to return the exact edge passed in when a retrieval method (EdgeTo/EdgeBetween) is called.
// If I call EdgeTo(to,from), this means that since the edge exists, and was added as
// Edge{from,to} this function MUST return Edge{from,to}. However, EdgeTo requires this
// be returned as Edge{to,from}. Thus there's a conflict that cannot be resolved between the
// two interface requirements.
type MutableUndirected interface {
	CostGraph
	Mutable

	// Like EdgeBetween in Graph, AddUndirectedEdge adds an edge between two nodes.
	// If one or both nodes do not exist, the graph is expected to add them. However,
	// if the nodes already exist it should NOT replace existing nodes with e.From() or
	// e.To(). Overwriting nodes should explicitly be done with another call to AddNode()
	AddUndirectedEdge(e Edge, cost float64)

	// RemoveEdge clears the stored edge between two nodes. Calling this will never
	// remove a node. If the edge does not exist this is a no-op, not an error.
	RemoveUndirectedEdge(Edge)
}

// MutableDirected is an interface that ensures one can construct an arbitrary directed
// graph. Naturally, a MutableDirected works for both undirected and directed cases,
// but simply using a Mutable may be cleaner. As the documentation for Mutable
// notes, however, a graph cannot safely implement Mutable and MutableDirected
// at the same time, because of the functionality of a EdgeTo in Directed.
type MutableDirected interface {
	CostDirected
	Mutable

	// Like EdgeTo in Directed, AddDirectedEdge adds an edge with the given direction.
	// If one or both nodes do not exist, the graph is expected to add them. However,
	// if the nodes already exist it should NOT replace existing nodes with e.From() or
	// e.To(). Overwriting nodes should explicitly be done with another call to AddNode()
	AddDirectedEdge(e Edge, cost float64)

	// Removes an edge FROM e.From TO e.To. If no such edge exists, this is a no-op,
	// not an error.
	RemoveDirectedEdge(Edge)
}

// A function that returns the cost of following an edge
type CostFunc func(Edge) float64

func uniformCost(e Edge) float64 {
	if e == nil {
		return math.Inf(1)
	}
	return 1
}

// Estimates the cost of travelling between two nodes
type HeuristicCostFunc func(Node, Node) float64

// CopyUndirected copies the undirected graph src into dst; maintaining all
// node IDs without clearing items already in dst.
func CopyUndirected(dst MutableUndirected, src Undirected) {
	var cost CostFunc
	if src, ok := src.(Coster); ok {
		cost = src.Cost
	} else {
		cost = uniformCost
	}
	for _, node := range src.Nodes() {
		dst.AddNode(node)
		for _, succ := range src.From(node) {
			edge := src.EdgeBetween(node, succ)
			dst.AddUndirectedEdge(edge, cost(edge))
		}
	}
}

// CopyDirected copies the directed graph src into dst; maintaining all node
// IDs without clearing items already in dst.
func CopyDirected(dst MutableDirected, src Directed) {
	var cost CostFunc
	if src, ok := src.(Coster); ok {
		cost = src.Cost
	} else {
		cost = uniformCost
	}
	for _, node := range src.Nodes() {
		succs := src.From(node)
		dst.AddNode(node)
		for _, succ := range succs {
			edge := src.EdgeFromTo(node, succ)
			dst.AddDirectedEdge(edge, cost(edge))
		}
	}

}
