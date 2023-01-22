package browser

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Home         key.Binding
	End          key.Binding
	ToggleSelect key.Binding
	Select       key.Binding
	DeSelect     key.Binding
	SelectAll    key.Binding
	DeSelectAll  key.Binding
	Quit         key.Binding
	RunShell     key.Binding
	GoHome       key.Binding
	Refresh      key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "esc"),
		key.WithHelp("←/esc", "previous folder"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "open folder/file"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("shift+up", "pgup"),
		key.WithHelp("shift+↑/pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("shift+down", "pgdown"),
		key.WithHelp("shift+↓/pgdn", "page down"),
	),
	Home: key.NewBinding(
		key.WithKeys("ctrl+up", "home", "alt+up"),
		key.WithHelp("ctrl+↑/end/alt+↑", "go to beginning"),
	),
	End: key.NewBinding(
		key.WithKeys("ctrl+down", "end", "alt+down"),
		key.WithHelp("ctrl+↓/end/alt+↓", "go to end"),
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
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit Hermit"),
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
