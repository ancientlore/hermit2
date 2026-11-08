package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ancientlore/hermit2/views"
)

func TestOptionsLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	optsPath := filepath.Join(tmpDir, OptionsFileName)

	// 1. Loading a non-existent file returns defaults without error
	loaded, err := LoadOptionsFrom(optsPath)
	if err != nil {
		t.Fatalf("LoadOptionsFrom non-existent failed: %v", err)
	}
	if loaded.SortBy != views.SortByName || loaded.SortDirection != views.SortForward || loaded.Filter != "" || len(loaded.FilterHistory) != 0 {
		t.Errorf("unexpected defaults: %+v", loaded)
	}

	// 2. Save custom options and verify TOML contents and round-trip load
	want := Options{
		SortBy:        views.SortByDate,
		SortDirection: views.SortReverse,
		Filter:        "*.go;*.md",
		FilterHistory: []string{"*.go;*.md", "*.txt"},
	}
	if err := SaveOptionsTo(optsPath, want); err != nil {
		t.Fatalf("SaveOptionsTo failed: %v", err)
	}

	rawBytes, err := os.ReadFile(optsPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	raw := string(rawBytes)
	for _, substr := range []string{
		`sort_by = "date"`,
		`sort_direction = "reverse"`,
		`filter = "*.go;*.md"`,
	} {
		if !strings.Contains(raw, substr) {
			t.Errorf("expected %q in TOML output:\n%s", substr, raw)
		}
	}

	got, err := LoadOptionsFrom(optsPath)
	if err != nil {
		t.Fatalf("LoadOptionsFrom failed: %v", err)
	}
	if got.SortBy != want.SortBy || got.SortDirection != want.SortDirection || got.Filter != want.Filter || !slices.Equal(got.FilterHistory, want.FilterHistory) {
		t.Errorf("round-trip mismatch:\ngot:  %+v\nwant: %+v", got, want)
	}
}

