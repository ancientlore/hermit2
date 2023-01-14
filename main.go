package main

import (
	"fmt"
	"os"

	"github.com/ancientlore/hermit2/browser"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m, err := browser.New(os.DirFS("/"), os.Getenv("HOME"))
	if err != nil {
		fmt.Printf("Error opening folder: %v", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
