// Package agentconfig persists exactly one piece of Human/installation-level state: which Agent
// provider Gnomon uses by default. Not a general preferences system. Lives at the OS-appropriate
// per-user config location (os.UserConfigDir()), never inside any project's .gnomon/ tree.
package agentconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the entire persisted shape.
type Config struct {
	DefaultProvider string `json:"default_provider"`
}

// path resolves the config file's location: <os.UserConfigDir()>/gnomon/config.json.
func path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gnomon", "config.json"), nil
}

// Load reads the persisted config. A missing file is not an error — it means "no default
// configured yet."
func Load() (Config, error) {
	p, err := path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// createTempFile and renameFile are Save's write and rename steps as overridable seams — tests
// inject a failure here instead of depending on real filesystem permissions (which root bypasses,
// and which Windows does not enforce the same way for directories).
var (
	createTempFile = os.CreateTemp
	renameFile     = os.Rename
)

// Save persists cfg atomically: written to a temp file, then renamed into place, so a concurrent
// reader never observes a partial file, and a failure never disturbs the previous config.
func Save(cfg Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := createTempFile(dir, ".config-*.json.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := renameFile(tmpPath, p); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
