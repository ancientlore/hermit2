package browser

import (
	"github.com/ancientlore/hermit2/popup"
	"github.com/ancientlore/hermit2/scroller"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

const filterMaxLen = 256

var filterHistory = popup.NewHistory()

// FilterDialogModel is a popup dialog for entering file wildcard filters,
// modeled after the original C++ FilterDialog class.
type FilterDialogModel struct {
	parent Model
	edit   popup.Edit
}

// NewFilterDialogModel creates a new filter dialog overlaid on the given browser Model.
func NewFilterDialogModel(parent Model) FilterDialogModel {
	return FilterDialogModel{
		parent: parent,
		edit:   popup.NewEdit(47, filterMaxLen, parent.Data.Filter(), filterHistory),
	}
}

// Init initializes the filter dialog model.
func (m FilterDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles key presses and window resize events for the filter dialog.
func (m FilterDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mod, cmd := m.parent.Update(msg)
		if bm, ok := mod.(Model); ok {
			m.parent = bm
		}
		return m, cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			newFilter := m.edit.Accept()
			m.parent.Data.SetFilter(newFilter)
			m.parent.HeaderTail = newFilter
			m.parent.MoveCursor(0)
			return m.parent, nil
		case "esc":
			return m.parent, nil
		default:
			if key.Matches(msg, scroller.DefaultKeyMap.Quit) {
				m.parent.Data.Close()
				return m, tea.Quit
			}
			m.edit.HandleKey(msg)
		}
	}

	return m, nil
}

// View renders the filter dialog overlaid on top of the parent browser view.
func (m FilterDialogModel) View() tea.View {
	p := popup.New(51, 8, "Filter Files")

	p.WriteText(2, 2, "Enter the new filter using wildcards (* and ?).")
	p.WriteText(2, 3, "An empty filter gets all files, ';' for many.")
	p.DrawEdit(2, 4, &m.edit)
	p.WriteText(2, 6, "Then press Enter to accept or Esc to cancel.")

	return p.Overlay(m.parent.View())
}

