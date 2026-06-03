// Package identity holds identifying information for this instance
package identity

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

type Identity struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Role  string `json:"role"`
}

func LoadOrCreate(path string, label string) (*Identity, error) {
	f, err := os.ReadFile(path)
	if err == nil {
		var id Identity
		if json.Unmarshal(f, &id) == nil {
			return &id, nil
		}
	}

	id := &Identity{
		ID:    uuid.NewString(),
		Label: label,
		Role:  "tv",
	}

	data, _ := json.MarshalIndent(id, "", "  ")
	_ = os.WriteFile(path, data, 0o644)

	return id, nil
}
