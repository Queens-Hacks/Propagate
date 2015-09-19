package simulate

import "encoding/json"

func marshalState(s state) ([]byte, error) {
	return json.Marshal(s)
}
