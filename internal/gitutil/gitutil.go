// Package gitutil shells out to the git binary for the few operations Gnomon CLI actually
// needs. It is concrete, not an interface, and is tested against real temporary Git repositories
// rather than faked.
package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNoIdentity indicates no usable Git identity (user.name + user.email) is configured.
var ErrNoIdentity = errors.New("no git identity configured")

// Identity resolves "Name <email>" from the repository's Git configuration at root, matching
// Git's own commit-author convention.
func Identity(root string) (string, error) {
	name, err := gitConfig(root, "user.name")
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", ErrNoIdentity
	}
	email, err := gitConfig(root, "user.email")
	if err != nil {
		return "", err
	}
	if email == "" {
		return "", ErrNoIdentity
	}
	return fmt.Sprintf("%s <%s>", name, email), nil
}

// HasChanges reports the one plain boolean cli/DERIVED_FACTS.md defines for the working tree:
// any staged, unstaged, or untracked difference from HEAD — no further breakdown. Per that
// document, this is deliberately coarse; a caller must never infer semantic finalize-readiness
// from it.
func HasChanges(root string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("running git status: %w", err)
	}
	return strings.TrimSpace(out.String()) != "", nil
}

func gitConfig(root, key string) (string, error) {
	cmd := exec.Command("git", "config", key)
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		// `git config` exits non-zero when the key is simply unset — that's absence, not failure.
		if _, ok := err.(*exec.ExitError); ok {
			return "", nil
		}
		return "", fmt.Errorf("running git config %s: %w", key, err)
	}
	return strings.TrimSpace(out.String()), nil
}
