package scroller

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding
	Quit     key.Binding
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
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit Hermit"),
	),
}
