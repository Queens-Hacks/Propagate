package sim

import (
	"github.com/Queens-Hacks/propagate/sandbox"
	"github.com/Sirupsen/logrus"
)

type tileType int

const (
	dirtTile tileType = iota
	airTile
	plantTile
)

type Location struct {
	x int
	y int
}

type growthRoot struct {
	PlantId int
	Loc     Location
	node    *sandbox.Node
}

type plantInfo struct {
	PlantId int      `json: "plantId"`
	Parent  Location `json: "parent"`
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

type newStateInfo struct {
	ch   <-chan sandbox.NewState
	root *growthRoot
}

const sunAccumulationRate int = 10

// This is called by a timer every n time units
func SimulateTick(s *state) {
	responses := make([]newStateInfo, len(s.roots))

	// Go through each of the plants
	for i := range s.roots {
		// XXX Actually generate a real worldstate - this is an empty one!
		var worldState sandbox.WorldState

		ch := s.roots[i].node.Update(worldState)

		responses = append(responses, newStateInfo{ch, &s.roots[i]})
	}

	for _, response := range responses {
		newState := <-response.ch

		newX := response.root.Loc.x
		newY := response.root.Loc.y

		// XXX newState should actually contain information about the type of
		// operation performed
		if newState.MoveDir == sandbox.Left {
			newX -= 1
		} else if newState.MoveDir == sandbox.Right {
			newX += 1
		} else if newState.MoveDir == sandbox.Up {
			newY -= 1
		} else if newState.MoveDir == sandbox.Down {
			newY += 1
		} else {
			continue
		}

		// Can't move there, it's out of bounds!
		if newY < 0 || newY > len(s.World) {
			logrus.Info("newY out of bounds")
			continue
		}
		if newX < 0 || newX > len(s.World[newY]) {
			logrus.Info("newY out of bounds")
			continue
		}

		// Update the tile entry in the world map with the new growth
		tile := &s.World[newY][newX]
		tile.T = plantTile
		tile.Plant = &plantInfo{
			PlantId: response.root.PlantId,
			Parent:  response.root.Loc,
			Age:     0,
		}

		// Move the growth root to the new location
		response.root.Loc.x = newX
		response.root.Loc.y = newY
	}
}

func AddPlant(s *state, loc Location, id int) *growthRoot {
	// Get the plant information for stuff like the source code
	plant := &s.Plants[id]

	// Create the sandbox node for the plant object
	node := sandbox.AddNode(plant.Source)

	// Create the root node for the object, and append it to the roots list
	root := growthRoot{id, loc, node}
	s.roots = append(s.roots, root)

	// Update the tile under the new root node
	tile := &s.World[loc.y][loc.x]
	tile.T = plantTile
	tile.Plant = &plantInfo{id, loc, 0}

	// Return a reference to the root node we previously appended
	return &s.roots[len(s.roots)-1]
}
