package browser

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Left         key.Binding
	Right        key.Binding
	ToggleSelect key.Binding
	Select       key.Binding
	DeSelect     key.Binding
	SelectAll    key.Binding
	DeSelectAll  key.Binding
	RunShell     key.Binding
	GoHome       key.Binding
	Refresh      key.Binding
}

var DefaultKeyMap = KeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "esc"),
		key.WithHelp("←/esc", "previous folder"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "open folder/file"),
	),
	ToggleSelect: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("space/↲", "toggle select"),
	),
	Select: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "select"),
	),
	DeSelect: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "deselect"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("alt+=", "ctrl+="),
		key.WithHelp("alt+=/ctrl+=", "select all"),
	),
	DeSelectAll: key.NewBinding(
		key.WithKeys("alt+-", "ctrl+-"),
		key.WithHelp("alt+-/ctrl+-", "deselect all"),
	),
	RunShell: key.NewBinding(
		key.WithKeys("$"),
		key.WithHelp("$", "run shell"),
	),
	GoHome: key.NewBinding(
		key.WithKeys("~", "alt+h", "ctrl+h"),
		key.WithHelp("~/alt+h/ctrl+h", "navigate to home folder"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("alt+r", "ctrl+r", "f5"),
		key.WithHelp("alt+r/ctrl+r/f5", "refresh"),
	),
}
