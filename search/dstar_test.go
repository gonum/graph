package dstar_test

import (
	"github.com/gonum/graph/concrete"
	"github.com/gonum/graph/search"
)

func TestDStarLite(t *testing.T) {
	tg := concrete.NewTileGraph(10, 10, true)
	realGraph := concrete.NewTileGraph(10, 10, true)
	realGraph.SetPassability(4, 1, false)
	realGraph.SetPassability(4, 2, false)
	realGraph.SetPassability(4, 3, false)
	realGraph.SetPassability(4, 4, false)
	realGraph.SetPassability(4, 5, false)
	realGraph.SetPassability(4, 6, false)
	realGraph.SetPassability(4, 7, false)
	realGraph.SetPassability(4, 8, false)
	realGraph.SetPassability(4, 9, false)

	rows, cols := tg.Dimensions()
	dStarInstance := search.InitDStar(concrete.GonumNode(5), tg.CoordsToNode(rows-1, cols-1), tg, nil, nil)

	path := []graph.Node{concrete.GonumNode(5)}

	var succ graph.Node
	var err error
	for succ, err = dStarInstance.Step(); err != nil && succ != nil && succ.ID() != tg.CoordsToID(rows-1, cols-1); succ, err = dStarInstance.Step() {
		path = append(path, succ)

		knownSuccs := tg.Neighbors(succ)
		realSuccs := realGraph.Neighbors(succ)
		knownSet := set.NewSet()
		for _, k := range knownSuccs {
			knownSet.Add(k)
		}
		realSet := set.NewSet()
		for _, k := range realSuccs {
			realSet.Add(k)
		}

		knownSet.Diff(knownSet, realSet)
		updatedEdges := []graph.Edge{}
		for _, toRemove := range knownSet.Elements() {
			node := toRemove.(graph.Node)
			updatedEdges = append(updatedEdges, concrete.GonumEdge{H: succ, T: node})
			row, col := tg.IDToCoords(node.ID())
			tg.SetPassability(row, col, false)
		}

		dStarInstance.Update(nil, updatedEdges)
	}

	if succ == nil || err != nil {
		t.Error("Got erroneous error: %s\nPath before error encountered: %#v", err, path)
	}
}
