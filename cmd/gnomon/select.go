package main

import (
	"fmt"
	"os"

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
