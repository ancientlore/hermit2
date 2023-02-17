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

const (
	timeFormatOld = "Mon Jan _2  2006"
	timeFormatNew = "Mon Jan _2 15:04"
)

var (
	normal      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	bold        = lipgloss.NewStyle().Foreground(lipgloss.Color("#AA00AA"))
	special     = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	specialbold = lipgloss.NewStyle().Foreground(lipgloss.Color("#770077"))
)

// FSView is a viewer for a fs.FS.
type FSView struct {
	fsys     fs.FS         // The filesystem being browsed
	root     string        // The name for the root of the file system
	folder   string        // The current folder in the file system
	entries  []fs.DirEntry // The list of directory entries read
	selected []bool        // Whether an entry is selected
}

// Folder returns the name of the current folder.
func (fsv FSView) Folder() string {
	return filepath.Join(fsv.root, filepath.FromSlash(fsv.folder))
}

// At returns the directory entry at position i.
func (fsv FSView) At(i int) fs.DirEntry {
	if i >= 0 && i < len(fsv.entries) {
		return fsv.entries[i]
	}
	return nil
}

// Selected returns whether the entry at position i is selected.
func (fsv FSView) Selected(i int) bool {
	if i >= 0 && i < len(fsv.selected) {
		return fsv.selected[i]
	}
	return false
}

// Select sets the selected flag at position i to b.
func (fsv *FSView) Select(i int, b bool) {
	if i >= 0 && i < len(fsv.selected) {
		fsv.selected[i] = b
	}
}

// ToggleSelect toggles the selected flag at position i.
func (fsv *FSView) ToggleSelect(i int) {
	if i >= 0 && i < len(fsv.selected) {
		fsv.selected[i] = !fsv.selected[i]
	}
}

// Len returns the number of file entries.
func (fsv FSView) Len() int {
	return len(fsv.entries)
}

// Render formats the line at position i using the base style and view width.
func (fsv FSView) Render(i, width int, baseStyle lipgloss.Style) string {
	var s string
	choice := fsv.entries[i]
	// Is this choice selected?
	checked := " " // not selected
	if fsv.selected[i] {
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
		s = baseStyle.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, info.Mode(), info.Size(), info.ModTime().Format(format), ns.Render(choice.Name())))
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
		s = baseStyle.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, "?", 0, "", ns.Render(choice.Name())))
	}
	return s
}

// Footer formats the footer using the base style and view width.
func (fsv FSView) Footer(i, width int, baseStyle lipgloss.Style) string {
	sel := 0
	for i := range fsv.selected {
		if fsv.selected[i] {
			sel++
		}
	}
	return baseStyle.Render(fmt.Sprintf("Ctrl+C to exit    %d / %d selected", sel, len(fsv.entries)))
}

// Init initializes a new file system view.
func (fsv *FSView) Init(fsys fs.FS, root, folder string) error {
	rf := strings.TrimPrefix(folder, "/")
	if len(rf) == 0 {
		rf = "."
	}
	entries, err := fs.ReadDir(fsys, rf)
	if err != nil {
		return err
	}
	sort.Sort(sortByExt(entries))

	fsv.entries = entries
	fsv.selected = make([]bool, len(entries))
	fsv.root = root
	fsv.folder = folder
	fsv.fsys = fsys

	return nil
}
