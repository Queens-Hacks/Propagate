package sim

import (
	"encoding/json"
)

func MarshalState(s state) ([]byte, error) {
	return json.Marshal(s)
}

type tileDiff struct {
	Location Location `json:"location"`
	NewTile  tile     `json:"newTile"`
}

type stateDiff struct {
	TileDiffs []tileDiff `json:"tileDiffs"`
	NewPlants []plant    `json:"plants"`
}

func NewBlankDiff() stateDiff {
	return stateDiff{make([]tileDiff, 0), make([]plant, 0)}
}

func (s *stateDiff) makeDirt(x, y int) {
	s.TileDiffs = append(s.TileDiffs, tileDiff{Location{x, y}, tile{dirtTile, nil}})
}

func (s *stateDiff) makeAir(x, y int) {
	s.TileDiffs = append(s.TileDiffs, tileDiff{Location{x, y}, tile{airTile, nil}})
}

func (s *stateDiff) makePlant(x, y int, pi *plantInfo, p plant) {
	s.TileDiffs = append(s.TileDiffs, tileDiff{Location{x, y}, tile{plantTile, pi}})
	s.NewPlants = append(s.NewPlants, p)
}
