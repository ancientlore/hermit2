package scroller

import "charm.land/bubbles/v2/key"

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
		key.WithHelp("↑", "move cursor up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move cursor down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "esc"),
		key.WithHelp("←/esc", "view previous folder"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("shift+up", "pgup"),
		key.WithHelp("shift+↑/pgup", "move cursor one page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("shift+down", "pgdown"),
		key.WithHelp("shift+↓/pgdn", "move cursor one page down"),
	),
	Home: key.NewBinding(
		key.WithKeys("ctrl+up", "home", "alt+up"),
		key.WithHelp("ctrl+↑/end/alt+↑", "move cursor to beginning"),
	),
	End: key.NewBinding(
		key.WithKeys("ctrl+down", "end", "alt+down"),
		key.WithHelp("ctrl+↓/end/alt+↓", "move cursor to end"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit Hermit"),
	),
}

type ImageKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	ZoomIn    key.Binding
	ZoomOut   key.Binding
	ResetZoom key.Binding
	Grayscale key.Binding
	Back      key.Binding
	Quit      key.Binding
}

var DefaultImageKeyMap = ImageKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "pan image up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "pan image down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "pan image left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "pan image right"),
	),
	ZoomIn: key.NewBinding(
		key.WithKeys("+", "=", "z"),
		key.WithHelp("+/=/z", "zoom in"),
	),
	ZoomOut: key.NewBinding(
		key.WithKeys("-", "_"),
		key.WithHelp("-", "zoom out"),
	),
	ResetZoom: key.NewBinding(
		key.WithKeys("0"),
		key.WithHelp("0", "reset zoom"),
	),
	Grayscale: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "toggle grayscale"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "exit image view"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "exit Hermit"),
	),
}
