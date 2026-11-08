package views

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
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
	ellipsis    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
)

type fsEntry struct {
	entry    fs.DirEntry
	info     fs.FileInfo
	selected bool
}

// FS is a viewer for a fs.FS.
type FS struct {
	fsys       fs.FS         // The filesystem being browsed
	root       string        // The name for the root of the file system
	folder     string        // The current folder in the file system
	allEntries []fsEntry     // All directory entries read from the folder
	entries    []fsEntry     // Visible (filtered and sorted) directory entries
	sortType   SortType      // Current sort type
	sortDir    SortDirection // Current sort direction
	filter     string        // Current wildcard file filter
}

// Title returns the full name of the current folder.
func (fsv FS) Title() string {
	return filepath.Join(fsv.root, filepath.FromSlash(fsv.folder))
}

// Folder returns the current folder in the file system.
func (fsv FS) Folder() string {
	return fsv.folder
}

// Root returns the name of the root file system.
func (fsv FS) Root() string {
	return fsv.root
}

// FS returns the file system being viewed.
func (fsv FS) FS() fs.FS {
	return fsv.fsys
}

// At returns the directory entry at position i.
func (fsv FS) At(i int) fs.DirEntry {
	if i >= 0 && i < len(fsv.entries) {
		return fsv.entries[i].entry
	}
	return nil
}

// Selected returns whether the entry at position i is selected.
func (fsv FS) Selected(i int) bool {
	if i >= 0 && i < len(fsv.entries) {
		return fsv.entries[i].selected
	}
	return false
}

// Select sets the selected flag at position i to b.
func (fsv *FS) Select(i int, b bool) {
	if i >= 0 && i < len(fsv.entries) {
		fsv.entries[i].selected = b
		fsv.setAllEntrySelected(fsv.entries[i].entry.Name(), b)
	}
}

// ToggleSelect toggles the selected flag at position i.
func (fsv *FS) ToggleSelect(i int) {
	if i >= 0 && i < len(fsv.entries) {
		b := !fsv.entries[i].selected
		fsv.entries[i].selected = b
		fsv.setAllEntrySelected(fsv.entries[i].entry.Name(), b)
	}
}

func (fsv *FS) setAllEntrySelected(name string, b bool) {
	for i := range fsv.allEntries {
		if fsv.allEntries[i].entry.Name() == name {
			fsv.allEntries[i].selected = b
			return
		}
	}
}

// Len returns the number of file entries.
func (fsv FS) Len(width int) int {
	return len(fsv.entries)
}

// Close closes the viewer, if necessary.
func (fsv FS) Close() error {
	return nil
}

// Render formats the line at position i using the base style and view width.
func (fsv FS) Render(i, width int, baseStyle lipgloss.Style) string {
	if i < 0 || i >= len(fsv.entries) {
		return ""
	}
	item := fsv.entries[i]
	choice := item.entry
	// Is this choice selected?
	checked := " " // not selected
	if item.selected {
		checked = "*" // selected!
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
	ns = ns.Background(baseStyle.GetBackground())

	// Render the row
	var prefix string
	info := item.info
	if info != nil {
		n := time.Now().Local()
		t := info.ModTime().Local()
		format := timeFormatNew
		if t.Year() < n.Year() {
			format = timeFormatOld
		}
		prefix = fmt.Sprintf("%s %11s %10d %s ", checked, info.Mode(), info.Size(), info.ModTime().Format(format))
	} else {
		prefix = fmt.Sprintf("%s %11s %10d %s ", checked, "?", 0, "")
	}

	es := ellipsis.Background(baseStyle.GetBackground())
	name := renderName(choice.Name(), width-lipgloss.Width(prefix), ns, es)
	return baseStyle.Render(prefix + name)
}

func renderName(s string, maxWidth int, nameStyle, ellipsisStyle lipgloss.Style) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return nameStyle.Render(s)
	}
	if maxWidth <= 3 {
		return ellipsisStyle.Render("..."[:maxWidth])
	}
	r := []rune(s)
	avail := maxWidth - 3
	frontRunes := (avail + 1) / 2
	backRunes := avail - frontRunes
	if frontRunes > len(r) {
		frontRunes = len(r)
	}
	if backRunes > len(r)-frontRunes {
		backRunes = len(r) - frontRunes
	}
	front := r[:frontRunes]
	back := r[len(r)-backRunes:]
	for (len(front) > 0 || len(back) > 0) && lipgloss.Width(string(front))+3+lipgloss.Width(string(back)) > maxWidth {
		if len(front) >= len(back) && len(front) > 0 {
			front = front[:len(front)-1]
		} else if len(back) > 0 {
			back = back[1:]
		}
	}
	return nameStyle.Render(string(front)) + ellipsisStyle.Render("...") + nameStyle.Render(string(back))
}

// Footer formats the footer using the base style and view width.
func (fsv FS) Footer(i, width int, baseStyle lipgloss.Style) string {
	sel := 0
	for _, item := range fsv.entries {
		if item.selected {
			sel++
		}
	}
	return baseStyle.Render(fmt.Sprintf("? for help    %d / %d selected", sel, len(fsv.entries)))
}

// Init initializes a new file system view.
func (fsv *FS) Init(fsys fs.FS, root, folder string) error {
	rf := strings.TrimPrefix(folder, "/")
	if len(rf) == 0 {
		rf = "."
	}
	dirEntries, err := fs.ReadDir(fsys, rf)
	if err != nil {
		return err
	}

	fsv.allEntries = make([]fsEntry, len(dirEntries))
	for i, entry := range dirEntries {
		info, _ := entry.Info()
		fsv.allEntries[i] = fsEntry{
			entry: entry,
			info:  info,
		}
	}
	fsv.root = root
	fsv.folder = folder
	fsv.fsys = fsys
	fsv.sortType = defaultSortType
	fsv.sortDir = defaultSortDir
	fsv.filter = defaultFilter

	fsv.applyFilter()

	return nil
}
