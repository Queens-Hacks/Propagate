package sim

import (
	"fmt"

	"github.com/Queens-Hacks/Propagate/sandbox"
	"github.com/Sirupsen/logrus"
)

type tileType int

const (
	dirtTile tileType = iota
	airTile
	plantTile
)

type growthRoot struct {
	PlantId string
	Loc     Location
	node    *sandbox.Node
}

type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type plantInfo struct {
	PlantId string   `json: "plantId"`
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
	refCnt int
}

type gameState struct {
	World    [][]*tile         `json:"world"`
	Plants   map[string]*plant `json:"plants"`
	roots    []*growthRoot
	maxPlant int
}

type tileDiff struct {
	Loc  Location `json:"loc"`
	Tile tile     `json:"tile"`
}

type diff struct {
	TileDiffs     []tileDiff        `json:"tileDiff"`
	NewPlants     map[string]*plant `json:"newPlants"`
	RemovedPlants []string          `json:"removedPlants"`
}

type state struct {
	State gameState
	Diff  diff
}

// Records a reference to a plant, causing the plant to be kept in the structure
func (s *state) plantAddRef(plantId string) {
	s.GetPlant(plantId).refCnt++
}

// Records a reference to a plant, causing the plant to be removed from the structure
func (s *state) plantRelease(plantId string) {
	plant := s.GetPlant(plantId)
	plant.refCnt--

	// If the reference count has reached zero, remove the plant from the thing
	if plant.refCnt <= 0 {
		s.Diff.RemovedPlants = append(s.Diff.RemovedPlants, plantId)
		delete(s.State.Plants, plantId)
	}
}

func (s *state) Width() int {
	return len(s.State.World[0])
}

func (s *state) Height() int {
	return len(s.State.World)
}

func (s *state) GetPlant(plantId string) *plant {
	return s.State.Plants[plantId]
}

// Adds a species to the stateAndDiff, and returns the string key for the plant
// This plant is created with a refCnt of zero, but will not be dropped until
// its reference count hits zero again.
func (s *state) AddSpecies(p plant) string {
	s.State.maxPlant += 1
	key := fmt.Sprintf("%d", s.State.maxPlant)
	s.Diff.NewPlants[key] = &p
	s.State.Plants[key] = &p
	return key
}

// Set the tile at a location to a new tile
func (s *state) SetTile(loc Location, new tile) {
	// Manage the addref and releases
	if new.Plant != nil {
		s.plantAddRef(new.Plant.PlantId)
	}
	old := s.State.World[loc.Y][loc.X]
	if old.Plant != nil {
		s.plantRelease(old.Plant.PlantId)
	}

	// Actually update the tile and record the tilediffs
	*old = new
	s.Diff.TileDiffs = append(s.Diff.TileDiffs, tileDiff{loc, new})
}

func mkWorldState(s *state, _ *growthRoot) sandbox.WorldState {
	var ws sandbox.WorldState

	ws.Lighting[sandbox.Left] = 0
	ws.Lighting[sandbox.Right] = 0
	ws.Lighting[sandbox.Up] = 0
	ws.Lighting[sandbox.Down] = 0

	return ws
}

func applyChanges(s *state, root *growthRoot, in sandbox.NewState) {
	new := root.Loc

	// XXX in should actually contain information about the type of
	// operation performed
	if in.MoveDir == sandbox.Left {
		new.X -= 1
	} else if in.MoveDir == sandbox.Right {
		new.X += 1
	} else if in.MoveDir == sandbox.Up {
		new.Y -= 1
	} else if in.MoveDir == sandbox.Down {
		new.Y += 1
	} else {
		// Super sketchy way to represent do nothing?
		return
	}

	// Can't move there, it's out of bounds!
	if new.Y < 0 || new.Y > s.Height() {
		logrus.Info("newY out of bounds")
		return
	}
	if new.X < 0 || new.X > s.Width() {
		logrus.Info("newY out of bounds")
		return
	}

	s.SetTile(new, tile{plantTile, &plantInfo{
		PlantId: root.PlantId,
		Parent:  root.Loc,
		Age:     0,
	}})

	// XXX Should this go through a method rather than direct mutation?
	// Move the growth root to the new location
	root.Loc = new
}

type newStateInfo struct {
	ch   <-chan sandbox.NewState
	root *growthRoot
}

// This is called by a timer every n time units
func (s *state) SimulateTick() {
	responses := make([]newStateInfo, len(s.State.roots))

	// Tell each root to run until the next move operation
	for i := range s.State.roots {
		root := s.State.roots[i]
		ch := root.node.Update(mkWorldState(s, root))
		responses[i] = newStateInfo{ch, root}
	}

	for _, response := range responses {
		newState := <-response.ch
		applyChanges(s, response.root, newState)
	}
}

func (s *state) AddPlant(loc Location, id string) *growthRoot {
	plant = s.GetPlant(id)

	// Create the sandbox node for the plant object
	node := sandbox.AddNode(plant.Source)

	// Create the root node for the object, and append it to the roots list
	root := growthRoot{id, loc, node}
	s.State.roots = append(s.State.roots, &root)

	// Set the tile at the base of the plant to a plant tile
	s.SetTile(loc, tile{plantTile, &plantInfo{id, loc, 0}})

	// Return a reference to the root node we previously appended
	return &root
}
