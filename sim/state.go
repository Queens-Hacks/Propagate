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
	X int `json:"x"`
	Y int `json:"y"`
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

func mkWorldState(s *state, _ *growthRoot) sandbox.WorldState {
	var ws sandbox.WorldState

	ws.Lighting[sandbox.Left] = 0
	ws.Lighting[sandbox.Right] = 0
	ws.Lighting[sandbox.Up] = 0
	ws.Lighting[sandbox.Down] = 0

	return ws
}

func applyChanges(s *state, root *growthRoot, in sandbox.NewState) {
	newX := root.Loc.X
	newY := root.Loc.Y

	// XXX in should actually contain information about the type of
	// operation performed
	if in.MoveDir == sandbox.Left {
		newX -= 1
	} else if in.MoveDir == sandbox.Right {
		newX += 1
	} else if in.MoveDir == sandbox.Up {
		newY -= 1
	} else if in.MoveDir == sandbox.Down {
		newY += 1
	} else {
		return
	}

	// Can't move there, it's out of bounds!
	if newY < 0 || newY > len(s.World) {
		logrus.Info("newY out of bounds")
		return
	}
	if newX < 0 || newX > len(s.World[newY]) {
		logrus.Info("newY out of bounds")
		return
	}

	// Update the tile entry in the world map with the new growth
	tile := &s.World[newY][newX]
	tile.T = plantTile
	tile.Plant = &plantInfo{
		PlantId: root.PlantId,
		Parent:  root.Loc,
		Age:     0,
	}

	// Move the growth root to the new location
	root.Loc.X = newX
	root.Loc.Y = newY
}

// This is called by a timer every n time units
func SimulateTick(s *state) {
	responses := make([]newStateInfo, len(s.roots))

	// Tell each root to run until the next move operation
	for i := range s.roots {
		root := &s.roots[i]
		ch := root.node.Update(mkWorldState(s, root))
		responses[i] = newStateInfo{ch, root}
	}

	for _, response := range responses {
		newState := <-response.ch
		applyChanges(s, response.root, newState)
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
	tile := &s.World[loc.Y][loc.X]
	tile.T = plantTile
	tile.Plant = &plantInfo{id, loc, 0}

	// Return a reference to the root node we previously appended
	return &s.roots[len(s.roots)-1]
}
