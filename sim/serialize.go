package sim

import (
	"encoding/json"
)

func MarshalGameState(s state) ([]byte, error) {
	return json.Marshal(s.State)
}

func MarshalDiff(s state) ([]byte, error) {
	return json.Marshal(s.Diff)
}
