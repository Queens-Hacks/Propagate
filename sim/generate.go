package sim

func SimpleState(x, y int) state {
	world := make([][]tile, 0, y)
	for i := 0; i < y; i++ {
		t := airTile
		if i > y/2 {
			t = dirtTile
		}
		world = append(world, tileRow(t, x))
	}

	return state{world, map[string]plant{}, []growthRoot{}}
}

func tileRow(t tileType, size int) []tile {
	r := make([]tile, 0, size)
	for i := 0; i < size; i++ {
		r = append(r, tile{T: t})
	}
	return r
}
