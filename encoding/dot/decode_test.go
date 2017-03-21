// Copyright ©2017 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dot_test

import (
	"testing"

	"github.com/gonum/graph"
	"github.com/gonum/graph/encoding/dot"
	"github.com/gonum/graph/simple"
	dotparser "github.com/graphism/dot"
	"github.com/pkg/errors"
)

func TestRoundTrip(t *testing.T) {
	golden := []struct {
		want     string
		directed bool
	}{
		{
			want:     directed,
			directed: true,
		},
		{
			want:     undirected,
			directed: false,
		},
	}
	for i, g := range golden {
		file, err := dotparser.ParseString(g.want)
		if err != nil {
			t.Errorf("i=%d: unable to parse DOT file; %v", i, err)
			continue
		}
		if len(file.Graphs) != 1 {
			t.Errorf("i=%d: invalid number of graphs; expected 1, got %d", i, len(file.Graphs))
			continue
		}
		src := file.Graphs[0]
		var dst dot.Builder
		if g.directed {
			dst = newDirectedGraph()
		} else {
			dst = newUndirectedGraph()
		}
		if err := dot.Copy(dst, src); err != nil {
			t.Errorf("i=%d: unable to copy DOT graph; %v", i, err)
			continue
		}
		buf, err := dot.Marshal(dst, src.ID, "", "\t", false)
		if err != nil {
			t.Errorf("i=%d: unable to marshal graph; %v", i, dst)
			continue
		}
		got := string(buf)
		if got != g.want {
			t.Errorf("i=%d: graph content mismatch; expected `%s`, got `%s`", i, g.want, got)
			continue
		}
	}
}

const directed = `digraph G {
	// Node definitions.
	0 [label="foo 2"];
	1 [label="bar 2"];

	// Edge definitions.
	0 -> 1 [label="baz 2"];
}`

const undirected = `graph H {
	// Node definitions.
	0 [label="foo 2"];
	1 [label="bar 2"];

	// Edge definitions.
	0 -- 1 [label="baz 2"];
}`

// Below follows a minimal implementation of a graph capable of validating the
// round-trip encoding and decoding of DOT graphs with nodes and edges
// containing DOT attributes.

// DirectedGraph extends simple.DirectedGraph to add NewNode and NewEdge
// methods for creating user-defined nodes and edges.
//
// DirectedGraph implements the dot.Builder interface.
type DirectedGraph struct {
	*simple.DirectedGraph
}

// newDirectedGraph returns a new directed capable of creating user-defined
// nodes and edges.
func newDirectedGraph() *DirectedGraph {
	return &DirectedGraph{
		DirectedGraph: simple.NewDirectedGraph(0, 0),
	}
}

// NewNode adds a new node with a unique node ID to the graph.
func (g *DirectedGraph) NewNode() graph.Node {
	n := &Node{
		Node: simple.Node(g.NewNodeID()),
	}
	g.AddNode(n)
	return n
}

// NewEdge adds a new edge from the source to the destination node to the graph,
// or returns the existing edge if already present.
func (g *DirectedGraph) NewEdge(from, to graph.Node) graph.Edge {
	if e := g.Edge(from, to); e != nil {
		return e
	}
	e := &Edge{
		Edge: simple.Edge{F: from, T: to},
	}
	g.SetEdge(e)
	return e
}

// UndirectedGraph extends simple.UndirectedGraph to add NewNode and NewEdge
// methods for creating user-defined nodes and edges.
//
// UndirectedGraph implements the dot.Builder interface.
type UndirectedGraph struct {
	*simple.UndirectedGraph
}

// newUndirectedGraph returns a new undirected capable of creating user-defined
// nodes and edges.
func newUndirectedGraph() *UndirectedGraph {
	return &UndirectedGraph{
		UndirectedGraph: simple.NewUndirectedGraph(0, 0),
	}
}

// NewNode adds a new node with a unique node ID to the graph.
func (g *UndirectedGraph) NewNode() graph.Node {
	n := &Node{
		Node: simple.Node(g.NewNodeID()),
	}
	g.AddNode(n)
	return n
}

// NewEdge adds a new edge from the source to the destination node to the graph,
// or returns the existing edge if already present.
func (g *UndirectedGraph) NewEdge(from, to graph.Node) graph.Edge {
	if e := g.Edge(from, to); e != nil {
		return e
	}
	e := &Edge{
		Edge: simple.Edge{F: from, T: to},
	}
	g.SetEdge(e)
	return e
}

// Node extends simple.Node with a label field to test round-trip encoding and
// decoding of node DOT label attributes.
type Node struct {
	simple.Node
	// Node label.
	Label string
}

// UnmarshalDOTAttr decodes a single DOT attribute.
func (n *Node) UnmarshalDOTAttr(attr dot.Attribute) error {
	switch attr.Key {
	case "label":
		n.Label = attr.Value
	default:
		return errors.Errorf("unable to unmarshal node DOT attribute with key %q", attr.Key)
	}
	return nil
}

// DOTAttributes returns the DOT attributes of the node.
func (n *Node) DOTAttributes() []dot.Attribute {
	var attrs []dot.Attribute
	if len(n.Label) > 0 {
		attr := dot.Attribute{
			Key:   "label",
			Value: n.Label,
		}
		attrs = append(attrs, attr)
	}
	return attrs
}

// Edge extends simple.Edge with a label field to test round-trip encoding and
// decoding of edge DOT label attributes.
type Edge struct {
	simple.Edge
	// Edge label.
	Label string
}

// UnmarshalDOTAttr decodes a single DOT attribute.
func (e *Edge) UnmarshalDOTAttr(attr dot.Attribute) error {
	switch attr.Key {
	case "label":
		e.Label = attr.Value
	default:
		return errors.Errorf("unable to unmarshal edge DOT attribute with key %q", attr.Key)
	}
	return nil
}

// DOTAttributes returns the DOT attributes of the edge.
func (e *Edge) DOTAttributes() []dot.Attribute {
	var attrs []dot.Attribute
	if len(e.Label) > 0 {
		attr := dot.Attribute{
			Key:   "label",
			Value: e.Label,
		}
		attrs = append(attrs, attr)
	}
	return attrs
}
