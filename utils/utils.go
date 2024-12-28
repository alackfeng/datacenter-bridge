package utils

import (
	"encoding/json"
)

// FormatJson - format json or struct data.
func FormatJson(data interface{}) string {
	body2, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(body2)
}

// SliceRemoveDuplicate - remove duplicate elements from slice.
func SliceRemoveDuplicate[T comparable](slice []T) []T {
	keys := make(map[T]struct{})
	list := []T{}
	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}
