// Package project locates a Gnomon project's root and materializes the
// canonical .gnomon/ layout, per cli/PROJECT_INITIALIZATION.md.
package project

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed bundled
var bundled embed.FS

const bundledRoot = "bundled"

// SupportedContractVersion is the Workflow/Result Contract version this CLI understands.
const SupportedContractVersion = "1"

// GnomonDirName is the materialized Core root inside a project.
const GnomonDirName = ".gnomon"

// bundledSections are the canonical layout subdirectories materialized from the bundle.
var bundledSections = []string{"workflows", "specifications", "evaluations", "contracts", "decisions", "context"}

// Layout resolves paths within a located Gnomon project.
type Layout struct {
	Root string
}

func (l Layout) GnomonRoot() string        { return filepath.Join(l.Root, GnomonDirName) }
func (l Layout) WorkflowsDir() string      { return filepath.Join(l.GnomonRoot(), "workflows") }
func (l Layout) SpecificationsDir() string { return filepath.Join(l.GnomonRoot(), "specifications") }
func (l Layout) ApprovalsDir() string      { return filepath.Join(l.GnomonRoot(), "approvals") }
func (l Layout) ContractVersionPath() string {
	return filepath.Join(l.GnomonRoot(), "CONTRACT_VERSION")
}

// Locate walks upward from start looking for a .gnomon directory, mirroring how Git locates .git.
func Locate(start string) (Layout, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return Layout{}, err
	}
	for {
		candidate := filepath.Join(dir, GnomonDirName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return Layout{Root: dir}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return Layout{}, fmt.Errorf("no %s project found above %s (run `gnomon init` first)", GnomonDirName, start)
		}
		dir = parent
	}
}

// ContractVersion reads the project-local contract version, "" if absent.
func (l Layout) ContractVersion() (string, error) {
	b, err := os.ReadFile(l.ContractVersionPath())
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// IsInitialized reports the structural facts PROJECT_INITIALIZATION.md defines as "initialized".
func (l Layout) IsInitialized() bool {
	info, err := os.Stat(l.GnomonRoot())
	if err != nil || !info.IsDir() {
		return false
	}
	v, err := l.ContractVersion()
	return err == nil && v != ""
}

// MissingSections reports which canonical .gnomon/ subdirectories are absent — the structural-
// integrity check `gnomon validate` performs, reusing exactly the same canonical layout
// knowledge (bundledSections, plus approvals/) that Materialize already fills from, so the two
// can never define "canonical layout" differently.
func (l Layout) MissingSections() []string {
	sections := append(append([]string{}, bundledSections...), "approvals")
	var missing []string
	for _, section := range sections {
		info, err := os.Stat(filepath.Join(l.GnomonRoot(), section))
		if err != nil || !info.IsDir() {
			missing = append(missing, section)
		}
	}
	return missing
}

// Materialize fills any missing part of the canonical layout at root/.gnomon, never overwriting
// existing content, and writes CONTRACT_VERSION last. It refuses if an existing, unsupported
// contract version is already present.
func Materialize(root string) error {
	l := Layout{Root: root}

	if v, err := l.ContractVersion(); err != nil {
		return err
	} else if v != "" && v != SupportedContractVersion {
		return fmt.Errorf("project contract version %q is not supported by this CLI (supports %q)", v, SupportedContractVersion)
	}

	if err := os.MkdirAll(l.ApprovalsDir(), 0o755); err != nil {
		return err
	}

	for _, section := range bundledSections {
		dest := filepath.Join(l.GnomonRoot(), section)
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		if err := fillFromBundle(section, dest); err != nil {
			return err
		}
	}

	if v, err := l.ContractVersion(); err != nil {
		return err
	} else if v == "" {
		if err := os.WriteFile(l.ContractVersionPath(), []byte(SupportedContractVersion+"\n"), 0o644); err != nil {
			return err
		}
	}

	return nil
}

// fillFromBundle copies every file the bundle declares for section into destDir, skipping any
// file that already exists there — the "write only what is absent" rule.
func fillFromBundle(section, destDir string) error {
	srcDir := path.Join(bundledRoot, section)
	entries, err := fs.ReadDir(bundled, srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		destPath := filepath.Join(destDir, e.Name())
		if _, err := os.Stat(destPath); err == nil {
			continue
		}
		data, err := fs.ReadFile(bundled, path.Join(srcDir, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
