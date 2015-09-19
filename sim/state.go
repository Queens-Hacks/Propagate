package sim

type tileType int

const (
	dirtTile tileType = iota
	airTile
	plantTile
)

type tile struct {
	T       tileType `json:"tileType"`
	PlantId int      `json: "plantId"`
}

type plant struct {
	Color string `json:"color"`
}

type state struct {
	World  [][]tile `json:"world"`
	Plants []plant  `json:"plants"`
}
