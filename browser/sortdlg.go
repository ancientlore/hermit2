package browser

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/ancientlore/hermit2/popup"
	"github.com/ancientlore/hermit2/scroller"
	"github.com/ancientlore/hermit2/views"
)

// SortDialogModel is a popup dialog for configuring directory sorting options.
type SortDialogModel struct {
	parent   Model
	sortType views.SortType
	sortDir  views.SortDirection
}

// NewSortDialogModel creates a new sorting dialog overlaid on the given browser Model.
func NewSortDialogModel(parent Model) SortDialogModel {
	return SortDialogModel{
		parent:   parent,
		sortType: parent.Data.SortType(),
		sortDir:  parent.Data.SortDirection(),
	}
}

// Init initializes the sorting dialog model.
func (m SortDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles key presses and window resize events for the sorting dialog.
func (m SortDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mod, cmd := m.parent.Update(msg)
		if bm, ok := mod.(Model); ok {
			m.parent = bm
		}
		return m, cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "n", "N":
			m.sortType = views.SortByName
		case "e", "E":
			m.sortType = views.SortByExt
		case "s", "S":
			m.sortType = views.SortBySize
		case "d", "D":
			m.sortType = views.SortByDate
		case "f", "F":
			m.sortDir = views.SortForward
		case "r", "R":
			m.sortDir = views.SortReverse
		case "enter":
			m.parent.Data.SetSort(m.sortType, m.sortDir)
			if err := saveOptions(); err != nil {
				m.parent.footer = err.Error()
			}
			return m.parent, nil
		case "esc":
			return m.parent, nil
		default:
			if key.Matches(msg, scroller.DefaultKeyMap.Quit) {
				m.parent.Data.Close()
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the sorting dialog overlaid on top of the parent browser view.
func (m SortDialogModel) View() tea.View {
	p := popup.New(47, 10, "Sorting")

	p.WriteText(2, 2, "Type the highlighted letter to select")
	p.WriteText(2, 3, "sorting options.")
	p.WriteText(2, 5, "Sort by:")
	p.WriteText(2, 6, "Direction:")
	p.WriteText(2, 8, "Press Enter when complete or Esc to cancel.")

	p.DrawField(14, 5, "Name", m.sortType == views.SortByName)
	p.DrawField(20, 5, "Extension", m.sortType == views.SortByExt)
	p.DrawField(31, 5, "Size", m.sortType == views.SortBySize)
	p.DrawField(37, 5, "Date", m.sortType == views.SortByDate)
	p.DrawField(14, 6, "Forward", m.sortDir == views.SortForward)
	p.DrawField(23, 6, "Reverse", m.sortDir == views.SortReverse)

	return p.Overlay(m.parent.View())
}
