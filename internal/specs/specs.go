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

// List returns every currently existing Specification identity, sorted by identity. SPEC-000 (the
// template) is never included. Purely a listing — callers must not read priority into the order.
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

// Slugify mechanically normalizes s into a lowercase, hyphen-separated, filename-safe form:
// non-alphanumeric runs collapse to a single hyphen, no leading/trailing hyphen. No semantic
// transformation (no stop-word removal, no NLP) — callers needing a meaningful slug supply
// their own already-concise input.
func Slugify(s string) string {
	var b strings.Builder
	lastHyphen := true // suppress a leading hyphen
	for _, r := range strings.ToLower(s) {
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

// CreateDraft writes a new Draft from template, substituting only the identity and display title —
// never generating behavioral content — under a name built from slug. slug must already be
// normalized (Slugify) by the caller; CreateDraft performs no semantic judgment of its own.
func CreateDraft(specsDir string, id Identity, title, slug string, template []byte) (string, error) {
	if slug == "" {
		return "", fmt.Errorf("a non-empty slug is required")
	}
	content := renderTemplate(template, id, title)

	filename := fmt.Sprintf("%s-%s.md", id, slug)
	destPath := filepath.Join(specsDir, filename)
	if _, err := os.Stat(destPath); err == nil {
		return "", fmt.Errorf("specification file already exists: %s", destPath)
	}
	if err := os.WriteFile(destPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return destPath, nil
}

// renderTemplate applies Draft Creation's only substitutions: the identity's number and the title.
func renderTemplate(template []byte, id Identity, title string) string {
	numeric := strings.TrimPrefix(string(id), "SPEC-")
	content := string(template)
	content = strings.ReplaceAll(content, "[ID]", numeric)
	content = strings.ReplaceAll(content, "[Use Case Name]", title)
	content = strings.ReplaceAll(content, "[Use case name]", title)
	return content
}

// RequestLinePrefix starts the block in which a Draft keeps the request it was started from
// (`gnomon define "<request>"`), directly below its title.
const RequestLinePrefix = "> Request: "

// AuthoredLines returns the non-blank lines of spec that a freshly created Draft from template
// would not contain — what someone actually wrote. Structural lines (separators, bare list
// markers) never count.
func AuthoredLines(template, spec []byte, id Identity) []string {
	rendered := map[string]bool{}
	for _, line := range strings.Split(renderTemplate(template, id, Title(id, spec)), "\n") {
		rendered[strings.TrimSpace(line)] = true
	}
	var authored []string
	inRequest := false
	for _, line := range strings.Split(string(spec), "\n") {
		t := strings.TrimSpace(line)
		// The request a Draft was started from (gnomon define) is not Specification content.
		if strings.HasPrefix(line, RequestLinePrefix) || (inRequest && strings.HasPrefix(line, ">")) {
			inRequest = true
			continue
		}
		inRequest = false
		if t == "" || rendered[t] || strings.Trim(t, "-*#=_|: ") == "" {
			continue
		}
		authored = append(authored, t)
	}
	return authored
}

// EffectivelyEmpty reports whether spec contains nothing beyond what Draft Creation would have
// produced from at least one of the given templates — approving it would approve no behavior.
func EffectivelyEmpty(templates [][]byte, spec []byte, id Identity) bool {
	for _, t := range templates {
		if len(AuthoredLines(t, spec, id)) == 0 {
			return true
		}
	}
	return false
}

// ClosestTemplate returns the template with the fewest authored lines relative to spec — the one
// it was most likely created from — or nil if templates is empty.
func ClosestTemplate(templates [][]byte, spec []byte, id Identity) []byte {
	var best []byte
	bestN := -1
	for _, t := range templates {
		if n := len(AuthoredLines(t, spec, id)); bestN < 0 || n < bestN {
			best, bestN = t, n
		}
	}
	return best
}

var titleHeadingPattern = regexp.MustCompile(`^#\s*SPEC-\S+\s*[—–-]\s*(.+?)\s*$`)

// Title mechanically extracts the display title from a Specification's first-line heading
// (template convention: "# SPEC-[ID] — [Use Case Name]"). Falls back to the identity itself when
// the heading is absent or malformed.
func Title(id Identity, content []byte) string {
	for _, line := range strings.Split(string(content), "\n") {
		if m := titleHeadingPattern.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			return m[1]
		}
		if strings.TrimSpace(line) != "" {
			break // only the first non-blank line is ever considered the heading
		}
	}
	return string(id)
}

// substitutedTemplateTokens are the literal tokens renderTemplate itself replaces when creating a
// Specification; keep in sync with it.
var substitutedTemplateTokens = map[string]bool{
	"[ID]":            true,
	"[Use Case Name]": true,
	"[Use case name]": true,
}

// bracketToken matches one single-line, non-nested "[...]" span — a candidate template
// placeholder token.
var bracketToken = regexp.MustCompile(`\[[^\[\]]+\]`)

// RemainingPlaceholders returns every placeholder token from template that still appears verbatim
// in spec, in order of first appearance — the signal that a Specification still reads like the
// untouched template. A token CreateDraft itself substitutes is never included; a bracketed span
// the Human wrote themselves, absent from the template, is never reported either.
func RemainingPlaceholders(template, spec []byte) []string {
	specText := string(spec)
	var remaining []string
	for _, tok := range templatePlaceholderTokens(template) {
		if strings.Contains(specText, tok) {
			remaining = append(remaining, tok)
		}
	}
	return remaining
}

// templatePlaceholderTokens extracts every distinct "[...]" placeholder candidate from template,
// in order of first appearance, excluding: the text part of a markdown link or image (immediately
// followed by "("), a span inside inline code (an odd number of backticks precede it on the same
// line), a span inside a fenced code block, and any token CreateDraft itself substitutes.
func templatePlaceholderTokens(template []byte) []string {
	var tokens []string
	seen := map[string]bool{}
	inFence := false

	for _, line := range strings.Split(string(template), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		for _, loc := range bracketToken.FindAllStringIndex(line, -1) {
			start, end := loc[0], loc[1]
			if end < len(line) && line[end] == '(' {
				continue // markdown link/image text, e.g. "[label](url)"
			}
			if strings.Count(line[:start], "`")%2 == 1 {
				continue // inside an inline code span
			}
			tok := line[start:end]
			if substitutedTemplateTokens[tok] || seen[tok] {
				continue
			}
			seen[tok] = true
			tokens = append(tokens, tok)
		}
	}
	return tokens
}
