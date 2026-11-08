package views

import (
	"io/fs"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"charm.land/lipgloss/v2"
)

func TestFSSorting(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)

	mapFS := fstest.MapFS{
		"zeta":          {Mode: fs.ModeDir | 0755, ModTime: t1},
		"zeta/file.txt": {Data: []byte("x"), ModTime: t1},
		"alpha":         {Mode: fs.ModeDir | 0755, ModTime: t2},
		"alpha/sub.txt": {Data: []byte("x"), ModTime: t2},
		".git":          {Mode: fs.ModeDir | 0755, ModTime: t3},
		".git/HEAD":     {Data: []byte("x"), ModTime: t3},
		".env":          {Data: []byte("12"), ModTime: t3},
		"b.go":          {Data: []byte("12345"), ModTime: t1},
		"a.txt":         {Data: []byte("1"), ModTime: t2},
		"c.md":          {Data: []byte("123"), ModTime: t3},
	}

	var fsv FS
	if err := fsv.Init(mapFS, "/", "."); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	names := func() []string {
		out := make([]string, fsv.Len(80))
		for i := range out {
			out[i] = fsv.At(i).Name()
		}
		return out
	}

	// Select "b.go" and verify selection stays attached to "b.go" across sorts
	for i := 0; i < fsv.Len(80); i++ {
		if fsv.At(i).Name() == "b.go" {
			fsv.Select(i, true)
		}
	}
	assertSelected := func(target string) {
		t.Helper()
		for i := 0; i < fsv.Len(80); i++ {
			want := fsv.At(i).Name() == target
			if fsv.Selected(i) != want {
				t.Fatalf("entry %s selected=%v, want %v", fsv.At(i).Name(), fsv.Selected(i), want)
			}
		}
	}

	// 1. SortByName Forward
	fsv.SetSort(SortByName, SortForward)
	wantNameFwd := []string{".git", "alpha", "zeta", ".env", "a.txt", "b.go", "c.md"}
	if got := names(); !slices.Equal(got, wantNameFwd) {
		t.Errorf("SortByName Forward = %v, want %v", got, wantNameFwd)
	}
	assertSelected("b.go")

	// 2. SortByName Reverse (directories & dotfiles still precede regular files)
	fsv.SetSort(SortByName, SortReverse)
	wantNameRev := []string{".git", "zeta", "alpha", ".env", "c.md", "b.go", "a.txt"}
	if got := names(); !slices.Equal(got, wantNameRev) {
		t.Errorf("SortByName Reverse = %v, want %v", got, wantNameRev)
	}
	assertSelected("b.go")

	// 3. SortByExt Forward (.go, .md, .txt)
	fsv.SetSort(SortByExt, SortForward)
	wantExtFwd := []string{".git", "alpha", "zeta", ".env", "b.go", "c.md", "a.txt"}
	if got := names(); !slices.Equal(got, wantExtFwd) {
		t.Errorf("SortByExt Forward = %v, want %v", got, wantExtFwd)
	}
	assertSelected("b.go")

	// 4. SortByExt Reverse (.txt, .md, .go)
	fsv.SetSort(SortByExt, SortReverse)
	wantExtRev := []string{".git", "zeta", "alpha", ".env", "a.txt", "c.md", "b.go"}
	if got := names(); !slices.Equal(got, wantExtRev) {
		t.Errorf("SortByExt Reverse = %v, want %v", got, wantExtRev)
	}
	assertSelected("b.go")

	// 5. SortBySize Forward (a.txt=1, c.md=3, b.go=5)
	fsv.SetSort(SortBySize, SortForward)
	wantSizeFwd := []string{".git", "alpha", "zeta", ".env", "a.txt", "c.md", "b.go"}
	if got := names(); !slices.Equal(got, wantSizeFwd) {
		t.Errorf("SortBySize Forward = %v, want %v", got, wantSizeFwd)
	}
	assertSelected("b.go")

	// 6. SortBySize Reverse (b.go=5, c.md=3, a.txt=1)
	fsv.SetSort(SortBySize, SortReverse)
	wantSizeRev := []string{".git", "alpha", "zeta", ".env", "b.go", "c.md", "a.txt"}
	if got := names(); !slices.Equal(got, wantSizeRev) {
		t.Errorf("SortBySize Reverse = %v, want %v", got, wantSizeRev)
	}
	assertSelected("b.go")

	// 7. SortByDate Forward (b.go=t1, a.txt=t2, c.md=t3)
	fsv.SetSort(SortByDate, SortForward)
	wantDateFwd := []string{".git", "zeta", "alpha", ".env", "b.go", "a.txt", "c.md"}
	if got := names(); !slices.Equal(got, wantDateFwd) {
		t.Errorf("SortByDate Forward = %v, want %v", got, wantDateFwd)
	}
	assertSelected("b.go")

	// 8. SortByDate Reverse (c.md=t3, a.txt=t2, b.go=t1)
	fsv.SetSort(SortByDate, SortReverse)
	wantDateRev := []string{".git", "alpha", "zeta", ".env", "c.md", "a.txt", "b.go"}
	if got := names(); !slices.Equal(got, wantDateRev) {
		t.Errorf("SortByDate Reverse = %v, want %v", got, wantDateRev)
	}
	assertSelected("b.go")

	// Reset defaults
	fsv.SetSort(SortByName, SortForward)
	_ = fs.ValidPath(".")
}

func TestFSSymlinks(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	mapFS := fstest.MapFS{
		"real_dir":          {Mode: fs.ModeDir | 0755, ModTime: t1},
		"real_dir/item.txt": {Data: []byte("hello"), ModTime: t1},
		"real_file.txt":     {Data: []byte("world"), ModTime: t1},
		"dir_link":          {Data: []byte("real_dir"), Mode: fs.ModeSymlink | 0777, ModTime: t1},
		"file_link.txt":     {Data: []byte("real_file.txt"), Mode: fs.ModeSymlink | 0777, ModTime: t1},
	}

	var fsv FS
	if err := fsv.Init(mapFS, "/", "."); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	names := make([]string, fsv.Len(80))
	for i := range names {
		names[i] = fsv.At(i).Name()
	}

	// dir_link should sort with directories (before file_link.txt and real_file.txt)
	want := []string{"dir_link", "real_dir", "file_link.txt", "real_file.txt"}
	if !slices.Equal(names, want) {
		t.Errorf("symlink sort order = %v, want %v", names, want)
	}

	if !fsv.IsDir(0) {
		t.Errorf("expected dir_link IsDir(0) == true")
	}
	if fsv.IsDir(2) {
		t.Errorf("expected file_link.txt IsDir(2) == false")
	}

	// Wide render should inline " -> real_dir"
	wideLine := fsv.Render(0, 100, lipgloss.NewStyle())
	if !strings.Contains(wideLine, " -> ") || !strings.Contains(wideLine, "real_dir") {
		t.Errorf("expected wide render to include symlink target, got %q", wideLine)
	}

	// Narrow render where only the name fits should omit " -> "
	narrowLine := fsv.Render(0, 48, lipgloss.NewStyle())
	if strings.Contains(narrowLine, " -> ") {
		t.Errorf("expected narrow render to omit symlink target, got %q", narrowLine)
	}
}
