package simulation

type plant struct {
	color string
}

type tileType int

const (
	dirtTile tileType = iota
	airTile
	plantTile
)

type tile struct {
	t       tileType
	plantId int
}

type state struct {
	world  [][]tile
	plants map[int]plant
}
