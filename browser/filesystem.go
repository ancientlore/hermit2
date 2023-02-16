package browser

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type FSView struct {
	fsys     fs.FS         // The filesystem being browsed
	root     string        // The name for the root of the file system
	folder   string        // The current folder in the file system
	entries  []fs.DirEntry // The list of directory entries read
	selected []bool        // Whether an entry is selected
}

// Folder returns the name of the current folder.
func (fs FSView) Folder() string {
	return filepath.Join(fs.root, filepath.FromSlash(fs.folder))
}

// At returns the directory entry at position i.
func (fs FSView) At(i int) fs.DirEntry {
	if i >= 0 && i < len(fs.entries) {
		return fs.entries[i]
	}
	return nil
}

// Selected returns whether the entry at position i is selected.
func (fs FSView) Selected(i int) bool {
	if i >= 0 && i < len(fs.selected) {
		return fs.selected[i]
	}
	return false
}

// Len returns the number of file entries.
func (fs FSView) Len() int {
	return len(fs.entries)
}

// Render formats the line at position i using the base style and view width.
func (fs FSView) Render(i, width int, baseStyle lipgloss.Style) string {
	var s string
	choice := fs.entries[i]
	// Is this choice selected?
	checked := " " // not selected
	if fs.selected[i] {
		checked = "*" // selected!
	}

	// Render the row
	info, err := choice.Info()
	if err == nil {
		n := time.Now().Local()
		t := info.ModTime().Local()
		format := timeFormatNew
		if t.Year() < n.Year() {
			format = timeFormatOld
		}
		ns := normal
		if choice.IsDir() {
			ns = bold
			if strings.HasPrefix(choice.Name(), ".") {
				ns = specialbold
			}
		} else if strings.HasPrefix(choice.Name(), ".") {
			ns = special
		}
		s = baseStyle.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, info.Mode(), info.Size(), info.ModTime().Format(format), ns.Render(choice.Name()))) + "\n"
	} else {
		ns := normal
		if choice.IsDir() {
			ns = bold
			if strings.HasPrefix(choice.Name(), ".") {
				ns = specialbold
			}
		} else if strings.HasPrefix(choice.Name(), ".") {
			ns = special
		}
		s = baseStyle.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, "?", 0, "", ns.Render(choice.Name()))) + "\n"
	}
	return s
}

// NewFileSystem creates a new file system view.
func NewFileSystem(fsys fs.FS, root, folder string) (*FSView, error) {
	rf := strings.TrimPrefix(folder, "/")
	if len(rf) == 0 {
		rf = "."
	}
	entries, err := fs.ReadDir(fsys, rf)
	if err != nil {
		return nil, err
	}
	sort.Sort(sortByExt(entries))
	return &FSView{
		entries:  entries,
		selected: make([]bool, len(entries)),
		root:     root,
		folder:   folder,
		fsys:     fsys,
	}, nil
}
