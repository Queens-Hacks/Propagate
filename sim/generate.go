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
			world[y][x] = &Tile{Type: AirTile}
		}
	}

	rand.Seed(time.Now().UTC().UnixNano())
	starting := rand.Intn(height/8) + (height/4)*3
	dirtHeight := starting
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if y > dirtHeight {
				world[y][x] = &Tile{Type: DirtTile}
			}
		}

		heightChange := rand.Float64()
		if heightChange < 0.25 && dirtHeight < height {
			dirtHeight++
		} else if heightChange > 0.75 && dirtHeight > 0 {
			dirtHeight--
		}

		if x > width-50 {
			if dirtHeight > starting {
				dirtHeight -= 1
			} else if dirtHeight < starting {
				dirtHeight += 1
			}
		}
	}

	return &State{
		gameState{world, map[string]*Species{}, []*Plant{}, []*growthRoot{}, 0},
		diff{[]tileDiff{}, map[string]*Species{}, []string{}, []spore{}},
	}
}
