package project

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Fingerprint identifies a shipped file's content independent of line endings and trailing
// whitespace, so a copy checked out on Windows still matches the version Gnomon shipped.
func Fingerprint(data []byte) string {
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	sum := sha256.Sum256([]byte(strings.TrimRight(strings.Join(lines, "\n"), "\n")))
	return hex.EncodeToString(sum[:])
}

// IsStock reports whether data is byte-for-byte (modulo whitespace) some version of the file
// Gnomon has shipped under name — an unmodified Gnomon copy, never a project customization or
// project-written knowledge. The fingerprint set is generated from git history (tools/stockgen).
func IsStock(name string, data []byte) bool {
	fp := Fingerprint(data)
	for _, h := range stockFingerprints[name] {
		if h == fp {
			return true
		}
	}
	return false
}

// WorkflowFile is one workflow in effect for a project.
type WorkflowFile struct {
	Name string // filename, e.g. "implementation.md"
	Data []byte

	// ProjectPath is set when the project's own .gnomon/workflows/ file is in effect — an intentional
	// customization or an additional, project-defined workflow. Empty means the bundled version.
	ProjectPath string
}

// Customized reports whether the project's own file is in effect rather than the bundled one.
func (w WorkflowFile) Customized() bool { return w.ProjectPath != "" }

// Workflows returns every workflow in effect, sorted by filename: the bundled set, with any
// project file of the same name replacing it, plus any additional project workflow. A project file
// that is an unmodified copy of some shipped version (typically left behind by an older
// `gnomon init`) is not a customization: the current bundled version is used instead, and the
// copy's path is returned in staleCopies so status/validate can say so.
func (l Layout) Workflows() (effective []WorkflowFile, staleCopies []string, err error) {
	byName := map[string]WorkflowFile{}

	entries, err := fs.ReadDir(bundled, path.Join(bundledRoot, "workflows"))
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		data, err := fs.ReadFile(bundled, path.Join(bundledRoot, "workflows", e.Name()))
		if err != nil {
			return nil, nil, err
		}
		byName[e.Name()] = WorkflowFile{Name: e.Name(), Data: data}
	}

	projectEntries, err := os.ReadDir(l.WorkflowsDir())
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}
	for _, e := range projectEntries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		p := filepath.Join(l.WorkflowsDir(), e.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, nil, err
		}
		if IsStock(e.Name(), data) {
			staleCopies = append(staleCopies, p)
			continue
		}
		byName[e.Name()] = WorkflowFile{Name: e.Name(), Data: data, ProjectPath: p}
	}

	for _, w := range byName {
		effective = append(effective, w)
	}
	sort.Slice(effective, func(i, j int) bool { return effective[i].Name < effective[j].Name })
	sort.Strings(staleCopies)
	return effective, staleCopies, nil
}

// Workflow returns the workflow in effect under name.
func (l Layout) Workflow(name string) (WorkflowFile, error) {
	all, _, err := l.Workflows()
	if err != nil {
		return WorkflowFile{}, err
	}
	for _, w := range all {
		if w.Name == name {
			return w, nil
		}
	}
	return WorkflowFile{}, fmt.Errorf("no workflow named %s (bundled or in %s)", name, l.WorkflowsDir())
}

// CustomizeWorkflow copies the bundled workflow name into .gnomon/workflows/ so the project can
// edit it. It refuses to overwrite an existing project file. The copy only takes effect once it
// differs from the shipped version; until then the bundled version keeps being used.
func (l Layout) CustomizeWorkflow(name string) (string, error) {
	data, err := fs.ReadFile(bundled, path.Join(bundledRoot, "workflows", name))
	if err != nil {
		return "", fmt.Errorf("no bundled workflow named %s", name)
	}
	if err := os.MkdirAll(l.WorkflowsDir(), 0o755); err != nil {
		return "", err
	}
	dest := filepath.Join(l.WorkflowsDir(), name)
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("%s already exists; edit it directly", dest)
	}
	return dest, os.WriteFile(dest, data, 0o644)
}

// knowledgeSections are the .gnomon/ directories that may hold project-written knowledge.
var knowledgeSections = []string{"context", "decisions", "contracts", "evaluations"}

// KnowledgeFiles returns, relative to the project root and sorted, every markdown file under the
// knowledge sections that the project actually wrote or changed. Unmodified Gnomon templates are
// excluded, so an agent is never handed an empty template as if it were project knowledge.
func (l Layout) KnowledgeFiles() ([]string, error) {
	var files []string
	for _, section := range knowledgeSections {
		dir := filepath.Join(l.GnomonRoot(), section)
		entries, err := os.ReadDir(dir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			p := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			if IsStock(e.Name(), data) || strings.TrimSpace(string(data)) == "" {
				continue
			}
			rel, err := filepath.Rel(l.Root, p)
			if err != nil {
				rel = p
			}
			files = append(files, filepath.ToSlash(rel))
		}
	}
	sort.Strings(files)
	return files, nil
}

// legacySpecTemplate is the template filename older `gnomon init` versions copied into
// .gnomon/specifications/. A project that edited it keeps using its own version.
const legacySpecTemplate = "SPEC-000-use-case-name.md"

// SpecTemplate returns the template a new Specification is created from: the bundled detailed
// template when detailed is set; otherwise a customized project template, if the project has one,
// or the short bundled default.
func (l Layout) SpecTemplate(detailed bool) ([]byte, error) {
	if detailed {
		return fs.ReadFile(bundled, path.Join(bundledRoot, "templates", "spec-detailed.md"))
	}
	if data, err := os.ReadFile(filepath.Join(l.SpecificationsDir(), legacySpecTemplate)); err == nil && !IsStock(legacySpecTemplate, data) {
		return data, nil
	}
	return fs.ReadFile(bundled, path.Join(bundledRoot, "templates", "spec.md"))
}

// SpecTemplateCandidates returns every template a Specification in this project may have been
// created from — the bundled short and detailed templates, plus the project's own legacy template
// file if present — for deciding whether a Specification still reads like an untouched template.
func (l Layout) SpecTemplateCandidates() [][]byte {
	var out [][]byte
	for _, name := range []string{"spec.md", "spec-detailed.md"} {
		if data, err := fs.ReadFile(bundled, path.Join(bundledRoot, "templates", name)); err == nil {
			out = append(out, data)
		}
	}
	if data, err := os.ReadFile(filepath.Join(l.SpecificationsDir(), legacySpecTemplate)); err == nil {
		out = append(out, data)
	}
	return out
}
