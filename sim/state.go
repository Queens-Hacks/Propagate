package sim

import (
	"fmt"
	"math/rand"

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
	PlantId string   `json:"plantId"`
	Parent  Location `json:"parent"`
	IsRoot  bool     `json:"isRoot"`
}

type Tile struct {
	T     TileType   `json:"tileType"`
	Plant *plantInfo `json:"plant"`
}

type Plant struct {
	Color  int    `json:"color"`
	Source string `json:"source"`
	Author string `json:"author"`
	refCnt int
}

type gameState struct {
	World    [][]*Tile         `json:"world"`
	Plants   map[string]*Plant `json:"plants"`
	roots    []*growthRoot
	maxPlant int
}

type tileDiff struct {
	Loc  Location `json:"loc"`
	Tile Tile     `json:"tile"`
}

type spore struct {
	Location Location `json:"location"`
	PlantId  string   `json:"plantId"`
}

func (s *State) AddSpore(loc Location, plantId string) {
	// TODO add in bounds checks
	s.diff.Spores = append(s.diff.Spores, spore{loc, plantId})
}

func (s *State) UpdateSpore(p *spore) bool {
	dx := rand.Intn(3) - 1
	dy := rand.Intn(2)
	p.Location.X += dx
	p.Location.Y += dy
	p.Location.X = (p.Location.X + s.Width()) % s.Width()

	t := s.GetTile(p.Location)
	if t.T == DirtTile {
		p.Location.Y--
		t = s.GetTile(p.Location)
		// dont plant if not on air tile
		if t.T != AirTile {
			return true
		}
		s.AddPlant(p.Location, p.PlantId, "")
		return true
	}
	return false
}

type diff struct {
	TileDiffs     []tileDiff        `json:"tileDiff"`
	NewPlants     map[string]*Plant `json:"newPlants"`
	RemovedPlants []string          `json:"removedPlants"`
	Spores        []spore           `json:"spores"`
}

func (d *diff) isEmpty() bool {
	return len(d.TileDiffs) == 0 && len(d.NewPlants) == 0 && len(d.RemovedPlants) == 0 && len(d.Spores) == 0
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

func (s *State) GetPlant(plantId string) *Plant {
	return s.state.Plants[plantId]
}

// Adds a species to the stateAndDiff, and returns the string key for the plant
// This plant is created with a refCnt of zero, but will not be dropped until
// its reference count hits zero again.
func (s *State) AddSpecies(color int, source string, author string) string {
	p := Plant{color, source, author, 0}
	s.state.maxPlant += 1
	key := fmt.Sprintf("%d", s.state.maxPlant)
	s.diff.NewPlants[key] = &p
	s.state.Plants[key] = &p
	return key
}

func (s *State) GetTile(loc Location) *Tile {
	return s.state.World[loc.Y][loc.X]
}

// Set the tile at a location to a new tile
func (s *State) SetTile(loc Location, new Tile) {
	// Manage the addref and releases
	if new.Plant != nil {
		s.plantAddRef(new.Plant.PlantId)
	}
	old := s.GetTile(loc)
	if old.Plant != nil {
		s.plantRelease(old.Plant.PlantId)
	}

	// Actually update the tile and record the tilediffs
	*old = new
	s.diff.TileDiffs = append(s.diff.TileDiffs, tileDiff{loc, new})
}

func (s *State) AddPlant(loc Location, id string, meta string) *growthRoot {
	plant := s.GetPlant(id)

	// Create the sandbox node for the plant object
	node := sandbox.AddNode(plant.Source, meta)

	s.lowerToDirt(&loc)

	// Create the root node for the object, and append it to the roots list
	root := growthRoot{id, loc, node}
	s.state.roots = append(s.state.roots, &root)

	isUnderground := false
	if s.GetTile(loc).T == DirtTile {
		isUnderground = true
	}

	// Set the tile at the base of the plant to a plant tile
	s.SetTile(loc, Tile{PlantTile, &plantInfo{id, loc, isUnderground}})

	// Return a reference to the root node we previously appended
	return &root
}

func (s *State) lowerToDirt(loc *Location) {
	var base int
	for y := 0; y < s.Height(); y++ {
		t := s.GetTile(Location{loc.X, y})
		if t.T == DirtTile {
			base = y - 1
			break
		}
	}
	loc.Y = base
}
