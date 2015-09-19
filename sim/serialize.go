package sim

import (
	"encoding/json"
)

func MarshalGameState(s State) ([]byte, error) {
	return json.Marshal(s.state)
}

func MarshalDiff(s State) ([]byte, error) {
	return json.Marshal(s.diff)
}
