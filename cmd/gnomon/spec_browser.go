package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"gnomon/internal/orchestrate"
	"gnomon/internal/present"
)

const (
	actionCreate   = "__create__"
	actionDiscover = "__discover__"
	actionBack     = "__back__"
)

// runSpecBrowser is `gnomon spec` with no arguments and a real interactive terminal: list every
// Specification, let the Human search/filter and navigate them, and offer Create/Discover as
// pinned actions alongside the list. Opening a Specification hands off to runSpecWorkspace; when
// that returns, the browser redraws itself with freshly re-derived state — nothing about
// availability or the list is ever remembered across the loop.
func runSpecBrowser(root string) error {
	for {
		summaries, err := orchestrate.ListSpecsForRoot(root)
		if err != nil {
			fmt.Print(present.RenderErrorColor(err, colorEnabled()))
			return err
		}

		extra := []filterItem{
			{value: actionCreate, label: "+ Create new Specification", searchText: "create"},
			{value: actionDiscover, label: "+ Discover next Specification", searchText: "discover"},
		}
		items := make([]filterItem, len(summaries))
		for i, s := range summaries {
			label := fmt.Sprintf("%-10s %-30s %s", s.ID, truncate(s.Title, 30), s.Lifecycle)
			items[i] = filterItem{value: s.ID, label: label, searchText: s.ID + " " + s.Title}
		}

		title := fmt.Sprintf("Specifications (%d)", len(summaries))
		if len(summaries) == 0 {
			title = "Specifications (none yet)"
		}
		choice, err := runFilterableMenu(title, extra, items, 12)
		if errors.Is(err, errMenuCancelled) {
			return nil
		}
		if err != nil {
			fmt.Print(present.RenderErrorColor(err, colorEnabled()))
			return err
		}

		switch choice {
		case actionCreate:
			if err := runSpecCreateInteractive(root); err != nil && !errors.Is(err, errMenuCancelled) {
				fmt.Print(present.RenderErrorColor(err, colorEnabled()))
			}
		case actionDiscover:
			if err := runSpecDiscoverInteractive(root); err != nil && !errors.Is(err, errMenuCancelled) {
				fmt.Print(present.RenderErrorColor(err, colorEnabled()))
			}
		default:
			if err := runSpecWorkspace(root, choice); err != nil {
				fmt.Print(present.RenderErrorColor(err, colorEnabled()))
			}
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

// runSpecCreateInteractive asks for a title (a plain line prompt — text input, not a menu, so it
// reuses stdinPrompt rather than a raw-mode widget) and, once given, performs the same
// deterministic orchestrate.SpecCreate every gnomon spec create <title> call uses, then opens the
// new Specification's workspace directly.
func runSpecCreateInteractive(root string) error {
	title, err := stdinPrompt("Title for the new Specification (blank to cancel): ")
	if err != nil {
		return err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return errMenuCancelled
	}
	report, err := orchestrate.SpecCreate(root, title)
	if err != nil {
		fmt.Print(present.RenderErrorColor(err, colorEnabled()))
		return nil
	}
	fmt.Print(present.RenderWithOptions(report, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))
	return runSpecWorkspace(root, report.Target)
}

// runSpecDiscoverInteractive is Discovery's continuous interaction: run the workflow, present
// whatever candidate comes back, and require the Human's explicit accept or cancel before
// anything is created — the Agent itself never creates the Specification.
func runSpecDiscoverInteractive(root string) error {
	fmt.Fprintln(os.Stderr, "Running Specification Discovery…")
	candidate, rep, err := orchestrate.DiscoverCandidate(root, "", agentChooserForInvocation())
	if rep != nil {
		fmt.Print(present.RenderWithOptions(rep, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))
	}
	if candidate == nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\nCandidate: %s\n", candidate.Title)
	if candidate.Rationale != "" {
		fmt.Fprintf(os.Stderr, "Rationale: %s\n", candidate.Rationale)
	}
	if candidate.DerivedFrom != "" {
		fmt.Fprintf(os.Stderr, "Derived from: %s\n", candidate.DerivedFrom)
	}
	if candidate.NonBlockingCaveat != "" {
		fmt.Fprintf(os.Stderr, "Note: %s\n", candidate.NonBlockingCaveat)
	}
	fmt.Fprintln(os.Stderr)

	choice, err := runSelectMenu("Create this Specification?", []selectItem{
		{value: "create", label: "Yes, create it"},
		{value: "cancel", label: "No, cancel"},
	})
	if err != nil || choice != "create" {
		return errMenuCancelled
	}

	report, err := orchestrate.AcceptDiscoveryCandidate(root, candidate)
	if err != nil {
		fmt.Print(present.RenderErrorColor(err, colorEnabled()))
		return nil
	}
	fmt.Print(present.RenderWithOptions(report, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))
	return runSpecWorkspace(root, report.Target)
}

// runSpecWorkspace is `gnomon spec <SPEC-id>`: print the derived state, then offer only the
// currently-available actions as an interactive menu, re-deriving everything fresh after each one
// completes. Unavailable actions are shown, with their reason, in the printed detail above the
// menu rather than as disabled menu entries.
func runSpecWorkspace(root, id string) error {
	for {
		detail, err := orchestrate.SpecDetailForRoot(root, id)
		if err != nil {
			fmt.Print(present.RenderErrorColor(err, colorEnabled()))
			return err
		}

		printSpecDetail(detail)

		var items []selectItem
		for _, a := range detail.Actions {
			if a.Available {
				items = append(items, selectItem{value: a.Verb, label: capitalize(a.Verb)})
			}
		}
		items = append(items, selectItem{value: actionBack, label: "Back"})

		choice, err := runSelectMenu(fmt.Sprintf("%s — choose an action", id), items)
		if err != nil {
			return nil // Esc/Ctrl+C backs out of the workspace, same as choosing Back
		}
		if choice == actionBack {
			return nil
		}

		if err := runSpecAction(root, id, choice); err != nil && !errors.Is(err, errMenuCancelled) {
			fmt.Print(present.RenderErrorColor(err, colorEnabled()))
		}
		fmt.Fprintln(os.Stderr, "\n(press Enter to continue)")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
}

func printSpecDetail(d *orchestrate.SpecDetail) {
	fmt.Printf("%s — %s\n", d.ID, d.Title)
	fmt.Printf("  Lifecycle:   %s\n", d.Lifecycle)
	fmt.Printf("  Fingerprint: %s\n", short(d.Fingerprint))
	if d.ActiveGrant != nil {
		fmt.Printf("  Approved by: %s\n", d.ActiveGrant.Approver)
		fmt.Printf("  Approved at: %s\n", d.ActiveGrant.ApprovedAt)
	}
	if len(d.Revisions) > 0 {
		fmt.Printf("  Revisions (%d, oldest first):\n", len(d.Revisions))
		for i, r := range d.Revisions {
			status := "superseded" // never revoked, but content has since moved on
			switch {
			case r.Revoked:
				status = "revoked"
			case d.ActiveGrant != nil && d.ActiveGrant.GrantID == r.GrantID:
				status = "current"
			}
			fmt.Printf("    %d. %s by %s at %s (%s)\n", i+1, short(r.Fingerprint), r.Approver, r.ApprovedAt, status)
		}
	}
	var unavailable []string
	for _, a := range d.Actions {
		if !a.Available {
			unavailable = append(unavailable, fmt.Sprintf("%s (%s)", a.Verb, a.Reason))
		}
	}
	if len(unavailable) > 0 {
		fmt.Printf("  Not currently available: %s\n", strings.Join(unavailable, "; "))
	}
	if len(d.Unreadable) > 0 {
		fmt.Printf("  Warning: %s could not be loaded; run `gnomon validate`.\n", strings.Join(d.Unreadable, ", "))
	}
	fmt.Println()
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func short(fingerprint string) string {
	if len(fingerprint) > 12 {
		return fingerprint[:12]
	}
	return fingerprint
}

// runSpecAction dispatches one selected contextual action to the exact same Core operation the
// removed dedicated commands used to call directly — the workspace is a presentation layer over
// the same application logic, never a second implementation of it. Agent-invoking actions pass no
// --agent override: gnomon spec itself registers no such flag (there is no per-invocation CLI
// flag to read inside an interactive session), so resolution falls through to the persisted
// default or, on first use, the same interactive chooser every other Agent-invoking path uses.
func runSpecAction(root, id, verb string) error {
	switch verb {
	case "define":
		rep, err := orchestrate.SpecDefine(root, id, "", agentChooserForInvocation())
		return renderInteractive(rep, err)
	case "approve":
		rep, err := orchestrate.Approve(root, id, stdinPrompt)
		return renderInteractive(rep, err)
	case "revoke":
		rep, err := orchestrate.Revoke(root, id, stdinPrompt)
		return renderInteractive(rep, err)
	case "implement":
		rep, err := orchestrate.Implement(root, id, "", agentChooserForInvocation())
		return renderInteractive(rep, err)
	case "test":
		rep, err := orchestrate.Test(root, id, "", agentChooserForInvocation())
		return renderInteractive(rep, err)
	default:
		return fmt.Errorf("unknown action %q", verb)
	}
}

func renderInteractive(rep *present.Report, err error) error {
	if rep == nil {
		return err
	}
	fmt.Print(present.RenderWithOptions(rep, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))
	return nil
}
