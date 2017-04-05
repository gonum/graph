// Copyright ©2015 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import (
	"math"
	"testing"

	"github.com/gonum/graph"
	"github.com/gonum/graph/graphs/gen"
	"github.com/gonum/graph/simple"
)

var (
	gnpDirected_10_tenth   = gnpDirected(10, 0.1)
	gnpDirected_100_tenth  = gnpDirected(100, 0.1)
	gnpDirected_1000_tenth = gnpDirected(1000, 0.1)
	gnpDirected_10_half    = gnpDirected(10, 0.5)
	gnpDirected_100_half   = gnpDirected(100, 0.5)
	gnpDirected_1000_half  = gnpDirected(1000, 0.5)

	gnpUndirected_10   = gnpUndirected(10)
	gnpUndirected_100  = gnpUndirected(100)
	gnpUndirected_1000 = gnpUndirected(1000)
)

func gnpDirected(n int, p float64) graph.Directed {
	g := simple.NewDirectedGraph(0, math.Inf(1))
	gen.Gnp(g, n, p, nil)
	return g
}

func gnpUndirected(n int) graph.Undirected {
	g := simple.NewUndirectedGraph(0, math.Inf(1))
	gen.Gnp(g, n, 1, nil)
	return g
}

func benchmarkTarjanSCC(b *testing.B, g graph.Directed) {
	var sccs [][]graph.Node
	for i := 0; i < b.N; i++ {
		sccs = TarjanSCC(g)
	}
	if len(sccs) == 0 {
		b.Fatal("unexpected number zero-sized SCC set")
	}
}

func benchmarkConnectedComponents(b *testing.B, g graph.Undirected) {
	var cc [][]graph.Node
	for i := 0; i < b.N; i++ {
		cc = ConnectedComponents(g)
	}
	if len(cc) == -1 {
		// Use cc in order to quiet compiler
		b.Fatal("unreachable")
	}
}

func benchmarkConnectedComponentsUF(b *testing.B, g graph.Undirected) {
	var cc [][]graph.Node
	for i := 0; i < b.N; i++ {
		cc = ConnectedComponentsUF(g)
	}
	if len(cc) == -1 {
		// Use cc in order to quiet compiler
		b.Fatal("unreachable")
	}
}

func BenchmarkTarjanSCCGnp_10_tenth(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_10_tenth)
}
func BenchmarkTarjanSCCGnp_100_tenth(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_100_tenth)
}
func BenchmarkTarjanSCCGnp_1000_tenth(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_1000_tenth)
}
func BenchmarkTarjanSCCGnp_10_half(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_10_half)
}
func BenchmarkTarjanSCCGnp_100_half(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_100_half)
}
func BenchmarkTarjanSCCGnp_1000_half(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_1000_half)
}

// Benchmark connected components original version and Union-Find version.

func BenchmarkConnectedComponents_10(b *testing.B) {
	benchmarkConnectedComponents(b, gnpUndirected_10)
}

func BenchmarkConnectedComponents_100(b *testing.B) {
	benchmarkConnectedComponents(b, gnpUndirected_100)
}

func BenchmarkConnectedComponents_1000(b *testing.B) {
	benchmarkConnectedComponents(b, gnpUndirected_1000)
}

func BenchmarkConnectedComponentsUF_10(b *testing.B) {
	benchmarkConnectedComponentsUF(b, gnpUndirected_10)
}

func BenchmarkConnectedComponentsUF_100(b *testing.B) {
	benchmarkConnectedComponentsUF(b, gnpUndirected_100)
}

func BenchmarkConnectedComponentsUF_1000(b *testing.B) {
	benchmarkConnectedComponentsUF(b, gnpUndirected_1000)
}
