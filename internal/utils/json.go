package utils

import (
	"encoding/json"
)

func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func FromJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}