package sim

import (
	"fmt"
	"math/rand"

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
	SpeciesId string
	Loc       Location
	Plant     *Plant
	node      *sandbox.Node
}

type Location struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type extraTileInfo struct {
	SpeciesId string   `json:"plantId"`
	Parent    Location `json:"parent"`
	IsRoot    bool     `json:"isRoot"`
	Plant     *Plant
}

type Tile struct {
	Type  TileType       `json:"tileType"`
	Extra *extraTileInfo `json:"plant"`
}

type Plant struct {
	Energy    int
	SpeciesId string
	tiles     map[*Tile]struct{}
}

type Species struct {
	Color  int    `json:"color"`
	Source string `json:"source"`
	Author string `json:"author"`
	refCnt int
}

type gameState struct {
	World    [][]*Tile           `json:"world"`
	Species  map[string]*Species `json:"plants"`
	plants   []*Plant
	roots    []*growthRoot
	maxPlant int
}

type tileDiff struct {
	Loc  Location `json:"loc"`
	Tile Tile     `json:"tile"`
}

type spore struct {
	Location  Location `json:"location"`
	SpeciesId string   `json:"plantId"`
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
	if t.Type == DirtTile {
		p.Location.Y--
		t = s.GetTile(p.Location)
		// dont plant if not on air tile
		if t.Type != AirTile {
			return true
		}

		plant := s.AddPlant(p.SpeciesId)
		s.AddGrowth(p.Location, plant, "")
		return true
	}
	return false
}

type diff struct {
	TileDiffs     []tileDiff          `json:"tileDiff"`
	NewPlants     map[string]*Species `json:"newPlants"`
	RemovedPlants []string            `json:"removedPlants"`
	Spores        []spore             `json:"spores"`
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
	s.GetSpecies(plantId).refCnt++
}

// Records a reference to a plant, causing the plant to be removed from the structure
func (s *State) plantRelease(plantId string) {
	plant := s.GetSpecies(plantId)
	plant.refCnt--

	// If the reference count has reached zero, remove the plant from the thing
	if plant.refCnt <= 0 {
		s.diff.RemovedPlants = append(s.diff.RemovedPlants, plantId)
		delete(s.state.Species, plantId)
	}
}

func (s *State) Width() int {
	return len(s.state.World[0])
}

func (s *State) Height() int {
	return len(s.state.World)
}

func (s *State) GetSpecies(plantId string) *Species {
	return s.state.Species[plantId]
}

// Adds a species to the stateAndDiff, and returns the string key for the plant
// This plant is created with a refCnt of zero, but will not be dropped until
// its reference count hits zero again.
func (s *State) AddSpecies(color int, source string, author string) string {
	p := Species{color, source, author, 0}
	s.state.maxPlant += 1
	key := fmt.Sprintf("%d", s.state.maxPlant)
	s.diff.NewPlants[key] = &p
	s.state.Species[key] = &p
	return key
}

func (s *State) GetTile(loc Location) *Tile {
	return s.state.World[loc.Y][loc.X]
}

/* collision detection */
func reasonableSetTile(old *Tile, new *Tile) bool {
	return old.Type != PlantTile
}

// Set the tile at a location to a new tile
func (s *State) SetTile(loc Location, new Tile) {
	logrus.Info(loc, new)
	// Collision detection
	if !reasonableSetTile(s.GetTile(loc), &new) {
		return
	}

	// Manage the addref and releases
	old := s.GetTile(loc)
	if new.Extra != nil {
		s.plantAddRef(new.Extra.SpeciesId)
		new.Extra.Plant.tiles[old] = struct{}{}
	}
	if old.Extra != nil {
		s.plantRelease(old.Extra.SpeciesId)
		delete(new.Extra.Plant.tiles, old)
	}

	// Actually update the tile and record the tilediffs
	*old = new
	s.diff.TileDiffs = append(s.diff.TileDiffs, tileDiff{loc, new})
}

func (s *State) AddPlant(speciesId string) *Plant {
	plant := &Plant{100, speciesId, map[*Tile]struct{}{}}
	s.state.plants = append(s.state.plants, plant)
	return plant
}

func (s *State) AddGrowth(loc Location, plant *Plant, meta string) *growthRoot {
	if plant == nil {
		panic("crap")
	}
	species := s.GetSpecies(plant.SpeciesId)

	// Create the sandbox node for the plant object
	node := sandbox.AddNode(species.Source, meta)

	// Create the root node for the object, and append it to the roots list
	root := growthRoot{plant.SpeciesId, loc, plant, node}
	s.state.roots = append(s.state.roots, &root)

	isUnderground := false
	if s.GetTile(loc).Type == DirtTile {
		isUnderground = true
	}

	// Set the tile at the base of the plant to a plant tile
	s.SetTile(loc, Tile{PlantTile, &extraTileInfo{plant.SpeciesId, loc, isUnderground, plant}})

	// Return a reference to the root node we previously appended
	return &root
}

func (s *State) lowerToDirt(loc *Location) {
	var base int
	for y := 0; y < s.Height(); y++ {
		t := s.GetTile(Location{loc.X, y})
		if t.Type == DirtTile {
			base = y - 1
			break
		}
	}
	loc.Y = base
}
