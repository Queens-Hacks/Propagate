package sim

import (
	"math/rand"
	"time"
)

func NewState(width, height int) *State {
	world := make([][]*Tile, 0, height)
	for y := 0; y < height; y++ {
		world = append(world, make([]*Tile, width))
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			world[y][x] = &Tile{T: AirTile}
		}
	}

	rand.Seed(time.Now().UTC().UnixNano())
	dirtHeight := rand.Intn(height/2) + height/2
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if y > dirtHeight {
				world[y][x] = &Tile{T: DirtTile}
			}
		}

		heightChange := rand.Float64()
		if heightChange < 0.25 && dirtHeight < height {
			dirtHeight++
		} else if heightChange > 0.75 && dirtHeight > 0 {
			dirtHeight--
		}
	}

	return &State{
		gameState{world, map[string]*Plant{}, []*growthRoot{}, 0},
		diff{[]tileDiff{}, map[string]*Plant{}, []string{}},
	}
}
