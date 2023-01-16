package browser

import (
	"fmt"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	timeFormatOld = "Mo Jan _2  2006"
	timeFormatNew = "Mo Jan _2 15:04"
)

var (
	normal      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	highlight   = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
	header      = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	footer      = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	bold        = lipgloss.NewStyle().Foreground(lipgloss.Color("#AA00AA"))
	special     = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	specialbold = lipgloss.NewStyle().Foreground(lipgloss.Color("#770077"))
)

type model struct {
	fsys     fs.FS         // The filesystem being browsed
	root     string        // The name for the root of the file system
	folder   string        // The current folder in the file system
	entries  []fs.DirEntry // The list of directory entries read
	selected []bool        // Whether an entry is selected
	cursor   int           // Current position of cursor
	offset   int           // The offset of the view (enables scrolling)
	width    int           // The width of the current view
	height   int           // The height of the current view
	footer   string        // Footer text
	prev     tea.Model     // The previous model (for going back)
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.footer = ""
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset--
				}
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				if m.cursor-m.offset+1 > m.height-3 {
					m.offset++
				}
			}

		case "right":
			if len(m.entries) > m.cursor {
				entry := m.entries[m.cursor]
				if entry.IsDir() {
					newModel, err := New(m.fsys, m.root, path.Join(m.folder, m.entries[m.cursor].Name()))
					if err != nil {
						m.footer = err.Error()
					} else {
						newModel.height = m.height
						newModel.width = m.width
						newModel.prev = m
						return *newModel, nil
					}
				}
			}

		case "left":
			if m.prev != nil {
				return m.prev, nil
			}
			a, _ := path.Split(m.folder)
			if a != "/" {
				a = strings.TrimSuffix(a, "/")
			}
			if a != m.folder {
				newModel, err := New(m.fsys, m.root, a)
				if err != nil {
					log.Print(err)
				} else {
					newModel.height = m.height
					newModel.width = m.width
					return newModel, nil
				}
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			ok := m.selected[m.cursor]
			if ok {
				m.selected[m.cursor] = false
			} else {
				m.selected[m.cursor] = true
			}
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := header.Width(m.width).Height(1).Render(filepath.Join(m.root, filepath.FromSlash(m.folder))) + "\n"

	// Iterate over our file entries
	for i := m.offset; i < len(m.entries)+m.offset && i < m.height+m.offset-3; i++ {
		// for i, choice := range m.entries {
		choice := m.entries[i]
		// Is the cursor pointing at this choice?
		style := normal.Width(m.width).Height(1)
		if m.cursor == i {
			style = highlight.Width(m.width).Height(1)
		}

		// Is this choice selected?
		checked := " " // not selected
		if m.selected[i] {
			checked = "*" // selected!
		}

		// Render the row
		info, _ := choice.Info()
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
		s += style.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, info.Mode(), info.Size(), info.ModTime().Format(format), ns.Render(choice.Name()))) + "\n"
	}
	repeat := m.height - 3 - len(m.entries)
	if repeat > 0 {
		s += strings.Repeat("\n", repeat)
	}

	// The footer
	f := m.footer
	if f == "" {
		f = "Press q to quit."
	}
	s += footer.Width(m.width).Height(1).Render(f) + "\n"

	// Send the UI for rendering
	return s
}

func New(fsys fs.FS, root, folder string) (*model, error) {
	rf := strings.TrimPrefix(folder, "/")
	if len(rf) == 0 {
		rf = "."
	}
	entries, err := fs.ReadDir(fsys, rf)
	if err != nil {
		return nil, err
	}
	sort.Sort(sortByExt(entries))
	return &model{
		entries:  entries,
		selected: make([]bool, len(entries)),
		cursor:   0,
		root:     root,
		folder:   folder,
		fsys:     fsys,
	}, nil
}
