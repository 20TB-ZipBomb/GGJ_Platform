package json

import (
	"encoding/json"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
)

// Generic wrapper for unmarshalling JSON with custom struct definitions.
func UnmarshalJSON[T any](source []byte) T {
	var target T

	if err := json.Unmarshal(source, &target); err != nil {
		logger.Errorf("Failed to unmarshal JSON: %v", err)
		return target
	}

	return target
}

// Generic wrapper for marshalling JSON with custom struct definitions.
func MarshalJSON[T any](t *T) string {
	bytes, err := json.Marshal(t)

	if err != nil {
		logger.Errorf("Failed to marshal JSON: %v", err)
		return ""
	}

	return string(bytes)
}

// Generic helper function for wrapping a JSON marshal to raw bytes.
func MarshalJSONBytes[T any](t *T) []byte {
	return []byte(MarshalJSON[T](t))
}
