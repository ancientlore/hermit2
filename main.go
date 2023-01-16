package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ancientlore/hermit2/browser"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Get working directory
	wd, err := os.Getwd()
	if err != nil {
		wd, _ = filepath.Abs(".")
	}

	// Process flags
	var (
		folder = flag.String("path", wd, "Startup path")
	)
	flag.Parse()

	absFolder, err := filepath.Abs(*folder)
	if err != nil {
		fmt.Printf("Unable to get absolute path: %s\n", err)
		os.Exit(1)
	}

	// Get file system to open and path from folder
	fsRoot := filepath.VolumeName(absFolder)
	fsPath := strings.TrimPrefix(absFolder, fsRoot)
	fsRoot += string(filepath.Separator)

	// fmt.Printf("Open folder %s on file system %s\n", fsPath, fsRoot)

	// Create a browser
	m, err := browser.New(os.DirFS(fsRoot), fsRoot, filepath.ToSlash(fsPath))
	if err != nil {
		fmt.Printf("Error opening folder: %v\n", err)
		os.Exit(1)
	}

	// Open tea with and run the initial model
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
