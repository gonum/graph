// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dynamic

import (
	"container/heap"
	"fmt"
	"math"

	"github.com/gonum/graph"
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/path"
)

// DStarLite implements the D* Lite dynamic re-planning path search algorithm.
//
//  doi:10.1109/tro.2004.838026 and ISBN:0-262-51129-0 pp476-483
//
type DStarLite struct {
	s, t *dStarLiteNode
	last *dStarLiteNode

	world MutableDirected
	queue dStarLiteQueue
	k_m   float64

	weight    graph.WeightFunc
	heuristic path.Heuristic
}

// MutableDirected is a mutable directed graph that can return nodes identified
// by id number.
type MutableDirected interface {
	graph.MutableDirected
	Node(id int) graph.Node
}

// NewDStarLite returns a new DStarLite planner for the path from s to t in g using the
// heuristic h. The mutable directed graph world is used to store shortest path information
// during path planning.
//
// If h is nil, the DStarLite will use the g.HeuristicCost method if g implements
// path.HeuristicCoster, falling back to NullHeuristic otherwise. If the graph does not
// implement graph.Weighter, graph.UniformCost is used. NewDStarLite will panic if g has
// a negative edge weight.
func NewDStarLite(s, t graph.Node, g graph.Graph, h path.Heuristic, world MutableDirected) *DStarLite {
	/*
	   procedure Initialize()
	   {02”} U = ∅;
	   {03”} k_m = 0;
	   {04”} for all s ∈ S rhs(s) = g(s) = ∞;
	   {05”} rhs(s_goal) = 0;
	   {06”} U.Insert(s_goal, [h(s_start, s_goal); 0]);
	*/

	d := &DStarLite{
		s: &dStarLiteNode{Node: s, rhs: math.Inf(1), g: math.Inf(1)},
		t: &dStarLiteNode{Node: t, g: math.Inf(1)},

		world: world,

		heuristic: h,
	}

	/*
		procedure Main()
		{29”} s_last = s_start;
		{30”} Initialize();
	*/
	d.last = d.s

	if g, ok := g.(graph.Weighter); ok {
		d.weight = g.Weight
	} else {
		d.weight = graph.UniformCost
	}
	if d.heuristic == nil {
		if g, ok := g.(path.HeuristicCoster); ok {
			d.heuristic = g.HeuristicCost
		} else {
			d.heuristic = path.NullHeuristic
		}
	}

	d.queue.indexOf = make(map[int]int)
	d.queue.insert(d.t, key{h(s, t), 0})

	for _, n := range g.Nodes() {
		switch n.ID() {
		case d.s.ID():
			world.AddNode(d.s)
		case d.t.ID():
			world.AddNode(d.t)
		default:
			world.AddNode(&dStarLiteNode{Node: n, rhs: math.Inf(1), g: math.Inf(1)})
		}
	}
	for _, u := range world.Nodes() {
		for _, v := range g.From(u) {
			w := d.weight(g.Edge(u, v))
			if w < 0 {
				panic("D* Lite: negative edge weight")
			}
			world.SetEdge(concrete.Edge{F: u, T: world.Node(v.ID())}, w)
		}
	}

	/*
		procedure Main()
		{31”} ComputeShortestPath();
	*/
	d.findShortestPath()

	return d
}

// keyFor is the CalculateKey procedure in the D* Lite paper.
func (d *DStarLite) keyFor(s *dStarLiteNode) key {
	/*
	   procedure CalculateKey(s)
	   {01”} return [min(g(s), rhs(s)) + h(s_start, s) + k_m; min(g(s), rhs(s))];
	*/
	k := key{1: math.Min(s.g, s.rhs)}
	k[0] = k[1] + d.heuristic(d.s.Node, s.Node) + d.k_m
	return k
}

// update is the UpdateVertex procedure in the D* Lite papers.
func (d *DStarLite) update(u *dStarLiteNode) {
	/*
	   procedure UpdateVertex(u)
	   {07”} if (g(u) != rhs(u) AND u ∈ U) U.Update(u,CalculateKey(u));
	   {08”} else if (g(u) != rhs(u) AND u /∈ U) U.Insert(u,CalculateKey(u));
	   {09”} else if (g(u) = rhs(u) AND u ∈ U) U.Remove(u);
	*/
	uid := u.ID()
	inQueue := d.queue.has(uid)
	switch {
	case inQueue && u.g != u.rhs:
		d.queue.update(uid, d.keyFor(u))
	case !inQueue && u.g != u.rhs:
		d.queue.insert(u, d.keyFor(u))
	case inQueue && u.g == u.rhs:
		d.queue.remove(uid)
	}
}

// findShortestPath is the ComputeShortestPath procedure in the D* Lite papers.
func (d *DStarLite) findShortestPath() {
	/*
	   procedure ComputeShortestPath()
	   {10”} while (U.TopKey() < CalculateKey(s_start) OR rhs(s_start) > g(s_start))
	   {11”} u = U.Top();
	   {12”} k_old = U.TopKey();
	   {13”} k_new = CalculateKey(u);
	   {14”} if(k_old < k_new)
	   {15”}   U.Update(u, k_new);
	   {16”} else if (g(u) > rhs(u))
	   {17”}   g(u) = rhs(u);
	   {18”}   U.Remove(u);
	   {19”}   for all s ∈ Pred(u)
	   {20”}     if (s != s_goal) rhs(s) = min(rhs(s), c(s, u) + g(u));
	   {21”}     UpdateVertex(s);
	   {22”} else
	   {23”}   g_old = g(u);
	   {24”}   g(u) = ∞;
	   {25”}   for all s ∈ Pred(u) ∪ {u}
	   {26”}     if (rhs(s) = c(s, u) + g_old)
	   {27”}       if (s != s_goal) rhs(s) = min s'∈Succ(s)(c(s, s') + g(s'));
	   {28”}     UpdateVertex(s);
	*/
	for d.queue.Len() != 0 { // We use d.queue.Len since d.queue does not return an infinite key when empty.
		u := d.queue.top()
		if !u.key.less(d.keyFor(d.s)) && d.s.rhs <= d.s.g {
			break
		}
		switch kNew := d.keyFor(u); {
		case u.key.less(kNew):
			d.queue.update(u.ID(), kNew)
		case u.g > u.rhs:
			u.g = u.rhs
			d.queue.remove(u.ID())
			for _, _s := range d.world.To(u) {
				s := _s.(*dStarLiteNode)
				if s.ID() != d.t.ID() {
					s.rhs = math.Min(s.rhs, d.weight(d.world.Edge(s, u))+u.g)
				}
				d.update(s)
			}
		default:
			gOld := u.g
			u.g = math.Inf(1)
			for _, _s := range append(d.world.To(u), u) {
				s := _s.(*dStarLiteNode)
				if s.rhs == d.weight(d.world.Edge(s, u))+gOld {
					if s.ID() != d.t.ID() {
						s.rhs = math.Inf(1)
						for _, t := range d.world.From(s) {
							s.rhs = math.Min(s.rhs, d.weight(d.world.Edge(s, t))+t.(dStarLiteNode).g)
						}
					}
				}
				d.update(s)
			}
		}
	}
}

// Step performs one movement step along the best path towards the goal.
// It returns false if no further progression toward the goal can be
// achieved, either because the goal has been reached or because there
// is no path.
func (d *DStarLite) Step() bool {
	/*
	   procedure Main()
	   {32”} while (s_start != s_goal)
	   {33”} // if (g(rhs_start) = ∞) then there is no known path
	   {34”}   s_start = argmin s'∈Succ(s_start)(c(s_start, s') + g(s'));
	   {35”}   Move to s_start;
	*/
	if d.s.ID() == d.t.ID() {
		return false
	}
	if math.IsInf(d.s.rhs, 1) {
		return false
	}
	min := math.Inf(1)
	var next *dStarLiteNode
	for _, _s := range d.world.From(d.s) {
		s := _s.(*dStarLiteNode)
		w := d.weight(d.world.Edge(d.s, s)) + s.g
		if w < min {
			next = s
			min = w
		}
	}
	d.MoveTo(next)

	return true
}

// MoveTo moves to n in the world graph.
func (d *DStarLite) MoveTo(n graph.Node) {
	d.last = d.s
	d.s = d.world.Node(n.ID()).(*dStarLiteNode)
	d.k_m += d.heuristic(d.last, d.s)
}

// UpdateWorld updates or adds edges in the world graph. UpdateWorld will
// panic if changes includes a negative edge weight.
func (d *DStarLite) UpdateWorld(changes []concrete.WeightedEdge) {
	/*
	   procedure Main()
	   {36”}   Scan graph for changed edge costs;
	   {37”}   if any edge costs changed
	   {38”}     k_m = k_m + h(s_last, s_start);
	   {39”}     s_last = s_start;
	   {40”}     for all directed edges (u, v) with changed edge costs
	   {41”}       c_old = c(u, v);
	   {42”}       Update the edge cost c(u, v);
	   {43”}       if (c_old > c(u, v))
	   {44”}         if (u != s_goal) rhs(u) = min(rhs(u), c(u, v) + g(v));
	   {45”}       else if (rhs(u) = c_old + g(v))
	   {46”}         if (u != s_goal) rhs(u) = min s'∈Succ(u)(c(u, s') + g(s'));
	   {47”}       UpdateVertex(u);
	   {48”}     ComputeShortestPath()
	*/
	if len(changes) == 0 {
		return
	}
	d.k_m += d.heuristic(d.last, d.s)
	d.last = d.s
	for _, e := range changes {
		if e.Cost < 0 {
			panic("D* Lite: negative edge weight")
		}
		cOld := d.weight(e.Edge)
		u := d.worldNodeFor(e.From())
		v := d.worldNodeFor(e.To())
		d.world.SetEdge(concrete.Edge{F: u, T: v}, e.Cost)
		if cOld > e.Cost {
			if u.ID() != d.t.ID() {
				u.rhs = math.Min(u.rhs, e.Cost+v.g)
			}
		} else if u.rhs == cOld+v.g {
			if u.ID() != d.t.ID() {
				u.rhs = math.Inf(1)
				for _, t := range d.world.From(u) {
					u.rhs = math.Min(u.rhs, d.weight(d.world.Edge(u, t))+t.(dStarLiteNode).g)
				}
			}
		}
		d.update(u)
	}
	d.findShortestPath()
}

func (d *DStarLite) worldNodeFor(n graph.Node) *dStarLiteNode {
	switch w := d.world.Node(n.ID()).(type) {
	case *dStarLiteNode:
		return w
	case graph.Node:
		panic(fmt.Sprintf("D* Lite: illegal world node type: %T", w))
	default:
		return &dStarLiteNode{Node: n, rhs: math.Inf(1), g: math.Inf(1)}
	}
}

// Here returns the current location.
func (d *DStarLite) Here() graph.Node {
	return d.s.Node
}

// Path returns the path from the current location to the goal and the
// weight of the path,
func (d *DStarLite) Path() (p []graph.Node, weight float64) {
	u := d.s
	p = []graph.Node{u.Node}
	for u.ID() != d.t.ID() {
		if math.IsInf(u.rhs, 1) {
			return nil, math.Inf(1)
		}

		min := math.Inf(1)
		var (
			next *dStarLiteNode
			cost float64
		)
		for _, _v := range d.world.From(u) {
			v := _v.(*dStarLiteNode)
			w := d.weight(d.world.Edge(u, v))
			if rhs := w + v.g; rhs < min {
				next = v
				min = rhs
				cost = w
			}
		}
		u = next
		weight += cost
		p = append(p, u.Node)
	}
	return p, weight
}

/*
The pseudocode uses the following functions to manage the priority
queue:

      * U.Top() returns a vertex with the smallest priority of all
        vertices in priority queue U.
      * U.TopKey() returns the smallest priority of all vertices in
        priority queue U. (If is empty, then U.TopKey() returns [∞;∞].)
      * U.Pop() deletes the vertex with the smallest priority in
        priority queue U and returns the vertex.
      * U.Insert(s, k) inserts vertex s into priority queue with
        priority k.
      * U.Update(s, k) changes the priority of vertex s in priority
        queue U to k. (It does nothing if the current priority of vertex
        s already equals k.)
      * Finally, U.Remove(s) removes vertex s from priority queue U.
*/

// key is a D* Lite priority queue key.
type key [2]float64

// less returns whether k is less than other. From ISBN:0-262-51129-0 pp476-483:
//
//  k ≤ k' iff k_1 < k'_1 OR (k_1 == k'_1 AND k_2 ≤ k'_2)
//
func (k key) less(other key) bool {
	return k[0] < other[0] || (k[0] == other[0] && k[1] < other[1])
}

// dStarLiteNode adds D* Lite accounting to a graph.Node.
type dStarLiteNode struct {
	graph.Node
	key key
	rhs float64
	g   float64
}

// dStarLiteQueue is an D* Lite priority queue.
type dStarLiteQueue struct {
	indexOf map[int]int
	nodes   []*dStarLiteNode
}

func (q *dStarLiteQueue) Less(i, j int) bool {
	return q.nodes[i].key.less(q.nodes[j].key)
}

func (q *dStarLiteQueue) Swap(i, j int) {
	q.indexOf[q.nodes[i].ID()] = j
	q.indexOf[q.nodes[j].ID()] = i
	q.nodes[i], q.nodes[j] = q.nodes[j], q.nodes[i]
}

func (q *dStarLiteQueue) Len() int {
	return len(q.nodes)
}

func (q *dStarLiteQueue) Push(x interface{}) {
	n := x.(*dStarLiteNode)
	q.indexOf[n.ID()] = len(q.nodes)
	q.nodes = append(q.nodes, n)
}

func (q *dStarLiteQueue) Pop() interface{} {
	n := q.nodes[len(q.nodes)-1]
	q.nodes = q.nodes[:len(q.nodes)-1]
	delete(q.indexOf, n.ID())
	return n
}

// has returns whether the node identified by id is in the queue.
func (q *dStarLiteQueue) has(id int) bool {
	_, ok := q.indexOf[id]
	return ok
}

// top returns the top node in the queue. Note that instead of
// returning a key [∞;∞] when q is empty, the caller checks for
// an empty queue by calling q.Len.
func (q *dStarLiteQueue) top() *dStarLiteNode {
	return q.nodes[0]
}

// insert puts the node u into the queue with the key k.
func (q *dStarLiteQueue) insert(u *dStarLiteNode, k key) {
	heap.Push(q, u)
}

// update updates the node in the queue identified by id with the key k.
func (q *dStarLiteQueue) update(id int, k key) {
	i, ok := q.indexOf[id]
	if !ok {
		return
	}
	q.nodes[i].key = k
	heap.Fix(q, i)
}

// remove removes the node identified by id from the queue.
func (q *dStarLiteQueue) remove(id int) {
	i, ok := q.indexOf[id]
	if !ok {
		return
	}
	heap.Remove(q, i)
}
