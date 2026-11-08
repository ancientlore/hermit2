package browser

import (
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ancientlore/hermit2/config"
	"github.com/ancientlore/hermit2/views"
)

func TestOptionsPersistence(t *testing.T) {
	optsPath := filepath.Join(t.TempDir(), config.OptionsFileName)
	t.Cleanup(func() {
		_, _ = config.InitOptionsAt("")
		ApplyOptions(config.DefaultOptions())
	})

	opts, err := config.InitOptionsAt(optsPath)
	if err != nil {
		t.Fatalf("InitOptionsAt failed: %v", err)
	}
	ApplyOptions(opts)

	t1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)
	mapFS := fstest.MapFS{
		"small.txt":     {Data: []byte("a"), ModTime: t2},
		"large.go":      {Data: []byte("abcdef"), ModTime: t1},
		"sub/inside.go": {Data: []byte("package sub"), ModTime: t1},
	}

	bm, err := New(mapFS, "/", ".")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	// 1. Change sort to Extension + Reverse via SortDialogModel and press Enter
	sortDlg := NewSortDialogModel(m)
	mod, _ := sortDlg.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	sortDlg = mod.(SortDialogModel)
	mod, _ = sortDlg.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	sortDlg = mod.(SortDialogModel)
	mod, _ = sortDlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = mod.(Model)

	// 2. Apply filter "*.go" via FilterDialogModel and press Enter
	filtDlg := NewFilterDialogModel(m)
	mod, _ = filtDlg.Update(tea.KeyPressMsg{Code: '*', Text: "*.go"})
	filtDlg = mod.(FilterDialogModel)
	mod, _ = filtDlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = mod.(Model)

	// 2b. Navigate to "sub" via GotoDialogModel and press Enter
	gotoDlg := NewGotoDialogModel(m)
	mod, _ = gotoDlg.Update(tea.KeyPressMsg{Code: 's', Text: "sub"})
	gotoDlg = mod.(GotoDialogModel)
	mod, _ = gotoDlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = mod.(Model)

	// 3. Verify options.toml on disk reflects the updated sort, filter, filter history, and goto history
	persisted, err := config.LoadOptionsFrom(optsPath)
	if err != nil {
		t.Fatalf("LoadOptionsFrom failed: %v", err)
	}
	if persisted.SortBy != views.SortByExt || persisted.SortDirection != views.SortReverse {
		t.Errorf("unexpected persisted sort: by=%v dir=%v", persisted.SortBy, persisted.SortDirection)
	}
	if persisted.Filter != "*.go" {
		t.Errorf("expected persisted filter *.go, got %q", persisted.Filter)
	}
	if !slices.Equal(persisted.FilterHistory, []string{"*.go"}) {
		t.Errorf("expected persisted filter history [*.go], got %v", persisted.FilterHistory)
	}
	if !slices.Equal(persisted.GotoHistory, []string{"/sub"}) {
		t.Errorf("expected persisted goto history [/sub], got %v", persisted.GotoHistory)
	}

	// 4. Simulate a fresh startup by resetting in-memory defaults and re-applying persisted options
	ApplyOptions(config.DefaultOptions())
	ApplyOptions(persisted)

	reloadedBM, err := New(mapFS, "/", ".")
	if err != nil {
		t.Fatalf("New after reload failed: %v", err)
	}
	if reloadedBM.Data.SortType() != views.SortByExt || reloadedBM.Data.SortDirection() != views.SortReverse {
		t.Errorf("reloaded browser sort mismatch: by=%v dir=%v", reloadedBM.Data.SortType(), reloadedBM.Data.SortDirection())
	}
	if reloadedBM.Data.Filter() != "*.go" || reloadedBM.HeaderTail != "*.go" {
		t.Errorf("reloaded browser filter mismatch: filter=%q headerTail=%q", reloadedBM.Data.Filter(), reloadedBM.HeaderTail)
	}
	if reloadedBM.Data.Len(80) != 2 || reloadedBM.Data.At(0).Name() != "sub" || reloadedBM.Data.At(1).Name() != "large.go" {
		t.Errorf("expected sub and large.go visible after reload, got len=%d", reloadedBM.Data.Len(80))
	}
	if !slices.Equal(gotoHistory.Items(), []string{"/sub"}) {
		t.Errorf("expected reloaded gotoHistory [/sub], got %v", gotoHistory.Items())
	}
}

