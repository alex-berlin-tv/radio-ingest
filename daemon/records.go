package daemon

import (
	"encoding/json"
	"os"
)

// Contains multiple notification bodies for offline-testing.
type Record map[string]any

// Loads [Records] from a JSON file.
func RecordsFromFile(path string) (Record, error) {
	dt, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rsl Record
	if err := json.Unmarshal(dt, &rsl); err != nil {
		return nil, err
	}
	return rsl, nil
}

// Saves a [Records] instance to a JSON file.
func (r Record) SaveToJson(path string) error {
	dt, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, dt, 0644); err != nil {
		return err
	}
	return nil
}
