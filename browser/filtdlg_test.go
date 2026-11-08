package browser

import (
	"strings"
	"testing"
	"testing/fstest"

	tea "charm.land/bubbletea/v2"
)

func TestFilterDialog(t *testing.T) {
	mapFS := fstest.MapFS{
		"main.go":    {Data: []byte("package main")},
		"readme.txt": {Data: []byte("hello")},
	}

	bm, err := New(mapFS, "/", ".")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	// Select both files
	m.Data.Select(0, true)
	m.Data.Select(1, true)

	// Open FilterDialog
	dlg := NewFilterDialogModel(m)
	v := dlg.View()
	if !strings.Contains(v.Content, "Filter Files") {
		t.Errorf("filter dialog view missing title:\n%s", v.Content)
	}

	// Type "*.go" and press Enter
	mod, _ := dlg.Update(tea.KeyPressMsg{Code: '*', Text: "*.go"})
	dlg = mod.(FilterDialogModel)
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := mod.(Model)

	if updated.Data.Filter() != "*.go" {
		t.Errorf("expected filter *.go, got %q", updated.Data.Filter())
	}
	if updated.HeaderTail != "*.go" {
		t.Errorf("expected HeaderTail *.go, got %q", updated.HeaderTail)
	}
	if updated.Data.Len(80) != 1 || updated.Data.At(0).Name() != "main.go" {
		t.Fatalf("expected only main.go visible, got len=%d", updated.Data.Len(80))
	}
	if !updated.Data.Selected(0) {
		t.Errorf("expected main.go to remain selected")
	}

	// Clear filter via FilterDialog (Delete/Backspace or empty) and verify readme.txt was deselected
	updated.Data.SetFilter("")
	for i := 0; i < updated.Data.Len(80); i++ {
		name := updated.Data.At(i).Name()
		if name == "readme.txt" && updated.Data.Selected(i) {
			t.Errorf("expected hidden file readme.txt to be deselected after filtering")
		}
	}
}

