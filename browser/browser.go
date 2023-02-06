package browser

import (
	"fmt"
	"io/fs"
	"log"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ancientlore/hermit2/config"
	"github.com/ancientlore/hermit2/text"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	timeFormatOld = "Mon Jan _2  2006"
	timeFormatNew = "Mon Jan _2 15:04"
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

type refreshMsg struct{}

func refreshCmd() tea.Msg {
	return refreshMsg{}
}

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

		lastKey = msg.String()

		// Cool, what was the actual key pressed?
		switch {

		// These keys should exit the program.
		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit

		// The "up" keys move the cursor up
		case key.Matches(msg, DefaultKeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" keys move the cursor down
		case key.Matches(msg, DefaultKeyMap.Down):
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}

		case key.Matches(msg, DefaultKeyMap.Right):
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
				} else if strings.HasPrefix(mime.TypeByExtension(path.Ext(entry.Name())), "text") {
					return text.New(path.Join(m.folder, m.entries[m.cursor].Name()), m, m.width, m.height), nil
				}
			}

		case key.Matches(msg, DefaultKeyMap.Left):
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

		case key.Matches(msg, DefaultKeyMap.GoHome):
			home := config.HomeFolder()
			home, _ = filepath.Abs(home)
			fsRoot := filepath.VolumeName(home)
			fsPath := strings.TrimPrefix(home, fsRoot)
			fsRoot += string(filepath.Separator)
			newModel, err := New(os.DirFS(fsRoot), fsRoot, filepath.ToSlash(fsPath))
			if err != nil {
				m.footer = err.Error()
			} else {
				newModel.height = m.height
				newModel.width = m.width
				return *newModel, nil
			}

		case key.Matches(msg, DefaultKeyMap.Refresh):
			return m, refreshCmd

		case key.Matches(msg, DefaultKeyMap.Home):
			m.cursor = 0

		case key.Matches(msg, DefaultKeyMap.End):
			m.cursor = len(m.entries) - 1

		case key.Matches(msg, DefaultKeyMap.PageUp):
			m.cursor -= m.height - 4
			if m.cursor < 0 {
				m.cursor = 0
			}

		case key.Matches(msg, DefaultKeyMap.PageDown):
			m.cursor += m.height - 4
			if m.cursor >= len(m.entries) {
				m.cursor = len(m.entries) - 1
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case key.Matches(msg, DefaultKeyMap.ToggleSelect):
			if m.cursor < len(m.entries) {
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

		case key.Matches(msg, DefaultKeyMap.Select):
			if m.cursor < len(m.entries) {
				m.selected[m.cursor] = true
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
			}

		case key.Matches(msg, DefaultKeyMap.DeSelect):
			if m.cursor < len(m.entries) {
				m.selected[m.cursor] = false
				if m.cursor < len(m.entries)-1 {
					m.cursor++
				}
			}

		case key.Matches(msg, DefaultKeyMap.SelectAll):
			for i := range m.selected {
				m.selected[i] = true
			}

		case key.Matches(msg, DefaultKeyMap.DeSelectAll):
			for i := range m.selected {
				m.selected[i] = false
			}

		case key.Matches(msg, DefaultKeyMap.RunShell):
			c := exec.Command(config.Shell())
			c.Dir = filepath.Join(m.root, filepath.FromSlash(m.folder))
			cmd := tea.ExecProcess(c, nil)
			return m, tea.Sequence(tea.ClearScreen, cmd, refreshCmd)
		}

		m.fixOffset()

	case refreshMsg:
		newModel, err := New(m.fsys, m.root, m.folder)
		if err != nil {
			log.Print(err)
		} else {
			newModel.height = m.height
			newModel.width = m.width
			newModel.prev = m.prev
			newModel.cursor = m.cursor
			newModel.offset = m.offset
			if newModel.cursor >= len(newModel.entries) {
				newModel.cursor = len(newModel.entries) - 1
				if newModel.cursor < 0 {
					newModel.cursor = 0
				}
			}
			newModel.fixOffset()
			return newModel, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m *model) fixOffset() {
	// Fix cursor location
	if m.cursor > len(m.entries)-1 {
		m.cursor = len(m.entries) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	// cursor before offset - offset needs to be decreased
	if m.cursor < m.offset {
		m.offset = m.cursor
	}

	// cursor greater than offset + window size - offset needs to be increased
	if m.cursor-m.offset+1 > m.height-3 {
		m.offset = m.cursor + 1 - m.height + 3
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

// TODO: remove this later
var lastKey string

func (m model) View() string {
	// The header
	s := header.Width(m.width).Height(1).Render(filepath.Join(m.root, filepath.FromSlash(m.folder))) + "\n"

	// Iterate over our file entries
	for i := m.offset; i < len(m.entries) && i < m.height+m.offset-3; i++ {
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
			s += style.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, info.Mode(), info.Size(), info.ModTime().Format(format), ns.Render(choice.Name()))) + "\n"
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
			s += style.Render(fmt.Sprintf("%s %11s %10d %s %s", checked, "?", 0, "", ns.Render(choice.Name()))) + "\n"
		}
	}
	repeat := m.height - 3 - len(m.entries)
	if repeat > 0 {
		s += strings.Repeat("\n", repeat)
	}

	// The footer
	f := m.footer
	if f == "" {
		if m.cursor < len(m.entries) {
			ext := path.Ext(m.entries[m.cursor].Name())
			f = mime.TypeByExtension(ext)
		}
	}
	if f == "" {
		f = "Ctrl-C to quit."
	}
	f += " Last key: " + lastKey

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
