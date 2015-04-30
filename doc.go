// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package graph implements functions and interfaces to deal with formal discrete graphs. It aims to
be first and foremost flexible, with speed as a strong second priority.

As a departure from previous versions of this package (tag v0.9 and earlier), graph interfaces are 
atomized. There is no longer any notion of a "graph" as a type, but 
rather each interface represents a property of a graph.

The interfaces in this package make immutability guarantees. Unless an input graph explicitly
requires mutability, you can assume it will not be altered. In addition, this graph will, where
possible, return your graph's Nodes and Edges back to you, rather than internally constructed
facsimiles. Exceptions will always be documented.
*/
package graph
