package simulation

func simpleState(x, y int) state {
	world := make([][]tile, 0, y)
	for i := 0; i < y; i++ {
		t := airTile
		if i > y/2 {
			t = dirtTile
		}
		append(world, tileRow(t, x))
	}

	for i := y / 2; i < y; i++ {

	}
}

func tileRow(t tileType, size int) []tile {
	r := make([]tile, 0, size)
	for i := 0; i < size; i++ {
		append(r, tile{t: t})
	}
}
