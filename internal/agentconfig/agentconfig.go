// Package agentconfig persists exactly one piece of Human/installation-level state: which Agent
// provider Gnomon should use by default when an invocation does not name one explicitly. This is
// deliberately not a general preferences system — if Gnomon ever needs more per-user state, that
// is a new decision to make then, not an assumed extension of this file's shape.
//
// The config lives at the OS-appropriate per-user configuration location (os.UserConfigDir()),
// never inside any project's own .gnomon/ tree — the default Agent is a property of the Human
// using Gnomon, not of any one project, and is never committed to any repository.
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
// configured yet," a normal, expected state distinct from any real failure, so callers can
// treat a zero-value Config and a freshly-written empty one identically.
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

// Save persists cfg atomically: written to a temp file in the same directory, then renamed into
// place, so a concurrent reader, or an interruption partway through, never observes a partial
// file — and, on failure, never disturbs whatever config previously existed there.
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

	tmp, err := os.CreateTemp(dir, ".config-*.json.tmp")
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
	if err := os.Rename(tmpPath, p); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}
