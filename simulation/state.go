package simulation

type tileType int

const (
	dirtTile tileType = iota
	airTile
	plantTile
)

type tile struct {
	t       tileType `json:"tileType"`
	plantId int      `json: "plantId"`
}

type plant struct {
	color string `json:"color"`
}

type state struct {
	world  [][]tile      `json:"world"`
	plants map[int]plant `json:"plants"`
}
