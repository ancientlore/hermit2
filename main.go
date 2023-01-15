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
	rd := filepath.FromSlash("/")
	wd, err := os.Getwd()
	if err == nil {
		a := strings.SplitN(wd, string(filepath.Separator), 2)
		rd = a[0] + string(filepath.Separator)
		if len(a) > 1 {
			wd = a[1]
		} else {
			wd = ""
		}
	}
	var (
		root   = flag.String("root", rd, "Root file system")
		folder = flag.String("path", wd, "Folder")
	)

	flag.Parse()

	fmt.Printf("Opening %q\n", *root)
	m, err := browser.New(os.DirFS(*root), *root, filepath.ToSlash(*folder))
	if err != nil {
		fmt.Printf("Error opening folder: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
