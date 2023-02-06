package text

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ancientlore/hermit2/scroller"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/huandu/xstrings"
)

var (
	normal    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	highlight = lipgloss.NewStyle().Background(lipgloss.Color("#7D56F4"))
	header    = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
	footer    = lipgloss.NewStyle().Background(lipgloss.Color("#888B7E"))
)

// Model implements a markdown viewer.
type model struct {
	scroller.Model

	title string   // Title text
	text  []string // Text of file
}

// Update handles messages in order to implement scrolling.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		mod tea.Model
		cmd tea.Cmd
	)
	mod, cmd = m.Model.Update(msg)
	if sm, ok := mod.(scroller.Model); ok {
		m.Model = sm
		return m, cmd
	}
	return mod, cmd
}

// View renders the markdown.
func (m model) View() string {
	s := header.Width(m.Width).Height(1).Render(m.title) + "\n"
	for i := m.Offset(); i < m.Lines && i < m.VisibleLines()+m.Offset(); i++ {
		choice := m.text[i]
		style := normal.Width(m.Width).Height(1)
		if m.Cursor() == i {
			style = highlight.Width(m.Width).Height(1)
		}
		s += style.Render(style.Render(choice)) + "\n"
	}
	return s
}

// New creates a new markdown model.
func New(path string, prev tea.Model, width, height int) *model {
	in, err := ioutil.ReadFile(filepath.FromSlash(path))
	if err != nil {
		in = []byte(err.Error())
	}

	lines := strings.Split(xstrings.ExpandTabs(strings.ReplaceAll(string(in), "\r", ""), 8), "\n")

	return &model{
		title: path,
		text:  lines,
		Model: scroller.Model{
			Lines:       len(lines),
			HeaderLines: 1,
			FooterLines: 0,
			Height:      height,
			Width:       width,
			Prev:        prev,
		},
	}
}
