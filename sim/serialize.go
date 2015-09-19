package sim

import "encoding/json"

func MarshalState(s state) ([]byte, error) {
	return json.Marshal(s)
}
