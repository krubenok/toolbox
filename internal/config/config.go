package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Dir returns the toolbox config directory (~/.toolbox).
func Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".toolbox")
}

// Path returns the full path to a config file within ~/.toolbox.
func Path(filename string) string {
	return filepath.Join(Dir(), filename)
}

// Load reads and unmarshals a JSON config file from ~/.toolbox.
// Returns os.ErrNotExist if file doesn't exist.
func Load(filename string, v any) error {
	data, err := os.ReadFile(Path(filename))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Save marshals and writes a JSON config file to ~/.toolbox.
// Creates the directory if it doesn't exist.
func Save(filename string, v any) error {
	dir := Dir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(Path(filename), data, 0644)
}
