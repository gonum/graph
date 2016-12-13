// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import "github.com/gonum/graph"

// A disjoint set is a collection of non-overlapping sets. That is, for any two sets in the
// disjoint set, their intersection is the empty set.
//
// A disjoint set has three principle operations: Make Set, Find, and Union.
//
// Make set creates a new set for an element (presuming it does not already exist in any set in
// the disjoint set), Find finds the set containing that element (if any), and Union merges two
// sets in the disjoint set. In general, algorithms operating on disjoint sets are "union-find"
// algorithms, where two sets are found with Find, and then joined with Union.
//
// A concrete example of a union-find algorithm can be found as discrete.Kruskal -- which unions
// two sets when an edge is created between two vertices, and refuses to make an edge between two
// vertices if they're part of the same set.
type disjointSet struct {
	master map[int]*disjointSetNode
}

type disjointSetNode struct {
	parent *disjointSetNode
	rank   int
	node   graph.Node
}

func newDisjointSet() *disjointSet {
	return &disjointSet{master: make(map[int]*disjointSetNode)}
}

// If the element isn't already somewhere in there, adds it to the master set and its own tiny set.
func (ds *disjointSet) makeSet(n graph.Node) {
	if _, ok := ds.master[n.ID()]; ok {
		return
	}
	dsNode := &disjointSetNode{rank: 0, node: n}
	dsNode.parent = dsNode
	ds.master[n.ID()] = dsNode
}

// Returns the set the element belongs to, or nil if none.
func (ds *disjointSet) find(n graph.Node) *disjointSetNode {
	dsNode, ok := ds.master[n.ID()]
	if !ok {
		return nil
	}
	// Path compression.
	p := find(dsNode)
	dsNode.parent = p
	return p
}

func find(dsNode *disjointSetNode) *disjointSetNode {
	if dsNode.parent != dsNode {
		dsNode.parent = find(dsNode.parent)
	}

	return dsNode.parent
}

// Unilons two subsets within the disjointSet.
//
// If x or y are not in this disjoint set, the behavior is undefined. If either pointer is nil,
// this function will panic.
func (ds *disjointSet) union(x, y *disjointSetNode) {
	if x == nil || y == nil {
		panic("Disjoint Set union on nil sets")
	}
	xRoot := find(x)
	yRoot := find(y)
	if xRoot == nil || yRoot == nil {
		return
	}

	if xRoot == yRoot {
		return
	}

	if xRoot.rank < yRoot.rank {
		xRoot.parent = yRoot
	} else if yRoot.rank < xRoot.rank {
		yRoot.parent = xRoot
	} else {
		yRoot.parent = xRoot
		xRoot.rank++
	}
}

// connectedComponents: Get each connected component of ds as slice of nodes.
func (ds *disjointSet) connectedComponents() [][]graph.Node {
	cc := make(map[int][]graph.Node)
	for _, x := range ds.master {
		parent := find(x)
		id := parent.node.ID()
		if _, ok := cc[id]; !ok {
			var c []graph.Node
			c = append(c, x.node)
			cc[id] = c
		} else {
			cc[id] = append(cc[id], x.node)
		}
	}
	// Convert map to slice.
	out := make([][]graph.Node, len(cc))
	index := 0
	for _, c := range cc {
		out[index] = c
		index++
	}
	return out
}
