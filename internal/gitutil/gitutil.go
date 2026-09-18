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

// Status is what Git-related guidance can derive: whether root is inside a Git repository at
// all, and — only when it is — whether the working tree has uncommitted changes. A project that
// is simply not a Git repository is a legitimate state (Gnomon's own init never requires Git);
// Changed is meaningless when Available is false.
type Status struct {
	Available bool
	Changed   bool
}

// Inspect derives Status for root. "Not a Git repository" is Git's own well-defined,
// deterministic diagnosis (checked with `git rev-parse --git-dir` before any working-tree
// command runs) and is reported as Status{}, nil — never an error, since Git is a capability
// Git-related guidance uses, not a prerequisite for inspecting root at all. Any other Git failure
// (git missing, permission denied, a corrupted repository) is genuinely unexpected and is
// returned as an error with Git's own stderr preserved, rather than being reduced to a bare exit
// code or silently mistaken for "no repository here".
func Inspect(root string) (Status, error) {
	available, err := isRepository(root)
	if err != nil {
		return Status{}, err
	}
	if !available {
		return Status{}, nil
	}
	changed, err := hasChanges(root)
	if err != nil {
		return Status{}, err
	}
	return Status{Available: true, Changed: changed}, nil
}

// isRepository reports whether root is inside a Git repository, using Git's own canonical check
// (`rev-parse --git-dir`) rather than inferring repository absence from some later command's
// failure. Git's standard "not a git repository" diagnosis on stderr is the expected, non-error
// absence case; any other failure (git not installed, a corrupted repository, permission denied)
// is returned as an error so it is never silently mistaken for "no repository here".
func isRepository(root string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok && strings.Contains(stderr.String(), "not a git repository") {
			return false, nil
		}
		return false, wrapGitError("checking git repository", err, stderr.String())
	}
	return true, nil
}

// hasChanges reports the one plain boolean cli/DERIVED_FACTS.md defines for the working tree:
// any staged, unstaged, or untracked difference from HEAD — no further breakdown. Per that
// document, this is deliberately coarse; a caller must never infer semantic finalize-readiness
// from it. Called only once root is already confirmed to be a Git repository.
func hasChanges(root string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = root
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, wrapGitError("running git status", err, stderr.String())
	}
	return strings.TrimSpace(out.String()) != "", nil
}

// wrapGitError attaches Git's own stderr, when there is any, to a failure — so an unexpected Git
// error never collapses to a bare "exit status N" with the actual diagnosis discarded.
func wrapGitError(action string, err error, stderr string) error {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, stderr)
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
