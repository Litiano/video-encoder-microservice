package utils

import (
	"encoding/json"
)

func IsJson(s string) error {
	var obj struct{}

	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return err
	}

	return nil
}
