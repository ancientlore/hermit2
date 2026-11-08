package browser

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ancientlore/hermit2/scroller"
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

func TestSymlinkNavigation(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	mapFS := fstest.MapFS{
		"target_dir":           {Mode: fs.ModeDir | 0755, ModTime: t1},
		"target_dir/hello.txt": {Data: []byte("hello from target"), ModTime: t1},
		"link_dir":             {Data: []byte("target_dir"), Mode: fs.ModeSymlink | 0777, ModTime: t1},
		"link_file.txt":        {Data: []byte("target_dir/hello.txt"), Mode: fs.ModeSymlink | 0777, ModTime: t1},
	}

	bm, err := New(mapFS, "/", "/root")
	// Note: testing from "/"
	bm, err = New(mapFS, "/", "/")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	bmMod, _ := bm.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m := bmMod.(Model)

	// Cursor 0 is "link_dir" (sorts first among directories)
	if got := m.Data.At(m.Cursor()).Name(); got != "link_dir" {
		t.Fatalf("expected cursor on link_dir, got %s", got)
	}

	// Press Right to traverse into link_dir
	mod, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	subModel, ok := mod.(Model)
	if !ok {
		t.Fatalf("expected browser Model when traversing symlink dir, got %T", mod)
	}
	// Logical path should be "/link_dir"
	if subModel.Data.Folder() != "/link_dir" {
		t.Errorf("expected logical folder /link_dir, got %s", subModel.Data.Folder())
	}
	if subModel.Header != filepath.Join("/", "link_dir") {
		t.Errorf("expected header %q, got %q", filepath.Join("/", "link_dir"), subModel.Header)
	}
	if subModel.Data.Len(80) != 1 || subModel.Data.At(0).Name() != "hello.txt" {
		t.Errorf("unexpected contents in symlinked directory")
	}

	// Press Left to return to where we were
	mod, _ = subModel.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	backModel, ok := mod.(Model)
	if !ok {
		t.Fatalf("expected browser Model on Left, got %T", mod)
	}
	if backModel.Data.Folder() != "/" {
		t.Errorf("expected to return to /, got %s", backModel.Data.Folder())
	}

	// Test File Info on link_dir
	infoMod, err := NewFileInfoModel(backModel.Data.FS(), backModel.Data.Folder(), backModel.Data.At(0), backModel)
	if err != nil {
		t.Fatalf("NewFileInfoModel failed: %v", err)
	}
	infoMod, _ = infoMod.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	infoView := infoMod.(scroller.Model[views.Text]).View().Content
	if !strings.Contains(infoView, "Link To:   target_dir") || !strings.Contains(infoView, "(symlink to directory)") {
		t.Errorf("expected file info to show symlink target and directory note, got:\n%s", infoView)
	}

	// Move cursor to link_file.txt (index 2) and press Right to view file content
	backModel.MoveCursor(2)
	if got := backModel.Data.At(backModel.Cursor()).Name(); got != "link_file.txt" {
		t.Fatalf("expected cursor on link_file.txt, got %s", got)
	}
	fileMod, _ := backModel.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	textScroller, ok := fileMod.(scroller.Model[views.Text])
	if !ok {
		t.Fatalf("expected text viewer for symlinked file, got %T", fileMod)
	}
	textMod, _ := textScroller.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if !strings.Contains(textMod.View().Content, "hello from target") {
		t.Errorf("expected symlinked file content, got:\n%s", textMod.View().Content)
	}
}


