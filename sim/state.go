package sim

import (
	"github.com/Queens-Hacks/propagate/sandbox"
)

type tileType int

const (
	dirtTile tileType = iota
	airTile
	plantTile
)

type location struct {
	x int
	y int
}

type growthRoot struct {
	PlantId int
	Loc     location
	node    *sandbox.Node
}

type plantInfo struct {
	PlantId int      `json: "plantId"`
	Parent  location `json: "parent"`
	Age     int      `json: "age"`
}

type tile struct {
	T     tileType   `json:"tileType"`
	Plant *plantInfo `json:"plant"`
}

type plant struct {
	Color  string `json:"color"`
	Source string `json:"source"`
	Author string `json:"author"`
}

type state struct {
	World  [][]tile `json:"world"`
	Plants []plant  `json:"plants"`
	roots  []growthRoot
}

const sunAccumulationRate int = 10

// This is called by a timer every n time units
func SimulateTick(s *state) {
	chs := make([]<-chan sandbox.NewState, len(s.roots))

	// Go through each of the plants
	for _, root := range s.roots {
		ch := root.node.Update(sandbox.WorldState{
			Lighting: make(map[sandbox.Direction]float64),
		})

		chs = append(chs, ch)
	}

	for _, ch := range chs {
		<-ch
		// newstate := <-ch

		// Perofrm the update
	}
}
