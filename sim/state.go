package sim

import (
	"fmt"

	"encoding/json"
	"sync"

	"github.com/Queens-Hacks/Propagate/sandbox"
	"github.com/Sirupsen/logrus"
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
	state        gameState
	diff         diff
	lock         sync.RWMutex
	marshalState []byte
	marshalDiff  []byte
}

// Finalize the current state information
func (s *State) Finalize() {
	s.lock.Lock()
	defer s.lock.Unlock()

	md, err := json.Marshal(s.diff)
	if err != nil {
		logrus.Fatal(err)
	}

	ms, err := json.Marshal(s.state)
	if err != nil {
		logrus.Fatal(err)
	}

	s.marshalDiff = md
	s.marshalState = ms

	newDiff := diff{[]tileDiff{}, map[string]*plant{}, []string{}}
	s.diff = newDiff
}

func (s *State) MarshalState() []byte {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.marshalState
}

func (s *State) MarshalDiff() []byte {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.marshalDiff
}

// Records a reference to a plant, causing the plant to be kept in the structure
func (s *State) plantAddRef(plantId string) {
	s.getPlant(plantId).refCnt++
}

// Records a reference to a plant, causing the plant to be removed from the structure
func (s *State) plantRelease(plantId string) {
	plant := s.getPlant(plantId)
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

func (s *State) getPlant(plantId string) *plant {
	return s.state.Plants[plantId]
}

// Adds a species to the stateAndDiff, and returns the string key for the plant
// This plant is created with a refCnt of zero, but will not be dropped until
// its reference count hits zero again.
func (s *State) addSpecies(p plant) string {
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

func (s *State) addPlant(loc Location, id string) *growthRoot {
	plant := s.getPlant(id)

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
