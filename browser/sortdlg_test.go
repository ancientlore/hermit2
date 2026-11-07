package browser

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ancientlore/hermit2/views"
	tea "charm.land/bubbletea/v2"
)

func TestSortDialog(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)

	mapFS := fstest.MapFS{
		"small.txt": {Data: []byte("a"), ModTime: t2},
		"large.go":  {Data: []byte("abcdef"), ModTime: t1},
	}

	bm, err := New(mapFS, "/", ".")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	dlg := NewSortDialogModel(m)
	v := dlg.View()
	if !strings.Contains(v.Content, "Sorting") || !strings.Contains(v.Content, "Sort by:") {
		t.Errorf("dialog view missing expected text:\n%s", v.Content)
	}

	// Press 's' (Size) and 'r' (Reverse), then Enter
	mod, _ := dlg.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	dlg = mod.(SortDialogModel)
	if dlg.sortType != views.SortBySize {
		t.Errorf("expected SortBySize, got %v", dlg.sortType)
	}

	mod, _ = dlg.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	dlg = mod.(SortDialogModel)
	if dlg.sortDir != views.SortReverse {
		t.Errorf("expected SortReverse, got %v", dlg.sortDir)
	}

	mod, _ = dlg.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updatedBrowser, ok := mod.(Model)
	if !ok {
		t.Fatalf("expected Model returned on Enter, got %T", mod)
	}
	if updatedBrowser.Data.SortType() != views.SortBySize || updatedBrowser.Data.SortDirection() != views.SortReverse {
		t.Errorf("browser sort not updated: type=%v dir=%v", updatedBrowser.Data.SortType(), updatedBrowser.Data.SortDirection())
	}
	if got := updatedBrowser.Data.At(0).Name(); got != "large.go" {
		t.Errorf("expected first file after reverse size sort to be large.go, got %s", got)
	}

	// Open dialog again, press 'n' (Name) and Esc -> should cancel without changing sort
	dlg2 := NewSortDialogModel(updatedBrowser)
	mod, _ = dlg2.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
	dlg2 = mod.(SortDialogModel)
	mod, _ = dlg2.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	cancelledBrowser := mod.(Model)
	if cancelledBrowser.Data.SortType() != views.SortBySize {
		t.Errorf("expected SortBySize preserved after Esc, got %v", cancelledBrowser.Data.SortType())
	}

	// Reset default sort
	cancelledBrowser.Data.SetSort(views.SortByName, views.SortForward)
}

