package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ancientlore/hermit2/browser"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	wd, _ := os.Getwd()

	var (
		root   = flag.String("root", "/", "Root file system")
		folder = flag.String("path", wd, "Folder")
	)

	flag.Parse()

	m, err := browser.New(os.DirFS(*root), *root, *folder)
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
