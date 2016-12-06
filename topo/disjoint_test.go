// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import (
	"testing"

	"github.com/gonum/graph/simple"
)

func TestDisjointSetMakeSet(t *testing.T) {
	ds := newDisjointSet()
	if ds.master == nil {
		t.Fatal("Internal disjoint set map erroneously nil")
	} else if len(ds.master) != 0 {
		t.Error("Disjoint set master map of wrong size")
	}

	ds.makeSet(simple.Node(3))
	if len(ds.master) != 1 {
		t.Error("Disjoint set master map of wrong size")
	}

	if node, ok := ds.master[3]; !ok {
		t.Error("Make set did not successfully add element")
	} else {
		if node == nil {
			t.Fatal("Disjoint set node from makeSet is nil")
		}

		if node.rank != 0 {
			t.Error("Node rank set incorrectly")
		}

		if node.parent != node {
			t.Error("Node parent set incorrectly")
		}
	}
}

func TestDisjointSetFind(t *testing.T) {
	ds := newDisjointSet()

	ds.makeSet(simple.Node(3))
	ds.makeSet(simple.Node(5))

	if ds.find(simple.Node(3)) == ds.find(simple.Node(simple.Node(5))) {
		t.Error("Disjoint sets incorrectly found to be the same")
	}
}

func TestUnion(t *testing.T) {
	ds := newDisjointSet()

	ds.makeSet(simple.Node(3))
	ds.makeSet(simple.Node(5))

	ds.union(ds.find(simple.Node(3)), ds.find(simple.Node(5)))

	if ds.find(simple.Node(3)) != ds.find(simple.Node(5)) {
		t.Error("Sets found to be disjoint after union")
	}
}

func BenchmarkCC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ds := newDisjointSet()
		ds.makeSet(simple.Node(3))
		ds.makeSet(simple.Node(5))

		ds.union(ds.find(simple.Node(3)), ds.find(simple.Node(5)))

		ds.connectedComponents()
	}
}
