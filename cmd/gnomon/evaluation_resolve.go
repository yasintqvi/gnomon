package main

import (
	"fmt"
	"os"
	"strings"

	"gnomon/internal/orchestrate"
	"gnomon/internal/present"
)

const (
	evalResolveAll      = "resolve-all"
	evalResolveSelected = "resolve-selected"
	evalExit            = "exit"
)

// runEvaluationResolution is the interactive post-result loop for `gnomon run verification|review
// <target>`, entered only when stdin is a real terminal and the run produced at least one
// actionable finding (run.go's own gate).
//
// Responsibility model: the CLI orchestrates (present the Agent's recommendation, collect
// authorization, validate eligibility, launch) and the Agent reasons and resolves — the CLI is not
// part of that conversation once a Resolution Agent is running. The Human confirms the recommended
// workflow, picks a different one, or skips, one finding at a time; only once every finding chosen
// for this round has been handled does Gnomon re-run Verification/Review fresh — a resolver never
// declares its own finding resolved. No finding state is ever written anywhere.
func runEvaluationResolution(root, identity, target string, report *present.Report, findings []orchestrate.EvaluationFinding, runErr error) error {
	for {
		fmt.Print(present.RenderWithOptions(report, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))

		if len(findings) == 0 {
			return finishEvaluation(runErr)
		}

		p := present.NewPalette(colorEnabled())
		fmt.Print(renderFindingList(identity, findings, p))

		choice, err := runSelectMenu("What would you like to do?", []selectItem{
			{value: evalResolveAll, label: "Walk through resolving all findings"},
			{value: evalResolveSelected, label: "Walk through resolving selected findings"},
			{value: evalExit, label: "Exit"},
		})
		if err != nil || choice == evalExit {
			// Ctrl+C/Ctrl+D/a read error are all treated as Exit, same as runSpecWorkspace's menu.
			return finishEvaluation(runErr)
		}

		var toWalk []orchestrate.EvaluationFinding
		switch choice {
		case evalResolveAll:
			toWalk = findings
		case evalResolveSelected:
			toWalk, err = promptFindingSelection(findings)
			if err != nil {
				continue // cancelled — back to the top of the loop, same findings, same report
			}
		}

		dispatched := false
		for _, f := range toWalk {
			ok, err := authorizeAndResolveFinding(root, identity, target, f)
			if err != nil {
				return err
			}
			if ok {
				dispatched = true
			}
		}
		if !dispatched {
			// Nothing ran (skipped, or refused by eligibility) — re-show the same findings rather
			// than re-evaluating what's already known.
			continue
		}

		fmt.Fprintf(os.Stderr, "\nRe-running %s to confirm...\n", identity)
		report, findings, runErr = orchestrate.RunEvaluation(root, identity, target, *runAgentFlag, agentChooserForInvocation())
		if report == nil {
			return runErr
		}
	}
}

// finishEvaluation mirrors renderReport's own exit-code convention (root.go). Called only after
// the relevant Report has already been printed.
func finishEvaluation(err error) error {
	if err != nil {
		os.Exit(1)
	}
	return nil
}

// authorizeAndResolveFinding walks the Human through exactly one finding — show it, require an
// explicit choice (confirm the recommendation, pick a different workflow, or skip) — then, only on
// confirmation, dispatches, carrying the finding as transient Handoff context. Returns
// dispatched=true only when a resolution Agent invocation actually ran.
func authorizeAndResolveFinding(root, identity, target string, f orchestrate.EvaluationFinding) (dispatched bool, err error) {
	p := present.NewPalette(colorEnabled())
	fmt.Print(renderFindingDetail(f, p))

	choice, err := runSelectMenu("What would you like to do?", findingMenuItems(f))
	if err != nil || choice == "skip" {
		return false, nil
	}

	approach := f.RecommendedWorkflow
	if choice == "choose" {
		approach, err = promptResolutionApproach()
		if err != nil {
			return false, nil // "Back" — treated the same as Skip for this finding
		}
	}

	return dispatchResolution(root, identity, target, f, approach)
}

// findingMenuItems builds the per-finding menu. "Continue with <X>" appears only when the
// recommendation is a real, validated, dispatchable workflow; "Choose another workflow" and "Skip"
// are always offered.
func findingMenuItems(f orchestrate.EvaluationFinding) []selectItem {
	var items []selectItem
	if f.RecommendedWorkflow != "" && orchestrate.IsDispatchableResolution(f.RecommendedWorkflow) {
		items = append(items, selectItem{value: "recommended", label: "Continue with " + resolutionLabel(f.RecommendedWorkflow)})
	}
	items = append(items, selectItem{value: "choose", label: "Choose another workflow"})
	items = append(items, selectItem{value: "skip", label: "Skip"})
	return items
}

// promptResolutionApproach is the fallback/override path, offered when the Human declines or has
// no Agent recommendation to confirm.
func promptResolutionApproach() (string, error) {
	choice, err := runSelectMenu("Choose a resolution workflow", []selectItem{
		{value: orchestrate.ResolveWithImplementation, label: "Implementation"},
		{value: orchestrate.ResolveWithTesting, label: "Testing"},
		{value: orchestrate.ResolveWithKnowledgeResolution, label: "Knowledge Resolution"},
		{value: "back", label: "Back"},
	})
	if err != nil || choice == "back" {
		return "", errMenuCancelled
	}
	return choice, nil
}

// dispatchResolution collects whatever this workflow's own normal target requires (an Approved
// Specification for Implementation, optional for Testing, none for Knowledge Resolution), then
// runs it through orchestrate.RunResolutionWorkflow, which enforces eligibility unmodified. Gnomon
// never infers, creates, or approves a Specification here.
func dispatchResolution(root, identity, target string, f orchestrate.EvaluationFinding, approach string) (bool, error) {
	if !orchestrate.IsDispatchableResolution(approach) {
		return false, nil
	}

	var specID string
	switch approach {
	case orchestrate.ResolveWithImplementation:
		answer, perr := stdinPrompt("Implementation requires an Approved Specification.\nWhich Specification governs this change? (SPEC-id): ")
		if perr != nil {
			return false, perr
		}
		specID = strings.TrimSpace(answer)
		if specID == "" {
			fmt.Fprintln(os.Stderr, "No Specification given — Implementation was not started.")
			return false, nil
		}
	case orchestrate.ResolveWithTesting:
		answer, perr := stdinPrompt("Which Specification does this relate to? (optional, blank for ad-hoc testing): ")
		if perr != nil {
			return false, perr
		}
		specID = strings.TrimSpace(answer)
	}

	handoff := orchestrate.ResolutionHandoff{
		OriginWorkflow: identity,
		OriginTarget:   target,
		FindingID:      f.ID,
		Classification: f.Classification,
		Summary:        f.Headline,
		Evidence:       f.Detail,
	}

	rep, runErr := orchestrate.RunResolutionWorkflow(root, approach, specID, handoff, "", agentChooserForInvocation())
	if rep == nil {
		return false, runErr
	}
	fmt.Print(present.RenderWithOptions(rep, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))
	if runErr != nil {
		// The chosen workflow refused or didn't complete — its own Report already explains why.
		return false, nil
	}
	return true, nil
}

// renderFindingList is the finding block printed between a fresh evaluation's Report and the
// resolution menu — pure and Palette-driven, so it's testable without a terminal.
func renderFindingList(identity string, findings []orchestrate.EvaluationFinding, p present.Palette) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(p.Heading(findingListHeader(identity, len(findings))))
	b.WriteString("\n\n")
	for _, f := range findings {
		b.WriteString(p.Code(f.ID))
		b.WriteString("  ")
		b.WriteString(classificationStyle(p, f.Classification))
		b.WriteString("\n")
		if f.Headline != "" {
			b.WriteString("  ")
			b.WriteString(f.Headline)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderFindingDetail is the fuller, single-finding view shown before authorizeAndResolveFinding's
// menu. "Decision required" is shown only when the result actually carried that information
// (Review only) — purely informational, never used to choose a workflow.
func renderFindingDetail(f orchestrate.EvaluationFinding, p present.Palette) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(p.Code(f.ID))
	b.WriteString("  ")
	b.WriteString(classificationStyle(p, f.Classification))
	b.WriteString("\n\n")
	if f.Headline != "" {
		b.WriteString(f.Headline)
		b.WriteString("\n")
	}
	if f.Detail != "" {
		for _, line := range strings.Split(f.Detail, "\n") {
			b.WriteString("  ")
			b.WriteString(p.Muted(line))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	if f.RecommendedWorkflow != "" {
		b.WriteString(p.Heading("Recommended workflow:"))
		b.WriteString("\n  ")
		b.WriteString(resolutionLabel(f.RecommendedWorkflow))
		b.WriteString("\n\n")
	}
	if f.DecisionRequired {
		b.WriteString(p.Heading("Decision required:"))
		b.WriteString("\n  Yes\n\n")
	}
	return b.String()
}

func findingListHeader(identity string, n int) string {
	switch identity {
	case "verification":
		return fmt.Sprintf("Verification found %d unresolved %s", n, pluralize(n, "obligation", "obligations"))
	case "review":
		return fmt.Sprintf("Review found %d %s", n, pluralize(n, "finding", "findings"))
	default:
		return fmt.Sprintf("%s found %d %s", capitalize(identity), n, pluralize(n, "finding", "findings"))
	}
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// classificationStyle maps each classification onto Gnomon's usual vocabulary: DEFECT/FAIL/CONFLICT
// as failure (red), RISK/UNVERIFIABLE/KNOWLEDGE GAP as needs-attention (yellow).
func classificationStyle(p present.Palette, classification string) string {
	switch classification {
	case "DEFECT", "FAIL", "CONFLICT":
		return p.Bad(classification)
	default: // RISK, UNVERIFIABLE, KNOWLEDGE GAP
		return p.Warn(classification)
	}
}

// resolutionLabel is the Human-facing label for a dispatchable resolution approach — display
// only; dispatch itself always goes through orchestrate.RunResolutionWorkflow's own identity
// strings, never this label.
func resolutionLabel(approach string) string {
	switch approach {
	case orchestrate.ResolveWithImplementation:
		return "Implementation"
	case orchestrate.ResolveWithTesting:
		return "Testing"
	case orchestrate.ResolveWithKnowledgeResolution:
		return "Knowledge Resolution"
	default:
		return approach
	}
}

// parseFindingSelection interprets the Human's typed finding IDs against the currently offered
// findings — case-insensitive, whitespace-separated, order-preserving, duplicates collapsed.
func parseFindingSelection(answer string, findings []orchestrate.EvaluationFinding) (selected []orchestrate.EvaluationFinding, invalid []string) {
	byID := make(map[string]orchestrate.EvaluationFinding, len(findings))
	for _, f := range findings {
		byID[strings.ToUpper(f.ID)] = f
	}
	seen := map[string]bool{}
	for _, tok := range strings.Fields(answer) {
		key := strings.ToUpper(tok)
		f, ok := byID[key]
		if !ok {
			invalid = append(invalid, tok)
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		selected = append(selected, f)
	}
	return selected, invalid
}

// promptFindingSelection is parseFindingSelection's interactive wrapper: retries on invalid input,
// treats a blank answer as cancel back to the top-level finding menu.
func promptFindingSelection(findings []orchestrate.EvaluationFinding) ([]orchestrate.EvaluationFinding, error) {
	var ids []string
	for _, f := range findings {
		ids = append(ids, f.ID)
	}
	for {
		answer, err := stdinPrompt(fmt.Sprintf("Finding IDs (%s), blank to cancel: ", strings.Join(ids, " ")))
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(answer) == "" {
			return nil, errMenuCancelled
		}
		selected, invalid := parseFindingSelection(answer, findings)
		if len(invalid) > 0 {
			fmt.Fprintf(os.Stderr, "Unrecognized finding ID(s): %s. Try again.\n\n", strings.Join(invalid, ", "))
			continue
		}
		if len(selected) == 0 {
			fmt.Fprintln(os.Stderr, "No finding IDs entered. Try again.")
			continue
		}
		return selected, nil
	}
}
