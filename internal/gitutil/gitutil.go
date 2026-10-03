// Package gitutil shells out to the git binary for the few operations Gnomon CLI actually
// needs. It is concrete, not an interface, and is tested against real temporary Git repositories
// rather than faked.
package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sort"
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

// Status is what Git-related guidance can derive: whether root is inside a Git repository, and —
// only when it is — whether the working tree has uncommitted changes. Changed is meaningless when
// Available is false.
type Status struct {
	Available bool
	Changed   bool
}

// Inspect derives Status for root. "Not a Git repository" is reported as Status{}, nil — never an
// error. Any other Git failure (missing binary, permission denied, corrupted repository) is
// genuinely unexpected and returned as an error with Git's own stderr preserved.
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
// (`rev-parse --git-dir`). Git's "not a git repository" stderr is the expected, non-error absence
// case; any other failure is returned as an error, never mistaken for "no repository here".
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

// hasChanges reports the one plain boolean cli/DERIVED_FACTS.md defines: any staged, unstaged, or
// untracked difference from HEAD — deliberately coarse; never infer finalize-readiness from it.
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

// Head returns the full commit id of HEAD in root's repository.
func Head(root string) (string, error) {
	return gitOutput(root, "resolving HEAD", "rev-parse", "--verify", "HEAD^{commit}")
}

// ResolveCommit returns the full commit id ref names, or an error if it names no commit.
func ResolveCommit(root, ref string) (string, error) {
	return gitOutput(root, "resolving "+ref, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
}

// ChangedSince lists every file under root that differs from commit base: committed since base,
// staged, unstaged, or untracked (ignored files excluded). Paths are relative to root, with
// forward slashes, sorted and unique.
func ChangedSince(root, base string) ([]string, error) {
	diff, err := gitOutput(root, "listing changes since "+base, "diff", "--name-only", "--relative", "--no-renames", base, "--")
	if err != nil {
		return nil, err
	}
	untracked, err := gitOutput(root, "listing untracked files", "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var paths []string
	for _, line := range strings.Split(diff+"\n"+untracked, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		paths = append(paths, line)
	}
	sort.Strings(paths)
	return paths, nil
}

func gitOutput(root, action string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", wrapGitError(action, err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}
