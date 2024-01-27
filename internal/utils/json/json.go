package json

import "encoding/json"

func UnmarshalJSON[T any](source []byte) (T, error) {
	var target T

	if err := json.Unmarshal(source, &target); err != nil {
		return target, err
	}

	return target, nil
}

func MarshalJSON[T any](t *T) (string, error) {
	bytes, err := json.Marshal(t)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
