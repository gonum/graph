// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	"container/heap"
	"math"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
)

var inf = math.Inf(1)

type searchFuncs struct {
	from, to                               func(graph.Node) []graph.Node
	isSuccessor, isPredecessor, isNeighbor func(graph.Node, graph.Node) bool
	weight                                 graph.WeightFunc
	heuristic                              graph.HeuristicWeightFunc
	edgeTo, edgeBetween                    func(graph.Node, graph.Node) graph.Edge
}

func genIsSuccessor(g graph.Directed) func(graph.Node, graph.Node) bool {
	return func(node, succ graph.Node) bool {
		return g.EdgeFromTo(node, succ) != nil
	}
}

func genIsPredecessor(g graph.Directed) func(graph.Node, graph.Node) bool {
	return func(node, succ graph.Node) bool {
		return g.EdgeFromTo(succ, node) != nil
	}
}

func genEdgeFor(g graph.Graph) func(u, v graph.Node) graph.Edge {
	return func(u, v graph.Node) graph.Edge {
		if g.HasEdge(u, v) {
			return edge{u, v}
		}
		return nil
	}
}

type edge struct {
	u, v graph.Node
}

func (e edge) From() graph.Node { return e.u }
func (e edge) To() graph.Node   { return e.v }

// Sets up the weight functions and successor functions so I don't have to do a type switch every
// time. This almost always does more work than is necessary, but since it's only executed once
// per function, and graph functions are rather costly, the "extra work" should be negligible.
func setupFuncs(g graph.Graph, weight graph.WeightFunc, heuristic graph.HeuristicWeightFunc) searchFuncs {
	var sf searchFuncs

	switch g := g.(type) {
	case graph.Directed:
		sf.from = g.From
		sf.to = g.To
		sf.isSuccessor = genIsSuccessor(g)
		sf.isPredecessor = genIsPredecessor(g)
		sf.isNeighbor = g.HasEdge
		sf.edgeBetween = g.EdgeFromTo
		sf.edgeTo = g.EdgeFromTo
	case graph.Undirected:
		sf.from = g.From
		sf.to = g.From
		isNeighbor := g.HasEdge
		sf.isSuccessor = isNeighbor
		sf.isPredecessor = isNeighbor
		sf.isNeighbor = isNeighbor
		sf.edgeBetween = g.EdgeBetween
		sf.edgeTo = g.EdgeBetween
	default:
		sf.from = g.From
		sf.to = g.From
		isNeighbor := g.HasEdge
		sf.isSuccessor = isNeighbor
		sf.isPredecessor = isNeighbor
		sf.isNeighbor = isNeighbor
		sf.edgeBetween = genEdgeFor(g)
		sf.edgeTo = genEdgeFor(g)
	}

	if heuristic != nil {
		sf.heuristic = heuristic
	} else {
		if g, ok := g.(graph.HeuristicWeighter); ok {
			sf.heuristic = g.HeuristicWeight
		} else {
			sf.heuristic = NullHeuristic
		}
	}

	if weight != nil {
		sf.weight = weight
	} else {
		if g, ok := g.(graph.Weighter); ok {
			sf.weight = g.Weight
		} else {
			sf.weight = graph.UniformCostWeight
		}
	}

	return sf
}

/** Sorts a list of edges by weight, agnostic to repeated edges as well as direction **/

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

/** Keeps track of a node's scores so they can be used in a priority queue for A* **/

type internalNode struct {
	graph.Node
	gscore, fscore float64
}

/* A* stuff */
type aStarPriorityQueue struct {
	indexList map[int]int
	nodes     []internalNode
}

func (pq *aStarPriorityQueue) Less(i, j int) bool {
	// As the heap documentation says, a priority queue is listed if the actual values
	// are treated as if they were negative
	return pq.nodes[i].fscore < pq.nodes[j].fscore
}

func (pq *aStarPriorityQueue) Swap(i, j int) {
	pq.indexList[pq.nodes[i].ID()] = j
	pq.indexList[pq.nodes[j].ID()] = i

	pq.nodes[i], pq.nodes[j] = pq.nodes[j], pq.nodes[i]
}

func (pq *aStarPriorityQueue) Len() int {
	return len(pq.nodes)
}

func (pq *aStarPriorityQueue) Push(x interface{}) {
	node := x.(internalNode)
	pq.nodes = append(pq.nodes, node)
	pq.indexList[node.ID()] = len(pq.nodes) - 1
}

func (pq *aStarPriorityQueue) Pop() interface{} {
	x := pq.nodes[len(pq.nodes)-1]
	pq.nodes = pq.nodes[:len(pq.nodes)-1]
	delete(pq.indexList, x.ID())

	return x
}

func (pq *aStarPriorityQueue) Fix(id int, newGScore, newFScore float64) {
	if i, ok := pq.indexList[id]; ok {
		pq.nodes[i].gscore = newGScore
		pq.nodes[i].fscore = newFScore
		heap.Fix(pq, i)
	}
}

func (pq *aStarPriorityQueue) Find(id int) (internalNode, bool) {
	loc, ok := pq.indexList[id]
	if ok {
		return pq.nodes[loc], true
	} else {
		return internalNode{}, false
	}

}

func (pq *aStarPriorityQueue) Exists(id int) bool {
	_, ok := pq.indexList[id]
	return ok
}

// General utility funcs

// Rebuilds a path backwards from the goal.
func rebuildPath(predecessors map[int]graph.Node, goal graph.Node) []graph.Node {
	if n, ok := goal.(internalNode); ok {
		goal = n.Node
	}
	path := []graph.Node{goal}
	curr := goal
	for prev, ok := predecessors[curr.ID()]; ok; prev, ok = predecessors[curr.ID()] {
		if n, ok := prev.(internalNode); ok {
			prev = n.Node
		}
		path = append(path, prev)
		curr = prev
	}

	// Reverse the path since it was built backwards
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
