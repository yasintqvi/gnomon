package orchestrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// DefineOptions are `gnomon define`'s options for a new request.
type DefineOptions struct {
	Title    string // the new Specification's title; derived from the request when empty
	Detailed bool   // use the detailed template
}

// specIDArgPattern is how `gnomon define` tells an existing Specification from a new request.
var specIDArgPattern = regexp.MustCompile(`(?i)^SPEC-\d+$`)

// requestLinePrefix starts the block a new Draft keeps its request in, so an interrupted Define
// can be continued and a repeated request recognized.
const requestLinePrefix = specs.RequestLinePrefix

// maxTitleRunes bounds a title derived from a request.
const maxTitleRunes = 60

// Define performs `gnomon define`:
//
//   - `gnomon define SPEC-004 [change]` runs Define on that existing Specification, handing the
//     Agent the change text when given.
//   - `gnomon define "<request>"` starts a new feature: it creates a Draft (or continues the Draft
//     already started from the same request), saves the request in it, and runs Define on it.
//
// A new Draft is never created while another run is active, and a request whose title matches a
// Specification that has been approved is refused rather than duplicated.
func Define(root string, args []string, opts DefineOptions, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	if len(args) == 0 || strings.TrimSpace(strings.Join(args, " ")) == "" {
		return nil, fmt.Errorf(`describe the new feature (gnomon define "Members can reserve a tool"), or name the Specification to define (gnomon define SPEC-004)`)
	}
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}

	if specIDArgPattern.MatchString(args[0]) {
		if opts.Title != "" || opts.Detailed {
			return nil, fmt.Errorf("--title and --detailed apply only when starting a new Specification from a request")
		}
		id := strings.ToUpper(args[0])
		ok, _, err := specs.Exists(l.SpecificationsDir(), specs.Identity(id))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf(`%s does not exist. To start a new feature, describe it instead: gnomon define "<what you want>"`, id)
		}
		return runDefine(root, id, strings.TrimSpace(strings.Join(args[1:], " ")), agentOverride, chooser, "")
	}

	request := strings.TrimSpace(strings.Join(args, " "))
	if active := probeRunLock(root); active != nil {
		return runActiveReport("", active), active
	}

	title := strings.TrimSpace(opts.Title)
	if title == "" {
		title = titleFromRequest(request)
	}
	slug := specs.Slugify(title)
	if slug == "" {
		return nil, fmt.Errorf("could not derive a title from the request; pass --title")
	}

	id, note, err := draftForRequest(root, l, title, slug, request, opts.Detailed)
	if err != nil {
		return nil, err
	}
	return runDefine(root, id, request, agentOverride, chooser, note)
}

// draftForRequest returns the Draft to define for a new request: the existing Draft with the same
// title when it was started from this request (or from no request), otherwise a new one.
func draftForRequest(root string, l project.Layout, title, slug, request string, detailed bool) (id, note string, err error) {
	matches, _ := filepath.Glob(filepath.Join(l.SpecificationsDir(), "SPEC-*-"+slug+".md"))
	for _, path := range matches {
		existing := specIdentityOf(filepath.Base(path))
		if existing == "" {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "", "", err
		}
		existingTitle := specs.Title(specs.Identity(existing), content)
		if latestApprovedRevision(l.ApprovalsDir(), existing) != nil {
			return "", "", fmt.Errorf("%s — %s already exists and has been approved. To change it: gnomon define %s \"<the change>\". For a separate feature, pass a different --title",
				existing, existingTitle, existing)
		}
		saved := savedRequest(string(content))
		if saved != "" && saved != request {
			return "", "", fmt.Errorf("%s — %s is a Draft started from a different request (%q). Continue it with: gnomon define %s; for a separate feature, pass a different --title",
				existing, existingTitle, saved, existing)
		}
		return existing, fmt.Sprintf("Continuing %s — %s, the Draft already started for this request: %s", existing, existingTitle, relSlash(l.Root, path)), nil
	}

	rep, err := createDraftSpec(root, title, slug, detailed)
	if err != nil {
		return "", "", err
	}
	id = rep.Target
	_, path, err := specs.Exists(l.SpecificationsDir(), specs.Identity(id))
	if err != nil {
		return "", "", err
	}
	if err := saveRequestInDraft(path, request); err != nil {
		return "", "", err
	}
	return id, fmt.Sprintf("Created %s — %s (Draft), with your request saved in it: %s", id, title, relSlash(l.Root, path)), nil
}

// runDefine runs Define on id, handing over request, and reports which Specification it was. A
// run that does not finish ready for approval leaves the Draft and says how to continue.
func runDefine(root, id, request, agentOverride string, chooser AgentChooser, note string) (*present.Report, error) {
	rep, err := runTargetedWorkflow(root, "specification-definition.md",
		runRequest{kind: targetSpec, target: id, request: request, agentOverride: agentOverride, chooser: chooser})
	var active *RunActiveError
	if errors.As(err, &active) {
		return rep, err
	}
	if rep == nil {
		rep = &present.Report{Outcome: present.Failed, Summary: "Define could not run"}
		if err != nil {
			rep.AddSection("Reason", err.Error())
		}
	}
	rep.Target = id
	if note != "" {
		rep.Sections = append([]present.Section{{Label: "Specification", Body: note}}, rep.Sections...)
	}
	if rep.Outcome != present.Success {
		rep.Next = fmt.Sprintf("%s stays a Draft%s. Continue it with: gnomon define %s",
			id, map[bool]string{true: " with your request saved in it", false: ""}[note != ""], id)
	}
	return rep, err
}

// titleFromRequest is the request's first sentence, shortened at a word boundary.
func titleFromRequest(request string) string {
	line := strings.TrimSpace(strings.SplitN(request, "\n", 2)[0])
	for _, end := range []string{". ", "? ", "! ", ": "} {
		if i := strings.Index(line, end); i > 0 {
			line = line[:i]
		}
	}
	line = strings.TrimRight(line, ".?!: ")
	if utf8.RuneCountInString(line) > maxTitleRunes {
		runes := []rune(line)
		cut := string(runes[:maxTitleRunes])
		if runes[maxTitleRunes] != ' ' { // cut mid-word: drop the partial word
			if i := strings.LastIndex(cut, " "); i > maxTitleRunes/2 {
				cut = cut[:i]
			}
		}
		line = strings.TrimRight(cut, " ,;")
	}
	return line
}

// saveRequestInDraft adds the request below the Draft's title line.
func saveRequestInDraft(path, request string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(content)
	quoted := requestLinePrefix + strings.ReplaceAll(strings.TrimSpace(request), "\n", "\n> ")
	if i := strings.Index(text, "\n"); i >= 0 {
		text = text[:i+1] + "\n" + quoted + "\n" + text[i+1:]
	} else {
		text += "\n\n" + quoted + "\n"
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

// savedRequest reads the request a Draft was started from, "" when it has none.
func savedRequest(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, requestLinePrefix) {
			continue
		}
		parts := []string{strings.TrimPrefix(line, requestLinePrefix)}
		for _, more := range lines[i+1:] {
			if !strings.HasPrefix(more, "> ") {
				break
			}
			parts = append(parts, strings.TrimPrefix(more, "> "))
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	}
	return ""
}
