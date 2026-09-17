package main

import "testing"

func TestMoveSelection_DownWrapsToStart(t *testing.T) {
	if got := moveSelection(1, 1, 2); got != 0 {
		t.Fatalf("expected wrap from last item to 0, got %d", got)
	}
}

func TestMoveSelection_UpWrapsToEnd(t *testing.T) {
	if got := moveSelection(0, -1, 2); got != 1 {
		t.Fatalf("expected wrap from 0 to the last item, got %d", got)
	}
}

func TestMoveSelection_MidRangeMovesNormally(t *testing.T) {
	if got := moveSelection(1, 1, 3); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
	if got := moveSelection(1, -1, 3); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestMoveSelection_SingleItemStaysPut(t *testing.T) {
	if got := moveSelection(0, 1, 1); got != 0 {
		t.Fatalf("expected a single-item list to stay at 0, got %d", got)
	}
	if got := moveSelection(0, -1, 1); got != 0 {
		t.Fatalf("expected a single-item list to stay at 0, got %d", got)
	}
}

func TestRunSelectMenu_RefusesWithNoItems(t *testing.T) {
	if _, err := runSelectMenu("Choose:", nil); err == nil {
		t.Fatalf("expected an empty item list to be refused before touching the terminal at all")
	}
}
