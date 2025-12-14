package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Dir returns the toolbox config directory (~/.toolbox).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".toolbox"), nil
}

// Path returns the full path to a config file within ~/.toolbox.
func Path(filename string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// Load reads and unmarshals a JSON config file from ~/.toolbox.
// Returns os.ErrNotExist if file doesn't exist.
func Load(filename string, v any) error {
	path, err := Path(filename)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Save marshals and writes a JSON config file to ~/.toolbox.
// Creates the directory if it doesn't exist.
func Save(filename string, v any) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	path, err := Path(filename)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
