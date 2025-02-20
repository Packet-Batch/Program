package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func (cfg *Config) Load(data []byte) error {
	err := json.Unmarshal(data, cfg)

	return err
}

func (cfg *Config) LoadFromFs(path string) error {
	f, err := os.Open(path)

	if err != nil {
		return err
	}

	defer f.Close()

	stat, err := f.Stat()

	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	data := make([]byte, stat.Size())

	_, err = f.Read(data)

	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	err = cfg.Load(data)

	if err != nil {
		return fmt.Errorf("failed to load and unmarshal data: %v", err)
	}

	return nil
}
