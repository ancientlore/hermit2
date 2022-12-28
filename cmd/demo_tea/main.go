package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	fsys     fs.FS
	folder   string
	entries  []fs.DirEntry
	selected []bool
	cursor   int
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}

		case "right":
			entry := m.entries[m.cursor]
			if entry.IsDir() {
				newModel, err := New(m.fsys, path.Join(m.folder, m.entries[m.cursor].Name()))
				if err != nil {
					//
				} else {
					return *newModel, nil
				}
			}

		case "left":
			a, _ := path.Split(m.folder)
			if a != "/" {
				a = strings.TrimSuffix(a, "/")
			}
			if a != m.folder {
				newModel, err := New(m.fsys, a)
				if err != nil {
					log.Print(err)
				} else {
					return *newModel, nil
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
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

const (
	timeFormatOld = "Mo Jan _2  2006"
	timeFormatNew = "Mo Jan _2 15:04"
)

func (m model) View() string {
	// The header
	s := m.folder + "\n\n"

	// Iterate over our file entries
	for i, choice := range m.entries {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
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
		s += fmt.Sprintf("%s %s %s %10d %s %s\n", cursor, checked, info.Mode(), info.Size(), info.ModTime().Format(format), choice.Name())
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func New(fsys fs.FS, root string) (*model, error) {
	rf := strings.TrimPrefix(root, "/")
	if len(rf) == 0 {
		rf = "."
	}
	entries, err := fs.ReadDir(fsys, rf)
	if err != nil {
		return nil, err
	}
	return &model{
		entries:  entries,
		selected: make([]bool, len(entries)),
		cursor:   0,
		folder:   root,
		fsys:     fsys,
	}, nil
}

func main() {
	m, err := New(os.DirFS("/"), os.Getenv("HOME"))
	if err != nil {
		fmt.Printf("Error opening folder: %v", err)
		os.Exit(1)
	}

	p := tea.NewProgram(*m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
