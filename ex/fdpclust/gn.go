package main

import (
	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/internal"
)

type GraphNode struct {
	id        int
	neighbors []graph.Node
	roots     []*GraphNode
}

func (g *GraphNode) Has(n graph.Node) bool {
	if n.ID() == g.id {
		return true
	}

	visited := internal.IntSet{g.id: struct{}{}}
	for _, root := range g.roots {
		if root.ID() == n.ID() {
			return true
		}

		if root.has(n, visited) {
			return true
		}
	}

	for _, neigh := range g.neighbors {
		if neigh.ID() == n.ID() {
			return true
		}

		if gn, ok := neigh.(*GraphNode); ok {
			if gn.has(n, visited) {
				return true
			}
		}
	}

	return false
}

func (g *GraphNode) has(n graph.Node, visited internal.IntSet) bool {
	for _, root := range g.roots {
		if visited.Has(root.ID()) {
			continue
		}

		visited.Add(root.ID())
		if root.ID() == n.ID() {
			return true
		}

		if root.has(n, visited) {
			return true
		}

	}

	for _, neigh := range g.neighbors {
		if visited.Has(neigh.ID()) {
			continue
		}

		visited.Add(neigh.ID())
		if neigh.ID() == n.ID() {
			return true
		}

		if gn, ok := neigh.(*GraphNode); ok {
			if gn.has(n, visited) {
				return true
			}
		}

	}

	return false
}

func (g *GraphNode) Order() int {
	n := 1
	visited := internal.IntSet{g.id: struct{}{}}

	for _, root := range g.roots {
		n++
		visited.Add(root.ID())
		n += root.order(visited)
	}

	for _, neigh := range g.neighbors {
		n++
		visited.Add(neigh.ID())
		if gn, ok := neigh.(*GraphNode); ok {
			n += gn.order(visited)
		}
	}

	return n
}

func (g *GraphNode) order(visited internal.IntSet) int {
	var n int
	for _, root := range g.roots {
		if visited.Has(root.ID()) {
			continue
		}
		visited.Add(root.ID())
		n++
		n += root.order(visited)
	}

	for _, neigh := range g.neighbors {
		if visited.Has(neigh.ID()) {
			continue
		}
		n++
		if gn, ok := neigh.(*GraphNode); ok {
			n += gn.order(visited)
		}
	}

	return n
}

func (g *GraphNode) Nodes() []graph.Node {
	toReturn := []graph.Node{g}
	visited := internal.IntSet{g.id: struct{}{}}

	for _, root := range g.roots {
		toReturn = append(toReturn, root)
		visited.Add(root.ID())

		toReturn = root.nodes(toReturn, visited)
	}

	for _, neigh := range g.neighbors {
		toReturn = append(toReturn, neigh)
		visited.Add(neigh.ID())

		if gn, ok := neigh.(*GraphNode); ok {
			toReturn = gn.nodes(toReturn, visited)
		}
	}

	return toReturn
}

func (g *GraphNode) nodes(list []graph.Node, visited internal.IntSet) []graph.Node {
	for _, root := range g.roots {
		if visited.Has(root.ID()) {
			continue
		}
		visited.Add(root.ID())
		list = append(list, graph.Node(root))

		list = root.nodes(list, visited)
	}

	for _, neigh := range g.neighbors {
		if visited.Has(neigh.ID()) {
			continue
		}

		list = append(list, neigh)
		if gn, ok := neigh.(*GraphNode); ok {
			list = gn.nodes(list, visited)
		}
	}

	return list
}

func (g *GraphNode) From(n graph.Node) []graph.Node {
	if n.ID() == g.ID() {
		return g.neighbors
	}

	visited := internal.IntSet{g.id: struct{}{}}
	for _, root := range g.roots {
		visited.Add(root.ID())

		if result := root.findNeighbors(n, visited); result != nil {
			return result
		}
	}

	for _, neigh := range g.neighbors {
		visited.Add(neigh.ID())

		if gn, ok := neigh.(*GraphNode); ok {
			if result := gn.findNeighbors(n, visited); result != nil {
				return result
			}
		}
	}

	return nil
}

func (g *GraphNode) findNeighbors(n graph.Node, visited internal.IntSet) []graph.Node {
	if n.ID() == g.ID() {
		return g.neighbors
	}

	for _, root := range g.roots {
		if visited.Has(root.ID()) {
			continue
		}
		visited.Add(root.ID())

		if result := root.findNeighbors(n, visited); result != nil {
			return result
		}
	}

	for _, neigh := range g.neighbors {
		if visited.Has(neigh.ID()) {
			continue
		}
		visited.Add(neigh.ID())

		if gn, ok := neigh.(*GraphNode); ok {
			if result := gn.findNeighbors(n, visited); result != nil {
				return result
			}
		}
	}

	return nil
}

func (g *GraphNode) HasEdge(n, neighbor graph.Node) bool {
	if n.ID() == g.id || neighbor.ID() == g.id {
		for _, neigh := range g.neighbors {
			if neigh.ID() == n.ID() || neigh.ID() == neighbor.ID() {
				return true
			}
		}

		return false
	}

	visited := internal.IntSet{g.id: struct{}{}}
	for _, root := range g.roots {
		visited.Add(root.ID())
		if root.edgeBetween(n, neighbor, visited) != nil {
			return true
		}
	}

	for _, neigh := range g.neighbors {
		visited.Add(neigh.ID())
		if gn, ok := neigh.(*GraphNode); ok {
			if gn.edgeBetween(n, neighbor, visited) != nil {
				return true
			}
		}
	}

	return false
}

func (g *GraphNode) EdgeBetween(n, neighbor graph.Node) graph.Edge {
	if n.ID() == g.id || neighbor.ID() == g.id {
		for _, neigh := range g.neighbors {
			if neigh.ID() == n.ID() || neigh.ID() == neighbor.ID() {
				return concrete.Edge{g, neigh}
			}
		}

		return nil
	}

	visited := internal.IntSet{g.id: struct{}{}}
	for _, root := range g.roots {
		visited.Add(root.ID())
		if result := root.edgeBetween(n, neighbor, visited); result != nil {
			return result
		}
	}

	for _, neigh := range g.neighbors {
		visited.Add(neigh.ID())
		if gn, ok := neigh.(*GraphNode); ok {
			if result := gn.edgeBetween(n, neighbor, visited); result != nil {
				return result
			}
		}
	}

	return nil
}

func (g *GraphNode) edgeBetween(n, neighbor graph.Node, visited internal.IntSet) graph.Edge {
	if n.ID() == g.id || neighbor.ID() == g.id {
		for _, neigh := range g.neighbors {
			if neigh.ID() == n.ID() || neigh.ID() == neighbor.ID() {
				return concrete.Edge{g, neigh}
			}
		}

		return nil
	}

	for _, root := range g.roots {
		if visited.Has(root.ID()) {
			continue
		}
		visited.Add(root.ID())
		if result := root.edgeBetween(n, neighbor, visited); result != nil {
			return result
		}
	}

	for _, neigh := range g.neighbors {
		if visited.Has(neigh.ID()) {
			continue
		}

		visited.Add(neigh.ID())
		if gn, ok := neigh.(*GraphNode); ok {
			if result := gn.edgeBetween(n, neighbor, visited); result != nil {
				return result
			}
		}
	}

	return nil
}

func (g *GraphNode) ID() int {
	return g.id
}

func (g *GraphNode) AddNeighbor(n *GraphNode) {
	g.neighbors = append(g.neighbors, graph.Node(n))
}

func (g *GraphNode) AddRoot(n *GraphNode) {
	g.roots = append(g.roots, n)
}

func NewGraphNode(id int) *GraphNode {
	return &GraphNode{id: id, neighbors: make([]graph.Node, 0), roots: make([]*GraphNode, 0)}
}
