package browser

import (
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	tea "charm.land/bubbletea/v2"
	"github.com/ancientlore/hermit2/config"
)

func TestGotoDialogAutocompleteAndDrillDown(t *testing.T) {
	t.Cleanup(func() {
		ApplyOptions(config.DefaultOptions())
	})
	ApplyOptions(config.DefaultOptions())

	mapFS := fstest.MapFS{
		"alpha/one/file.txt":   {Data: []byte("1")},
		"alpha/two/file.txt":   {Data: []byte("2")},
		"beta/three/file.txt":  {Data: []byte("3")},
		"gamma/file.txt":       {Data: []byte("4")},
		"rootfile.txt":         {Data: []byte("root")},
	}

	bm, err := New(mapFS, "/", ".")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	// Open GotoDialog via Alt+G
	dlgMod, _ := m.Update(tea.KeyPressMsg{Code: 'g', Mod: tea.ModAlt})
	dlg, ok := dlgMod.(GotoDialogModel)
	if !ok {
		t.Fatalf("expected GotoDialogModel on alt+g, got %T", dlgMod)
	}

	v := dlg.View()
	if !strings.Contains(v.Content, "Goto Location") {
		t.Errorf("goto dialog view missing title:\n%s", v.Content)
	}

	// Initially empty input should suggest top-level folders (alpha/, beta/, gamma/) and not rootfile.txt
	if len(dlg.suggestions) != 3 {
		t.Fatalf("expected 3 initial folder suggestions, got %d: %+v", len(dlg.suggestions), dlg.suggestions)
	}

	// Type "al" to filter down to "alpha/"
	mod, _ := dlg.Update(tea.KeyPressMsg{Code: 'a', Text: "al"})
	dlg = mod.(GotoDialogModel)
	if len(dlg.suggestions) != 1 || dlg.suggestions[0].display != "alpha/" {
		t.Fatalf("expected [alpha/] suggestion after typing 'al', got %+v", dlg.suggestions)
	}

	// Press Tab to drill into "alpha/"
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	dlg = mod.(GotoDialogModel)
	if dlg.edit.Value() != "alpha/" {
		t.Fatalf("expected edit value 'alpha/' after Tab, got %q", dlg.edit.Value())
	}
	if len(dlg.suggestions) != 2 || dlg.suggestions[0].display != "alpha/one/" || dlg.suggestions[1].display != "alpha/two/" {
		t.Fatalf("expected [alpha/one/, alpha/two/] suggestions after drilling into alpha/, got %+v", dlg.suggestions)
	}

	// Use Down arrow twice to select "alpha/two/" and press Enter to go there
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	dlg = mod.(GotoDialogModel)
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	dlg = mod.(GotoDialogModel)
	if dlg.selected != 1 {
		t.Fatalf("expected selected index 1, got %d", dlg.selected)
	}

	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated, ok := mod.(Model)
	if !ok {
		t.Fatalf("expected browser.Model after Enter, got %T", mod)
	}
	if updated.Data.Folder() != "/alpha/two" {
		t.Errorf("expected browser folder '/alpha/two', got %q", updated.Data.Folder())
	}

	// Verify "/alpha/two" was added to Goto History
	if !slices.Equal(gotoHistory.Items(), []string{"/alpha/two"}) {
		t.Errorf("expected gotoHistory [/alpha/two], got %v", gotoHistory.Items())
	}
}

func TestGotoDialogHistoryRecallAndDrill(t *testing.T) {
	t.Cleanup(func() {
		ApplyOptions(config.DefaultOptions())
	})
	ApplyOptions(config.Options{
		GotoHistory: []string{"/beta"},
	})

	mapFS := fstest.MapFS{
		"alpha/file.txt":      {Data: []byte("1")},
		"beta/three/file.txt": {Data: []byte("3")},
	}

	bm, err := New(mapFS, "/", "/alpha")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	dlg := NewGotoDialogModel(m)
	// First suggestion should be the history item "/beta"
	if len(dlg.suggestions) == 0 || !dlg.suggestions[0].isHistory || dlg.suggestions[0].targetPath != "/beta" {
		t.Fatalf("expected first suggestion to be history '/beta', got %+v", dlg.suggestions)
	}

	// Arrow Down to highlight "/beta", then press Right arrow to drill into "/beta/"
	mod, _ := dlg.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	dlg = mod.(GotoDialogModel)
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	dlg = mod.(GotoDialogModel)

	if dlg.edit.Value() != "/beta/" {
		t.Fatalf("expected edit value '/beta/' after drilling into history item, got %q", dlg.edit.Value())
	}
	if len(dlg.suggestions) != 1 || dlg.suggestions[0].display != "/beta/three/" {
		t.Fatalf("expected [/beta/three/] suggestion inside /beta/, got %+v", dlg.suggestions)
	}

	// Highlight "/beta/three/" and press Enter
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	dlg = mod.(GotoDialogModel)
	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := mod.(Model)

	if updated.Data.Folder() != "/beta/three" {
		t.Errorf("expected browser folder '/beta/three', got %q", updated.Data.Folder())
	}
}

