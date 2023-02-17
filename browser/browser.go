package browser

import (
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/ancientlore/hermit2/config"
	"github.com/ancientlore/hermit2/scroller"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type refreshMsg struct{}

func refreshCmd() tea.Msg {
	return refreshMsg{}
}

type model struct {
	scroller.Model[FSView]
	footer string
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					newModel, err := New(m.Data.fsys, m.Data.root, path.Join(m.Data.folder, m.Data.At(m.Cursor()).Name()))
					if err != nil {
						m.footer = err.Error()
					} else {
						newModel.Prev = m
						return *newModel, sizeCmd
					}
				} else if strings.HasPrefix(mime.TypeByExtension(path.Ext(entry.Name())), "text") {
					return NewTextModel(path.Join(m.Data.folder, m.Data.At(m.Cursor()).Name()), m), sizeCmd
				} else if entry.Type().IsRegular() {
					f, err := os.Open(path.Join(m.Data.folder, m.Data.At(m.Cursor()).Name()))
					if err == nil {
						defer f.Close()
						b := make([]byte, 512)
						n, err := f.Read(b)
						if err == nil {
							if strings.HasPrefix(http.DetectContentType(b[0:n]), "text") {
								return NewTextModel(path.Join(m.Data.folder, m.Data.At(m.Cursor()).Name()), m), sizeCmd
							}
						} else {
							m.footer = err.Error()
						}
					} else {
						m.footer = err.Error()
					}
				}
			}

		case key.Matches(msg, DefaultKeyMap.Left):
			if m.Prev != nil {
				return m.Prev, sizeCmd
			}
			a, _ := path.Split(m.Data.folder)
			if a != "/" {
				a = strings.TrimSuffix(a, "/")
			}
			if a != m.Data.folder {
				newModel, err := New(m.Data.fsys, m.Data.root, a)
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
			m.BumpCursor()

		case key.Matches(msg, DefaultKeyMap.Select):
			m.Data.Select(m.Cursor(), true)
			m.BumpCursor()

		case key.Matches(msg, DefaultKeyMap.DeSelect):
			m.Data.Select(m.Cursor(), false)
			m.BumpCursor()

		case key.Matches(msg, DefaultKeyMap.SelectAll):
			for i := 0; i < m.Data.Len(); i++ {
				m.Data.Select(i, true)
			}

		case key.Matches(msg, DefaultKeyMap.DeSelectAll):
			for i := 0; i < m.Data.Len(); i++ {
				m.Data.Select(i, false)
			}

		case key.Matches(msg, DefaultKeyMap.RunShell):
			c := exec.Command(config.Shell())
			c.Dir = filepath.Join(m.Data.root, filepath.FromSlash(m.Data.folder))
			cmd := tea.ExecProcess(c, nil)
			return m, tea.Sequence(tea.ClearScreen, cmd, refreshCmd, sizeCmd)

		default:
			handled = false
		}

	case refreshMsg:
		err := m.Data.Init(m.Data.fsys, m.Data.root, m.Data.folder)
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
		if scr, ok := mod.(scroller.Model[FSView]); ok {
			m.Model = scr
			return m, cmd
		}
		return mod, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func New(fsys fs.FS, root, folder string) (*model, error) {
	var m model
	err := m.Model.Data.Init(fsys, root, folder)
	if err != nil {
		return nil, err
	}
	m.Header = m.Data.Folder()
	return &m, nil
}
