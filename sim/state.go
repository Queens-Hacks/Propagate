package sim

import (
	"fmt"
	"github.com/Queens-Hacks/Propagate/sandbox"
)

type TileType int

const (
	DirtTile TileType = iota
	AirTile
	PlantTile
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

type Tile struct {
	T     TileType   `json:"tileType"`
	Plant *plantInfo `json:"plant"`
}

type plant struct {
	Color  string `json:"color"`
	Source string `json:"source"`
	Author string `json:"author"`
	refCnt int
}

type gameState struct {
	World    [][]*Tile         `json:"world"`
	Plants   map[string]*plant `json:"plants"`
	roots    []*growthRoot
	maxPlant int
}

type tileDiff struct {
	Loc  Location `json:"loc"`
	Tile Tile     `json:"tile"`
}

type diff struct {
	TileDiffs     []tileDiff        `json:"tileDiff"`
	NewPlants     map[string]*plant `json:"newPlants"`
	RemovedPlants []string          `json:"removedPlants"`
}

type State struct {
	state gameState
	diff  diff
}

// Records a reference to a plant, causing the plant to be kept in the structure
func (s *State) plantAddRef(plantId string) {
	s.GetPlant(plantId).refCnt++
}

// Records a reference to a plant, causing the plant to be removed from the structure
func (s *State) plantRelease(plantId string) {
	plant := s.GetPlant(plantId)
	plant.refCnt--

	// If the reference count has reached zero, remove the plant from the thing
	if plant.refCnt <= 0 {
		s.diff.RemovedPlants = append(s.diff.RemovedPlants, plantId)
		delete(s.state.Plants, plantId)
	}
}

func (s *State) Width() int {
	return len(s.state.World[0])
}

func (s *State) Height() int {
	return len(s.state.World)
}

func (s *State) GetPlant(plantId string) *plant {
	return s.state.Plants[plantId]
}

// Adds a species to the stateAndDiff, and returns the string key for the plant
// This plant is created with a refCnt of zero, but will not be dropped until
// its reference count hits zero again.
func (s *State) AddSpecies(p plant) string {
	s.state.maxPlant += 1
	key := fmt.Sprintf("%d", s.state.maxPlant)
	s.diff.NewPlants[key] = &p
	s.state.Plants[key] = &p
	return key
}

// Set the tile at a location to a new tile
func (s *State) SetTile(loc Location, new Tile) {
	// Manage the addref and releases
	if new.Plant != nil {
		s.plantAddRef(new.Plant.PlantId)
	}
	old := s.state.World[loc.Y][loc.X]
	if old.Plant != nil {
		s.plantRelease(old.Plant.PlantId)
	}

	// Actually update the tile and record the tilediffs
	*old = new
	s.diff.TileDiffs = append(s.diff.TileDiffs, tileDiff{loc, new})
}

func (s *State) AddPlant(loc Location, id string) *growthRoot {
	plant := s.GetPlant(id)

	// Create the sandbox node for the plant object
	node := sandbox.AddNode(plant.Source)

	// Create the root node for the object, and append it to the roots list
	root := growthRoot{id, loc, node}
	s.state.roots = append(s.state.roots, &root)

	// Set the tile at the base of the plant to a plant tile
	s.SetTile(loc, Tile{PlantTile, &plantInfo{id, loc, 0}})

	// Return a reference to the root node we previously appended
	return &root
}
