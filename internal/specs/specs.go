// Package specs derives Specification existence and identity facts, and performs the
// deterministic Draft Creation operation (Step 1's non-Agent carve-out).
package specs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Identity is a Specification's stable identity, e.g. "SPEC-003".
type Identity string

var identityPattern = regexp.MustCompile(`^SPEC-(\d+)-`)

// NextIdentity computes the next available Specification identity from current repository state
// only — the highest existing SPEC-NNN, plus one, zero-padded to at least three digits.
// SPEC-000 is permanently reserved for the template and never assigned.
func NextIdentity(specsDir string) (Identity, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return "", err
	}
	max := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := identityPattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n == 0 {
			continue
		}
		if n > max {
			max = n
		}
	}
	return Identity(fmt.Sprintf("SPEC-%03d", max+1)), nil
}

// List returns every currently existing Specification identity, sorted by identity — the same
// filesystem-derived existence fact NextIdentity and Exists already use, just enumerated rather
// than probed for one identity or computed as a maximum. SPEC-000 (the template) is never
// included, matching NextIdentity's own permanent reservation of it. This is purely a listing:
// callers must not read any priority or "current Specification" meaning into the resulting order.
func List(specsDir string) ([]Identity, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var ids []Identity
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := identityPattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n == 0 {
			continue
		}
		id := fmt.Sprintf("SPEC-%03d", n)
		if seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, Identity(id))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

// Exists reports whether a Specification with the given identity currently exists, and its path.
func Exists(specsDir string, id Identity) (bool, string, error) {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return false, "", err
	}
	prefix := string(id) + "-"
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == string(id)+".md" || strings.HasPrefix(e.Name(), prefix) {
			return true, filepath.Join(specsDir, e.Name()), nil
		}
	}
	return false, "", nil
}

// ReadContent reads a Specification's current raw content by identity.
func ReadContent(specsDir string, id Identity) ([]byte, error) {
	ok, path, err := Exists(specsDir, id)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("specification %s does not exist", id)
	}
	return os.ReadFile(path)
}

func slugify(title string) string {
	var b strings.Builder
	lastHyphen := true // suppress a leading hyphen
	for _, r := range strings.ToLower(title) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteRune('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// CreateDraft copies the Specification template, substituting only the identity and the given
// title — never generating any behavioral content — and writes the new Draft Specification file.
func CreateDraft(specsDir string, id Identity, title string) (string, error) {
	templatePath := filepath.Join(specsDir, "SPEC-000-use-case-name.md")
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("reading specification template: %w", err)
	}

	numeric := strings.TrimPrefix(string(id), "SPEC-")
	content := string(data)
	content = strings.ReplaceAll(content, "[ID]", numeric)
	content = strings.ReplaceAll(content, "[Use Case Name]", title)
	content = strings.ReplaceAll(content, "[Use case name]", title)

	filename := fmt.Sprintf("%s-%s.md", id, slugify(title))
	destPath := filepath.Join(specsDir, filename)
	if _, err := os.Stat(destPath); err == nil {
		return "", fmt.Errorf("specification file already exists: %s", destPath)
	}
	if err := os.WriteFile(destPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return destPath, nil
}
