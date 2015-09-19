package sim

import ()

func NewState(width, height int) *State {
	var s State

	world := make([][]*Tile, height, width)
	for y := 0; y < height; y++ {
		for x := 0; x < height; x++ {
			t := AirTile
			if y < height/2 {
				t = DirtTile
			}

			world[x][y] = &Tile{T: t}
		}
	}

	s.state.World = world

	return &s
}

func SimpleState(x, y int) *State {
	world := make([][]*Tile, 0, y)
	for i := 0; i < y; i++ {
		t := AirTile
		if i > y/2 {
			t = DirtTile
		}
		world = append(world, tileRow(t, x))
	}

	return &State{
		gameState{world, map[string]*Plant{}, []*growthRoot{}, 0},
		diff{[]tileDiff{}, map[string]*Plant{}, []string{}},
	}
}

func tileRow(t TileType, size int) []*Tile {
	r := make([]*Tile, 0, size)
	for i := 0; i < size; i++ {
		r = append(r, &Tile{T: t})
	}
	return r
}
