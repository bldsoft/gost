package json

import (
	"bytes"
	"encoding/json"
)

func MarshalJsonAndJoin(objects ...interface{}) ([]byte, error) {
	if len(objects) == 0 {
		return nil, nil
	}
	marshalled := make([][]byte, 0, len(objects))
	for _, obj := range objects {
		b, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if len(b) <= 2 { // "{}"
			continue
		}
		marshalled = append(marshalled, b)
	}

	if len(marshalled) == 1 {
		return marshalled[0], nil
	}

	for i, b := range marshalled {
		if i < len(objects)-1 {
			b = b[:len(b)-1]
		}
		if i > 0 {
			b = b[1:]
		}
		marshalled[i] = b
	}
	return bytes.Join(marshalled, []byte{','}), nil
}
