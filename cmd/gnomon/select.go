package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// selectItem pairs a value returned to the caller with the label shown to the Human.
type selectItem struct {
	value string
	label string
}

// moveSelection computes the next selected index for a delta of -1 (up) or +1 (down) over count
// items, wrapping at both ends. Pulled out of runSelectMenu as its own pure function specifically
// so this one piece of real logic is testable without a terminal.
func moveSelection(current, delta, count int) int {
	return ((current+delta)%count + count) % count
}

// runSelectMenu renders a minimal arrow-key selection menu to stderr (keeping stdout clean) and
// returns the chosen item's value once the Human presses Enter. It requires stdin to be a real
// terminal — callers must only invoke this once interactivity has already been confirmed
// (agentChooserForInvocation does this before ever wiring a chooser in).
//
// This is a small, self-contained widget built directly on golang.org/x/term's raw-mode support
// and plain ANSI cursor-movement codes — not a general terminal UI framework — because arrow-key
// navigation genuinely requires reading individual keystrokes before Enter, which only raw mode
// provides; nothing else about the CLI's presentation changes as a result of adding it.
func runSelectMenu(title string, items []selectItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no choices available")
	}

	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	selected := 0
	linesPerDraw := len(items) + 4 // title, blank, one line per item, blank, footer

	draw := func(redraw bool) {
		if redraw {
			fmt.Fprintf(os.Stderr, "\033[%dA", linesPerDraw)
		}
		fmt.Fprintf(os.Stderr, "\033[2K? %s\r\n", title)
		fmt.Fprint(os.Stderr, "\033[2K\r\n")
		for i, it := range items {
			fmt.Fprint(os.Stderr, "\033[2K")
			if i == selected {
				fmt.Fprintf(os.Stderr, "  ❯ %s\r\n", it.label)
			} else {
				fmt.Fprintf(os.Stderr, "    %s\r\n", it.label)
			}
		}
		fmt.Fprint(os.Stderr, "\033[2K\r\n")
		fmt.Fprint(os.Stderr, "\033[2K↑/↓ navigate • Enter select\r\n")
	}

	draw(false)

	one := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(one); err != nil {
			return "", err
		}
		switch one[0] {
		case '\r', '\n':
			fmt.Fprint(os.Stderr, "\r\n")
			return items[selected].value, nil
		case 3, 4: // Ctrl+C, Ctrl+D
			fmt.Fprint(os.Stderr, "\r\n")
			return "", fmt.Errorf("selection cancelled")
		case 0x1b: // ESC — the start of an arrow-key escape sequence
			seq := make([]byte, 2)
			if _, err := os.Stdin.Read(seq); err != nil {
				return "", err
			}
			if seq[0] != '[' {
				continue
			}
			switch seq[1] {
			case 'A': // up
				selected = moveSelection(selected, -1, len(items))
				draw(true)
			case 'B': // down
				selected = moveSelection(selected, 1, len(items))
				draw(true)
			}
		}
	}
}

// filterItem is one row in a filterable menu: value returned on selection, label displayed, and
// searchText matched against the typed filter (lowercased once up front) — kept distinct from
// label so a row can be found by more than just what's visually shown (e.g. matching a
// Specification's identity even though the label leads with its title).
type filterItem struct {
	value      string
	label      string
	searchText string
}

// errMenuCancelled is returned by runFilterableMenu (and may be returned by callers of
// runSelectMenu) when the Human explicitly backs out — Esc or Ctrl+C/Ctrl+D — as opposed to a
// real error. Callers distinguish the two with errors.Is.
var errMenuCancelled = fmt.Errorf("cancelled")

// runFilterableMenu renders a scrollable, type-to-filter menu to stderr and returns the selected
// item's value, errMenuCancelled if the Human backed out, or a real error. extraActions, when
// non-empty, are always shown first, unaffected by the filter (used for "+ Create new
// Specification" / "+ Discover next Specification" style entries that should stay reachable
// regardless of what's typed).
//
// The visible window is a fixed maxVisible rows regardless of how many items currently match, so
// every redraw clears and repaints exactly the same number of terminal lines — the same fixed-
// height redraw strategy runSelectMenu already uses, just with the row content varying instead of
// the row count.
func runFilterableMenu(title string, extraActions, items []filterItem, maxVisible int) (string, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, oldState)

	var filter string
	selected := 0
	linesPerDraw := 4 + maxVisible // title, filter line, blank, maxVisible rows, footer

	visible := func() []filterItem {
		q := strings.ToLower(filter)
		var matched []filterItem
		matched = append(matched, extraActions...)
		for _, it := range items {
			if q == "" || strings.Contains(strings.ToLower(it.searchText), q) {
				matched = append(matched, it)
			}
		}
		return matched
	}

	draw := func(redraw bool) {
		if redraw {
			fmt.Fprintf(os.Stderr, "\033[%dA", linesPerDraw)
		}
		matched := visible()
		if selected >= len(matched) {
			selected = 0
		}
		fmt.Fprintf(os.Stderr, "\033[2K? %s\r\n", title)
		filterDisplay := filter
		if filterDisplay == "" {
			filterDisplay = "(type to filter)"
		}
		fmt.Fprintf(os.Stderr, "\033[2K  Filter: %s\r\n", filterDisplay)
		fmt.Fprint(os.Stderr, "\033[2K\r\n")
		for i := 0; i < maxVisible; i++ {
			fmt.Fprint(os.Stderr, "\033[2K")
			if i >= len(matched) {
				fmt.Fprint(os.Stderr, "\r\n")
				continue
			}
			marker := "   "
			if i == selected {
				marker = " ❯ "
			}
			fmt.Fprintf(os.Stderr, "%s%s\r\n", marker, matched[i].label)
		}
		fmt.Fprint(os.Stderr, "\033[2K↑/↓ navigate • type to filter • Enter select • Esc cancel\r\n")
	}

	draw(false)

	one := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(one); err != nil {
			return "", err
		}
		switch {
		case one[0] == '\r' || one[0] == '\n':
			matched := visible()
			if len(matched) == 0 {
				continue
			}
			fmt.Fprint(os.Stderr, "\r\n")
			return matched[selected].value, nil
		case one[0] == 3 || one[0] == 4: // Ctrl+C, Ctrl+D
			fmt.Fprint(os.Stderr, "\r\n")
			return "", errMenuCancelled
		case one[0] == 0x1b: // ESC, or the start of an arrow-key escape sequence
			seq := make([]byte, 2)
			n, _ := os.Stdin.Read(seq)
			if n < 2 || seq[0] != '[' {
				fmt.Fprint(os.Stderr, "\r\n")
				return "", errMenuCancelled
			}
			switch seq[1] {
			case 'A':
				count := len(visible())
				if count > 0 {
					selected = moveSelection(selected, -1, count)
				}
				draw(true)
			case 'B':
				count := len(visible())
				if count > 0 {
					selected = moveSelection(selected, 1, count)
				}
				draw(true)
			}
		case one[0] == 0x7f || one[0] == 0x08: // Backspace
			if len(filter) > 0 {
				filter = filter[:len(filter)-1]
				selected = 0
				draw(true)
			}
		case one[0] >= 0x20 && one[0] < 0x7f: // printable ASCII
			filter += string(one[0])
			selected = 0
			draw(true)
		}
	}
}
