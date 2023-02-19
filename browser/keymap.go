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
	Help         key.Binding
	ViewBinary   key.Binding
	FileInfo     key.Binding
}

var DefaultKeyMap = KeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "esc"),
		key.WithHelp("←/esc", "view previous folder"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "view subfolder or file"),
	),
	ToggleSelect: key.NewBinding(
		key.WithKeys("enter", " "),
		key.WithHelp("space/↲", "toggle selection"),
	),
	Select: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "select directory entry"),
	),
	DeSelect: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "deselect directory entry"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("alt+=", "ctrl+="),
		key.WithHelp("alt+=/ctrl+=", "select all entries"),
	),
	DeSelectAll: key.NewBinding(
		key.WithKeys("alt+-", "ctrl+-"),
		key.WithHelp("alt+-/ctrl+-", "deselect all entries"),
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
		key.WithHelp("alt+r/ctrl+r/f5", "refresh directory listing"),
	),
	Help: key.NewBinding(
		key.WithKeys("alt+h", "ctrl+h", "?"),
		key.WithHelp("?", "show help"),
	),
	ViewBinary: key.NewBinding(
		key.WithKeys("#"),
		key.WithHelp("#", "view file bytes"),
	),
	FileInfo: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "view file information"),
	),
}
