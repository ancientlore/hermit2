package browser

import (
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ancientlore/hermit2/config"
	"github.com/ancientlore/hermit2/scroller"
	"github.com/ancientlore/hermit2/views"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type refreshMsg struct{}

func refreshCmd() tea.Msg {
	return refreshMsg{}
}

type Model struct {
	scroller.Model[views.FS]
	footer string
}

func (m Model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	handled := true

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		sizeCmd := func() tea.Msg { return tea.WindowSizeMsg{Width: m.Width(), Height: m.Height()} }

		// Cool, what was the actual key pressed?
		switch {

		case key.Matches(msg, DefaultKeyMap.Right):
			entry := m.Data.At(m.Cursor())
			if entry != nil {
				if entry.IsDir() {
					newModel, err := New(m.Data.FS(), m.Data.Root(), path.Join(m.Data.Folder(), m.Data.At(m.Cursor()).Name()))
					if err != nil {
						m.footer = err.Error()
					} else {
						newModel.Prev = m
						return *newModel, sizeCmd
					}
				} else {
					newModel, err := NewFileModel(m.Data.FS(), m.Data.Folder(), entry, m)
					if err == nil {
						return newModel, sizeCmd
					} else {
						m.footer = err.Error()
					}
				}
			}

		case key.Matches(msg, DefaultKeyMap.Left):
			if m.Prev != nil {
				return m.Prev, sizeCmd
			}
			a, _ := path.Split(m.Data.Folder())
			if a != "/" {
				a = strings.TrimSuffix(a, "/")
			}
			if a != m.Data.Folder() {
				newModel, err := New(m.Data.FS(), m.Data.Root(), a)
				if err != nil {
					log.Print(err)
				} else {
					return newModel, sizeCmd
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
				return *newModel, sizeCmd
			}

		case key.Matches(msg, DefaultKeyMap.Refresh):
			return m, refreshCmd

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case key.Matches(msg, DefaultKeyMap.ToggleSelect):
			m.Data.ToggleSelect(m.Cursor())
			m.MoveCursor(1)

		case key.Matches(msg, DefaultKeyMap.Select):
			m.Data.Select(m.Cursor(), true)
			m.MoveCursor(1)

		case key.Matches(msg, DefaultKeyMap.DeSelect):
			m.Data.Select(m.Cursor(), false)
			m.MoveCursor(1)

		case key.Matches(msg, DefaultKeyMap.SelectAll):
			for i := 0; i < m.Data.Len(m.Width()); i++ {
				m.Data.Select(i, true)
			}

		case key.Matches(msg, DefaultKeyMap.DeSelectAll):
			for i := 0; i < m.Data.Len(m.Width()); i++ {
				m.Data.Select(i, false)
			}

		case key.Matches(msg, DefaultKeyMap.RunShell):
			c := exec.Command(config.Shell())
			c.Dir = filepath.Join(m.Data.Root(), filepath.FromSlash(m.Data.Folder()))
			cmd := tea.ExecProcess(c, nil)
			return m, tea.Sequence(tea.ClearScreen, cmd, refreshCmd, sizeCmd)

		case key.Matches(msg, DefaultKeyMap.Help):
			newModel, err := NewHelpModel(m)
			if err == nil {
				return newModel, sizeCmd
			} else {
				m.footer = err.Error()
			}

		case key.Matches(msg, DefaultKeyMap.FileInfo):
			entry := m.Data.At(m.Cursor())
			if entry != nil {
				newModel, err := NewFileInfoModel(m.Data.FS(), m.Data.Folder(), entry, m)
				if err == nil {
					return newModel, sizeCmd
				} else {
					m.footer = err.Error()
				}
			}

		case key.Matches(msg, DefaultKeyMap.ViewBinary):
			entry := m.Data.At(m.Cursor())
			if entry != nil {
				newModel, err := NewBinaryFileModel(m.Data.FS(), m.Data.Folder(), entry, m)
				if err == nil {
					return newModel, sizeCmd
				} else {
					m.footer = err.Error()
				}
			}

		default:
			handled = false
		}

	case refreshMsg:
		err := m.Data.Init(m.Data.FS(), m.Data.Root(), m.Data.Folder())
		if err != nil {
			log.Print(err)
		} else {
			return m, func() tea.Msg { return tea.WindowSizeMsg{Width: m.Width(), Height: m.Height()} }
		}

	default:
		handled = false
	}

	if !handled {
		mod, cmd := m.Model.Update(msg)
		if scr, ok := mod.(scroller.Model[views.FS]); ok {
			m.Model = scr
			return m, cmd
		}
		return mod, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func New(fsys fs.FS, root, folder string) (*Model, error) {
	var m Model
	err := m.Model.Data.Init(fsys, root, folder)
	if err != nil {
		return nil, err
	}
	m.Header = m.Data.Title()
	return &m, nil
}
