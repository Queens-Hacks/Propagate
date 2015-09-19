package sim

import (
	"sync"
)

func SimpleState(x, y int) State {
	world := make([][]*Tile, 0, y)
	for i := 0; i < y; i++ {
		t := airTile
		if i > y/2 {
			t = dirtTile
		}
		world = append(world, tileRow(t, x))
	}

	return State{
		gameState{world, map[string]*plant{}, []*growthRoot{}, 0},
		diff{[]tileDiff{}, map[string]*plant{}, []string{}},
		sync.RWMutex{},
		[]byte{},
		[]byte{},
	}
}

func tileRow(t tileType, size int) []*Tile {
	r := make([]*Tile, 0, size)
	for i := 0; i < size; i++ {
		r = append(r, &Tile{T: t})
	}
	return r
}
