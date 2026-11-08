package views

import (
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"
)

func TestFSFilteringAndDeselection(t *testing.T) {
	mapFS := fstest.MapFS{
		"subdir":      {Mode: fs.ModeDir | 0755},
		"subdir/x.go": {Data: []byte("x")},
		"alpha.go":    {Data: []byte("1")},
		"beta.GO":     {Data: []byte("2")},
		"notes.txt":   {Data: []byte("3")},
		"readme.md":   {Data: []byte("4")},
	}

	var fsv FS
	fsv.SetFilter("")
	fsv.SetSort(SortByName, SortForward)
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

	// Select all entries initially
	for i := 0; i < fsv.Len(80); i++ {
		fsv.Select(i, true)
	}

	// Apply filter "*.go" -> directories ("subdir") + "*.go" files ("alpha.go", "beta.GO") remain visible;
	// "notes.txt" and "readme.md" are hidden AND must be deselected.
	fsv.SetFilter("*.go")
	wantGo := []string{"subdir", "alpha.go", "beta.GO"}
	if got := names(); !slices.Equal(got, wantGo) {
		t.Fatalf("Filter *.go = %v, want %v", got, wantGo)
	}
	for i := 0; i < fsv.Len(80); i++ {
		if !fsv.Selected(i) {
			t.Errorf("expected visible entry %s to remain selected", fsv.At(i).Name())
		}
	}

	// Clear filter ("") -> "notes.txt" and "readme.md" become visible again,
	// and must now be deselected because they were not visible under "*.go".
	fsv.SetFilter("")
	for i := 0; i < fsv.Len(80); i++ {
		name := fsv.At(i).Name()
		wantSel := name == "subdir" || name == "alpha.go" || name == "beta.GO"
		if fsv.Selected(i) != wantSel {
			t.Errorf("after clearing filter, entry %s selected=%v, want %v", name, fsv.Selected(i), wantSel)
		}
	}

	// Test multiple semicolon-separated wildcard patterns with '?' and '*'
	fsv.SetFilter("note?.txt;*.md")
	wantMulti := []string{"subdir", "notes.txt", "readme.md"}
	if got := names(); !slices.Equal(got, wantMulti) {
		t.Fatalf("Filter note?.txt;*.md = %v, want %v", got, wantMulti)
	}
	// Since alpha.go and beta.GO were hidden by this filter, they must now be deselected
	fsv.SetFilter("*.*")
	for i := 0; i < fsv.Len(80); i++ {
		name := fsv.At(i).Name()
		wantSel := name == "subdir"
		if fsv.Selected(i) != wantSel {
			t.Errorf("after *.* filter, entry %s selected=%v, want %v", name, fsv.Selected(i), wantSel)
		}
	}

	// Reset default filter
	fsv.SetFilter("")
}

