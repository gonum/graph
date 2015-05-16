// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

import (
	"container/heap"
	"errors"
	"math"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/internal"
)

// Returns an ordered list consisting of the nodes between start and goal. The path will be the
// shortest path assuming the function heuristic is admissible. The second return value is the
// weight, and the third is the number of nodes expanded while searching (useful info for tuning
// heuristics). Negative Costs will cause bad things to happen, as well as negative heuristic
// estimates.
//
// A heuristic is admissible if, for any node in the graph, the heuristic estimate of the weight
// between the node and the goal is less than or set to the true weight.
//
// Performance may be improved by providing a consistent heuristic (though one is not needed to
// find the optimal path), a heuristic is consistent if its value for a given node is less than
// (or equal to) the actual weight of reaching its neighbors + the heuristic estimate for the
// neighbor itself. You can force consistency by making your HeuristicCost function return
// max(NonConsistentHeuristicCost(neighbor,goal), NonConsistentHeuristicCost(self,goal) -
// Cost(self,neighbor)). If there are multiple neighbors, take the max of all of them.
//
// Cost and HeuristicCost take precedence for evaluating weight/heuristic distance. If one is not
// present (i.e. nil) the function will check the graph's interface for the respective interface:
// Coster for Cost and HeuristicCoster for HeuristicCost. If the correct one is present, it will
// use the graph's function for evaluation.
//
// Finally, if neither the argument nor the interface is present, the function will assume
// UniformCost for Cost and NullHeuristic for HeuristicCost.
//
// To run Uniform Cost Search, run A* with the NullHeuristic.
//
// To run Breadth First Search, run A* with both the NullHeuristic and UniformCost (or any weight
// function that returns a uniform positive value.)
func AStar(start, goal graph.Node, g graph.Graph, weight graph.WeightFunc, heuristic graph.HeuristicWeightFunc) (path []graph.Node, pathCost float64, nodesExpanded int) {
	sf := setupFuncs(g, weight, heuristic)
	from, weight, heuristic, edgeTo := sf.from, sf.weight, sf.heuristic, sf.edgeTo

	closedSet := make(map[int]internalNode)
	openSet := &aStarPriorityQueue{nodes: make([]internalNode, 0), indexList: make(map[int]int)}
	heap.Init(openSet)
	node := internalNode{start, 0, heuristic(start, goal)}
	heap.Push(openSet, node)
	predecessor := make(map[int]graph.Node)

	for openSet.Len() != 0 {
		curr := heap.Pop(openSet).(internalNode)

		nodesExpanded += 1

		if curr.ID() == goal.ID() {
			return rebuildPath(predecessor, goal), curr.gscore, nodesExpanded
		}

		closedSet[curr.ID()] = curr

		for _, neighbor := range from(curr.Node) {
			if _, ok := closedSet[neighbor.ID()]; ok {
				continue
			}

			g := curr.gscore + weight(edgeTo(curr.Node, neighbor))

			if existing, exists := openSet.Find(neighbor.ID()); !exists {
				predecessor[neighbor.ID()] = curr
				node = internalNode{neighbor, g, g + heuristic(neighbor, goal)}
				heap.Push(openSet, node)
			} else if g < existing.gscore {
				predecessor[neighbor.ID()] = curr
				openSet.Fix(neighbor.ID(), g, g+heuristic(neighbor, goal))
			}
		}
	}

	return nil, 0, nodesExpanded
}

// BreadthFirstSearch finds a path with a minimal number of edges from from start to goal.
//
// BreadthFirstSearch returns the path found and the number of nodes visited in the search.
// The returned path is nil if no path exists.
func BreadthFirstSearch(start, goal graph.Node, g graph.Graph) ([]graph.Node, int) {
	path, _, visited := AStar(start, goal, g, graph.UniformCostWeight, NullHeuristic)
	return path, visited
}

// Dijkstra's Algorithm is essentially a goalless Uniform Cost Search. That is, its results are
// roughly equivalent to running A* with the Null Heuristic from a single node to every other node
// in the graph -- though it's a fair bit faster because running A* in that way will recompute
// things it's already computed every call. Note that you won't necessarily get the same path
// you would get for A*, but the weight is guaranteed to be the same (that is, if multiple shortest
// paths exist, you may get a different shortest path).
//
// Like A*, Dijkstra's Algorithm likely won't run correctly with negative edge weights -- use
// Bellman-Ford for that instead.
//
// Dijkstra's algorithm usually only returns a weight map, however, since the data is available
// this version will also reconstruct the path to every node.
func Dijkstra(source graph.Node, g graph.Graph, weight graph.WeightFunc) (paths map[int][]graph.Node, costs map[int]float64) {

	sf := setupFuncs(g, weight, nil)
	from, weight, edgeTo := sf.from, sf.weight, sf.edgeTo

	nodes := g.Nodes()
	openSet := &aStarPriorityQueue{nodes: make([]internalNode, 0), indexList: make(map[int]int)}
	costs = make(map[int]float64, len(nodes)) // May overallocate, will change if it becomes a problem
	predecessor := make(map[int]graph.Node, len(nodes))
	nodeIDMap := make(map[int]graph.Node, len(nodes))
	heap.Init(openSet)

	costs[source.ID()] = 0
	heap.Push(openSet, internalNode{source, 0, 0})

	for openSet.Len() != 0 {
		node := heap.Pop(openSet).(internalNode)

		nodeIDMap[node.ID()] = node

		for _, neighbor := range from(node) {
			tmpCost := costs[node.ID()] + weight(edgeTo(node, neighbor))
			if weight, ok := costs[neighbor.ID()]; !ok {
				costs[neighbor.ID()] = tmpCost
				predecessor[neighbor.ID()] = node
				heap.Push(openSet, internalNode{neighbor, tmpCost, tmpCost})
			} else if tmpCost < weight {
				costs[neighbor.ID()] = tmpCost
				predecessor[neighbor.ID()] = node
				openSet.Fix(neighbor.ID(), tmpCost, tmpCost)
			}
		}
	}

	paths = make(map[int][]graph.Node, len(costs))
	for node := range costs { // Only reconstruct the path if one exists
		paths[node] = rebuildPath(predecessor, nodeIDMap[node])
	}
	return paths, costs
}

// The Bellman-Ford Algorithm is the same as Dijkstra's Algorithm with a key difference. They both
// take a single source and find the shortest path to every other (reachable) node in the graph.
// Bellman-Ford, however, will detect negative edge loops and abort if one is present. A negative
// edge loop occurs when there is a cycle in the graph such that it can take an edge with a
// negative weight over and over. A -(-2)> B -(2)> C isn't a loop because A->B can only be taken once,
// but A<-(-2)->B-(2)>C is one because A and B have a bi-directional edge, and algorithms like
// Dijkstra's will infinitely flail between them getting progressively lower costs.
//
// That said, if you do not have a negative edge weight, use Dijkstra's Algorithm instead, because
// it's faster.
//
// Like Dijkstra's, along with the costs this implementation will also construct all the paths for
// you. In addition, it has a third return value which will be true if the algorithm was aborted
// due to the presence of a negative edge weight cycle.
func BellmanFord(source graph.Node, g graph.Graph, weight graph.WeightFunc) (paths map[int][]graph.Node, costs map[int]float64, err error) {
	sf := setupFuncs(g, weight, nil)
	from, weight, edgeTo := sf.from, sf.weight, sf.edgeTo

	predecessor := make(map[int]graph.Node)
	costs = make(map[int]float64)
	nodeIDMap := make(map[int]graph.Node)
	nodeIDMap[source.ID()] = source
	costs[source.ID()] = 0
	nodes := g.Nodes()

	for i := 1; i < len(nodes)-1; i++ {
		for _, node := range nodes {
			nodeIDMap[node.ID()] = node
			succs := from(node)
			for _, succ := range succs {
				weight := weight(edgeTo(node, succ))
				nodeIDMap[succ.ID()] = succ

				if dist := costs[node.ID()] + weight; dist < costs[succ.ID()] {
					costs[succ.ID()] = dist
					predecessor[succ.ID()] = node
				}
			}

		}
	}

	for _, node := range nodes {
		for _, succ := range from(node) {
			weight := weight(edgeTo(node, succ))
			if costs[node.ID()]+weight < costs[succ.ID()] {
				return nil, nil, errors.New("Negative edge cycle detected")
			}
		}
	}

	paths = make(map[int][]graph.Node, len(costs))
	for node := range costs {
		paths[node] = rebuildPath(predecessor, nodeIDMap[node])
	}
	return paths, costs, nil
}

// Johnson's Algorithm generates the lowest weight path between every pair of nodes in the graph.
//
// It makes use of Bellman-Ford and a dummy graph. It creates a dummy node containing edges with a
// weight of zero to every other node. Then it runs Bellman-Ford with this dummy node as the source.
// It then modifies the all the nodes' edge weights (which gets rid of all negative weights).
//
// Finally, it removes the dummy node and runs Dijkstra's starting at every node.
//
// This algorithm is fairly slow. Its purpose is to remove negative edge weights to allow
// Dijkstra's to function properly. It's probably not worth it to run this algorithm if you have
// all non-negative edge weights. Also note that this implementation copies your whole graph into
// a GonumGraph (so it can add/remove the dummy node and edges and reweight the graph).
//
// Its return values are, in order: a map from the source node, to the destination node, to the
// path between them; a map from the source node, to the destination node, to the weight of the path
// between them; and a bool that is true if Bellman-Ford detected a negative edge weight cycle --
// thus causing it (and this algorithm) to abort (if aborted is true, both maps will be nil).
func Johnson(g graph.Graph, weight graph.WeightFunc) (nodePaths map[int]map[int][]graph.Node, nodeCosts map[int]map[int]float64, err error) {
	sf := setupFuncs(g, weight, nil)
	from, weight, edgeTo := sf.from, sf.weight, sf.edgeTo

	/* Copy graph into a mutable one since it has to be altered for this algorithm */
	dummyGraph := concrete.NewDirectedGraph(math.Inf(1))
	for _, node := range g.Nodes() {
		neighbors := from(node)
		dummyGraph.Has(node)
		dummyGraph.AddNode(node)
		for _, neighbor := range neighbors {
			e := edgeTo(node, neighbor)
			c := weight(e)
			// Make a new edge with from and to swapped;
			// works due to the fact that we're not returning
			// any edges in this so the contract doesn't need
			// to be fulfilled.
			if e.From().ID() != node.ID() {
				e = concrete.Edge{e.To(), e.From()}
			}

			dummyGraph.AddDirectedEdge(e, c)
		}
	}

	/* Step 1: Dummy node with 0 weight edge weights to every other node*/
	dummyNode := dummyGraph.NewNode()
	dummyGraph.AddNode(dummyNode)
	for _, node := range g.Nodes() {
		dummyGraph.AddDirectedEdge(concrete.Edge{dummyNode, node}, 0)
	}

	/* Step 2: Run Bellman-Ford starting at the dummy node, abort if it detects a cycle */
	_, costs, err := BellmanFord(dummyNode, dummyGraph, nil)
	if err != nil {
		return nil, nil, err
	}

	/* Step 3: reweight the graph and remove the dummy node */
	for _, node := range g.Nodes() {
		for _, succ := range from(node) {
			e := edgeTo(node, succ)
			dummyGraph.AddDirectedEdge(e, weight(e)+costs[node.ID()]-costs[succ.ID()])
		}
	}

	dummyGraph.RemoveNode(dummyNode)

	/* Step 4: Run Dijkstra's starting at every node */
	nodePaths = make(map[int]map[int][]graph.Node, len(g.Nodes()))
	nodeCosts = make(map[int]map[int]float64)

	for _, node := range g.Nodes() {
		nodePaths[node.ID()], nodeCosts[node.ID()] = Dijkstra(node, dummyGraph, nil)
	}

	return nodePaths, nodeCosts, nil
}

// Expands the first node it sees trying to find the destination. Depth First Search is *not*
// guaranteed to find the shortest path, however, if a path exists DFS is guaranteed to find it
// (provided you don't find a way to implement a Graph with an infinite depth.)
func DepthFirstSearch(start, goal graph.Node, g graph.Graph) []graph.Node {
	sf := setupFuncs(g, nil, nil)
	from := sf.from

	closedSet := make(internal.IntSet)
	predecessor := make(map[int]graph.Node)

	openSet := internal.NodeStack{start}
	for openSet.Len() != 0 {
		curr := openSet.Pop()

		if closedSet.Has(curr.ID()) {
			continue
		}

		if curr.ID() == goal.ID() {
			return rebuildPath(predecessor, goal)
		}

		closedSet.Add(curr.ID())

		for _, neighbor := range from(curr) {
			if closedSet.Has(neighbor.ID()) {
				continue
			}

			predecessor[neighbor.ID()] = curr
			openSet.Push(neighbor)
		}
	}

	return nil
}

// IsPath returns true for a connected path within a graph.
//
// IsPath returns true if, starting at path[0] and ending at path[len(path)-1], all nodes between
// are valid neighbors. That is, for each element path[i], path[i+1] is a valid successor.
//
// As special cases, IsPath returns true for a nil or zero length path, and for a path of length 1
// (only one node) but only if the node listed in path exists within the graph.
//
// Graph must be non-nil.
func IsPath(path []graph.Node, g graph.Graph) bool {
	var canReach func(u, v graph.Node) bool
	switch g := g.(type) {
	case graph.Directed:
		canReach = func(u, v graph.Node) bool {
			return g.EdgeFromTo(u, v) != nil
		}
	default:
		canReach = g.HasEdge
	}

	if path == nil || len(path) == 0 {
		return true
	} else if len(path) == 1 {
		return g.Has(path[0])
	}

	for i := 0; i < len(path)-1; i++ {
		if !canReach(path[i], path[i+1]) {
			return false
		}
	}

	return true
}

// An admissible, consistent heuristic that won't speed up computation time at all.
func NullHeuristic(_, _ graph.Node) float64 {
	return 0
}

/* Control flow graph stuff */

// A dominates B if and only if the only path through B travels through A.
//
// This returns all possible dominators for all nodes, it does not prune for strict dominators,
// immediate dominators etc.
//
func Dominators(start graph.Node, g graph.Graph) map[int]internal.Set {
	allNodes := make(internal.Set)
	nlist := g.Nodes()
	dominators := make(map[int]internal.Set, len(nlist))
	for _, node := range nlist {
		allNodes.Add(node)
	}

	to := setupFuncs(g, nil, nil).to

	for _, node := range nlist {
		dominators[node.ID()] = make(internal.Set)
		if node.ID() == start.ID() {
			dominators[node.ID()].Add(start)
		} else {
			dominators[node.ID()].Copy(allNodes)
		}
	}

	for somethingChanged := true; somethingChanged; {
		somethingChanged = false
		for _, node := range nlist {
			if node.ID() == start.ID() {
				continue
			}
			preds := to(node)
			if len(preds) == 0 {
				continue
			}
			tmp := make(internal.Set).Copy(dominators[preds[0].ID()])
			for _, pred := range preds[1:] {
				tmp.Intersect(tmp, dominators[pred.ID()])
			}

			dom := make(internal.Set)
			dom.Add(node)

			dom.Union(dom, tmp)
			if !internal.Equal(dom, dominators[node.ID()]) {
				dominators[node.ID()] = dom
				somethingChanged = true
			}
		}
	}

	return dominators
}

// A Postdominates B if and only if all paths from B travel through A.
//
// This returns all possible post-dominators for all nodes, it does not prune for strict
// postdominators, immediate postdominators etc.
func PostDominators(end graph.Node, g graph.Graph) map[int]internal.Set {
	from := setupFuncs(g, nil, nil).from

	allNodes := make(internal.Set)
	nlist := g.Nodes()
	dominators := make(map[int]internal.Set, len(nlist))
	for _, node := range nlist {
		allNodes.Add(node)
	}

	for _, node := range nlist {
		dominators[node.ID()] = make(internal.Set)
		if node.ID() == end.ID() {
			dominators[node.ID()].Add(end)
		} else {
			dominators[node.ID()].Copy(allNodes)
		}
	}

	for somethingChanged := true; somethingChanged; {
		somethingChanged = false
		for _, node := range nlist {
			if node.ID() == end.ID() {
				continue
			}
			succs := from(node)
			if len(succs) == 0 {
				continue
			}
			tmp := make(internal.Set).Copy(dominators[succs[0].ID()])
			for _, succ := range succs[1:] {
				tmp.Intersect(tmp, dominators[succ.ID()])
			}

			dom := make(internal.Set)
			dom.Add(node)

			dom.Union(dom, tmp)
			if !internal.Equal(dom, dominators[node.ID()]) {
				dominators[node.ID()] = dom
				somethingChanged = true
			}
		}
	}

	return dominators
}
