package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func (cfg *Config) SaveToFs(path string, raw *string) error {
	var contents []byte

	if raw != nil {
		contents = []byte(*raw)
	} else {
		d, err := json.MarshalIndent(cfg, "", "  ")

		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		contents = d
	}

	err := os.WriteFile(path, contents, 0644)

	if err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}
