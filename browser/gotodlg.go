package browser

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/ancientlore/hermit2/config"
	"github.com/ancientlore/hermit2/popup"
	"github.com/ancientlore/hermit2/scroller"
)

const (
	gotoPopupWidth  = 76
	gotoPopupHeight = 16
	gotoEditWidth   = gotoPopupWidth - 4
	gotoMaxLen      = 256
	gotoListY       = 5
	gotoListHeight  = 8
)

var (
	gotoHistory = popup.NewHistory()

	gotoItemStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2B55")).
			Foreground(lipgloss.Color("#55FFFF"))

	gotoHistStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2B55")).
			Foreground(lipgloss.Color("#FFFF00"))

	gotoSelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#55FFFF"))

	gotoSelectedHistStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFF00"))

	gotoEmptyStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2B55")).
			Foreground(lipgloss.Color("#AAAAAA"))
)

type gotoSuggestion struct {
	display    string // Rendered label in the list
	targetPath string // Path to open if Enter is pressed
	drillInput string // Edit box text if user drills deeper (Tab / Right / \)
	isHistory  bool   // True when sourced from Goto History
}

// GotoDialogModel is a popup dialog for jumping to a directory with live
// autocomplete suggestions and integrated Goto history.
type GotoDialogModel struct {
	parent       Model
	edit         popup.Edit
	suggestions  []gotoSuggestion
	selected     int
	offset       int
	cachedBase   string
	cachedDirs   []string
	hasCachedDir bool
}

// NewGotoDialogModel creates a new Goto Location dialog overlaid on the given browser Model.
func NewGotoDialogModel(parent Model) GotoDialogModel {
	m := GotoDialogModel{
		parent:   parent,
		edit:     popup.NewEdit(gotoEditWidth, gotoMaxLen, "", nil),
		selected: -1,
	}
	m.updateSuggestions()
	return m
}

// Init initializes the Goto dialog model.
func (m GotoDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles key presses and window resize events for the Goto dialog.
func (m GotoDialogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		mod, cmd := m.parent.Update(msg)
		if bm, ok := mod.(Model); ok {
			m.parent = bm
		}
		return m, cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m.parent, nil

		case "up":
			if m.selected > 0 {
				m.selected--
				m.fixOffset()
			} else if m.selected == 0 {
				m.selected = -1
			}
			return m, nil

		case "down":
			if len(m.suggestions) > 0 && m.selected < len(m.suggestions)-1 {
				m.selected++
				m.fixOffset()
			}
			return m, nil

		case "pgup":
			if len(m.suggestions) > 0 {
				if m.selected <= 0 {
					m.selected = 0
				} else {
					m.selected -= gotoListHeight
					if m.selected < 0 {
						m.selected = 0
					}
				}
				m.fixOffset()
			}
			return m, nil

		case "pgdown":
			if len(m.suggestions) > 0 {
				if m.selected < 0 {
					m.selected = gotoListHeight - 1
				} else {
					m.selected += gotoListHeight
				}
				if m.selected >= len(m.suggestions) {
					m.selected = len(m.suggestions) - 1
				}
				m.fixOffset()
			}
			return m, nil

		case "tab":
			if m.selected >= 0 && m.selected < len(m.suggestions) {
				m.drillInto(m.suggestions[m.selected])
			} else if len(m.suggestions) > 0 {
				m.drillInto(m.suggestions[0])
			}
			return m, nil

		case "right":
			if m.selected >= 0 && m.selected < len(m.suggestions) && m.edit.AtEnd() {
				m.drillInto(m.suggestions[m.selected])
				return m, nil
			}
			m.edit.HandleKey(msg)
			return m, nil

		case `\`, "/":
			if m.selected >= 0 && m.selected < len(m.suggestions) {
				m.drillInto(m.suggestions[m.selected])
				return m, nil
			}
			prev := m.edit.Value()
			m.edit.HandleKey(msg)
			if m.edit.Value() != prev {
				m.selected = -1
				m.offset = 0
				m.updateSuggestions()
			}
			return m, nil

		case "enter":
			target := strings.TrimSpace(m.edit.Value())
			if m.selected >= 0 && m.selected < len(m.suggestions) {
				target = m.suggestions[m.selected].targetPath
			}
			if target == "" {
				return m.parent, nil
			}
			newModel, err := m.navigate(target)
			if err != nil {
				m.parent.footer = err.Error()
				return m.parent, nil
			}
			gotoHistory.Add(canonicalModelPath(newModel))
			if err := saveOptions(); err != nil {
				newModel.footer = err.Error()
			}
			w, h := m.parent.Width(), m.parent.Height()
			if w > 0 && h > 0 {
				sized, _ := newModel.Update(tea.WindowSizeMsg{Width: w, Height: h})
				if bm, ok := sized.(Model); ok {
					*newModel = bm
				}
			}
			sizeCmd := func() tea.Msg { return tea.WindowSizeMsg{Width: w, Height: h} }
			return *newModel, sizeCmd

		default:
			if key.Matches(msg, scroller.DefaultKeyMap.Quit) {
				m.parent.Data.Close()
				return m, tea.Quit
			}
			prev := m.edit.Value()
			m.edit.HandleKey(msg)
			if m.edit.Value() != prev {
				m.selected = -1
				m.offset = 0
				m.updateSuggestions()
			}
		}
	}

	return m, nil
}

// View renders the Goto dialog overlaid on top of the parent browser view.
func (m GotoDialogModel) View() tea.View {
	p := popup.New(gotoPopupWidth, gotoPopupHeight, "Goto Location")

	p.WriteText(2, 2, "Type a directory name. ↑/↓ to select, Tab/→ to drill in.")
	p.DrawEdit(2, 3, &m.edit)

	if len(m.suggestions) == 0 {
		p.WriteText(2, gotoListY, gotoEmptyStyle.Render("  (no matching folders)"))
	} else {
		for row := 0; row < gotoListHeight; row++ {
			idx := m.offset + row
			if idx >= len(m.suggestions) {
				break
			}
			item := m.suggestions[idx]
			isSel := idx == m.selected

			itemSt := gotoItemStyle
			tagSt := gotoHistStyle
			if isSel {
				itemSt = gotoSelectedItemStyle
				tagSt = gotoSelectedHistStyle
			}

			marker := "  "
			if isSel {
				marker = "> "
			}

			var line string
			if item.isHistory {
				prefix := itemSt.Render(marker) + tagSt.Render("* ")
				avail := gotoEditWidth - 4
				text := truncateMiddle(item.display, avail)
				pad := avail - lipgloss.Width(text)
				if pad < 0 {
					pad = 0
				}
				line = prefix + tagSt.Render(text+strings.Repeat(" ", pad))
			} else {
				avail := gotoEditWidth - 2
				text := truncateMiddle(item.display, avail)
				pad := avail - lipgloss.Width(text)
				if pad < 0 {
					pad = 0
				}
				line = itemSt.Render(marker + text + strings.Repeat(" ", pad))
			}
			p.WriteText(2, gotoListY+row, line)
		}
	}

	p.WriteText(2, gotoPopupHeight-2, "Press Enter to go to folder, or Esc to cancel.")

	return p.Overlay(m.parent.View())
}

func (m *GotoDialogModel) drillInto(s gotoSuggestion) {
	m.edit.SetValue(s.drillInput)
	m.selected = -1
	m.offset = 0
	m.updateSuggestions()
}

func (m *GotoDialogModel) fixOffset() {
	if m.selected < 0 {
		m.offset = 0
		return
	}
	if m.selected < m.offset {
		m.offset = m.selected
	}
	if m.selected >= m.offset+gotoListHeight {
		m.offset = m.selected - gotoListHeight + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m *GotoDialogModel) preferredSep(input string) string {
	if strings.Contains(input, "/") && !strings.Contains(input, `\`) {
		return "/"
	}
	if strings.Contains(input, `\`) {
		return `\`
	}
	if strings.Contains(m.parent.Data.Root(), `\`) {
		return `\`
	}
	if strings.Contains(m.parent.Data.Root(), "/") {
		return "/"
	}
	return string(filepath.Separator)
}

func ensureTrailingSep(s, sep string) string {
	s = strings.TrimRight(s, `/\`)
	return s + sep
}

func (m *GotoDialogModel) updateSuggestions() {
	input := m.edit.Value()
	sep := m.preferredSep(input)
	m.suggestions = nil
	seenTargets := make(map[string]bool)

	addSuggestion := func(s gotoSuggestion) {
		key := strings.ToLower(filepath.Clean(s.targetPath))
		if seenTargets[key] {
			return
		}
		seenTargets[key] = true
		m.suggestions = append(m.suggestions, s)
	}

	// 1. When edit box is empty, show Goto History first, then current folder's subdirectories.
	if strings.TrimSpace(input) == "" {
		for _, h := range gotoHistory.Items() {
			addSuggestion(gotoSuggestion{
				display:    h,
				targetPath: h,
				drillInput: ensureTrailingSep(h, sep),
				isHistory:  true,
			})
		}
		curBase := m.currentDirPath()
		for _, dirName := range m.subdirectoriesOf(curBase) {
			target := m.joinTarget(curBase, dirName)
			addSuggestion(gotoSuggestion{
				display:    dirName + sep,
				targetPath: target,
				drillInput: dirName + sep,
			})
		}
		return
	}

	// 2. Non-empty input: split into rawDirPart and prefix.
	var rawDirPart, prefix, baseDir string
	lastSep := strings.LastIndexAny(input, `/\`)

	switch {
	case input == "~":
		rawDirPart = "~" + sep
		prefix = ""
		baseDir = config.HomeFolder()

	case len(input) == 2 && input[1] == ':':
		rawDirPart = input + sep
		prefix = ""
		baseDir = input + string(filepath.Separator)

	case lastSep >= 0:
		rawDirPart = input[:lastSep+1]
		prefix = input[lastSep+1:]
		baseDir = m.resolveDirPath(rawDirPart)

	default:
		rawDirPart = ""
		prefix = input
		baseDir = m.currentDirPath()
	}

	// Allow drilling into ".." when typing "." or ".."
	if prefix == "." || prefix == ".." {
		parentTarget := m.resolveDirPath(rawDirPart + "..")
		addSuggestion(gotoSuggestion{
			display:    rawDirPart + ".." + sep,
			targetPath: parentTarget,
			drillInput: rawDirPart + ".." + sep,
		})
	}

	dirs := m.subdirectoriesOf(baseDir)
	lowerPrefix := strings.ToLower(prefix)

	var prefixMatches, substrMatches []string
	for _, d := range dirs {
		ld := strings.ToLower(d)
		if prefix == "" || strings.HasPrefix(ld, lowerPrefix) {
			prefixMatches = append(prefixMatches, d)
		} else if strings.Contains(ld, lowerPrefix) {
			substrMatches = append(substrMatches, d)
		}
	}

	for _, d := range append(prefixMatches, substrMatches...) {
		target := m.joinTarget(baseDir, d)
		addSuggestion(gotoSuggestion{
			display:    rawDirPart + d + sep,
			targetPath: target,
			drillInput: rawDirPart + d + sep,
		})
	}

	// When typing a bare name (no path separator), also include matching Goto History items.
	if lastSep < 0 && input != "~" {
		for _, h := range gotoHistory.Items() {
			if strings.Contains(strings.ToLower(h), lowerPrefix) {
				addSuggestion(gotoSuggestion{
					display:    h,
					targetPath: h,
					drillInput: ensureTrailingSep(h, sep),
					isHistory:  true,
				})
			}
		}
	}
}

func (m *GotoDialogModel) currentDirPath() string {
	if m.parent.Data.Root() == "/" {
		return path.Clean(path.Join("/", m.parent.Data.Folder()))
	}
	return m.parent.Data.Title()
}

func canonicalModelPath(m *Model) string {
	if m.Data.Root() == "/" {
		return path.Clean(path.Join("/", m.Data.Folder()))
	}
	return m.Data.Title()
}

func (m *GotoDialogModel) joinTarget(baseDir, name string) string {
	if strings.Contains(baseDir, "/") && !strings.Contains(baseDir, `\`) {
		return path.Join(baseDir, name)
	}
	return filepath.Join(baseDir, name)
}

func (m *GotoDialogModel) resolveDirPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return m.currentDirPath()
	}

	// Expand leading ~
	if raw == "~" {
		return config.HomeFolder()
	}
	if strings.HasPrefix(raw, "~/") || strings.HasPrefix(raw, `~\`) {
		home := config.HomeFolder()
		rest := strings.TrimLeft(raw[2:], `/\`)
		if rest == "" {
			return home
		}
		return filepath.Join(home, filepath.FromSlash(rest))
	}

	// Drive letter root like "C:" or "C:\"
	if len(raw) >= 2 && raw[1] == ':' {
		vol := raw[:2]
		rest := strings.TrimLeft(raw[2:], `/\`)
		if rest == "" {
			return vol + string(filepath.Separator)
		}
		return filepath.Clean(vol + string(filepath.Separator) + filepath.FromSlash(rest))
	}

	// UNC path like "\\server\share"
	if strings.HasPrefix(raw, `\\`) {
		return filepath.Clean(raw)
	}

	// Rooted path starting with / or \
	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, `\`) {
		root := m.parent.Data.Root()
		if root == "/" || (strings.Contains(root, "/") && !strings.Contains(root, `\`)) {
			return path.Clean("/" + strings.TrimLeft(filepath.ToSlash(raw), "/"))
		}
		vol := filepath.VolumeName(root)
		rest := strings.TrimLeft(raw, `/\`)
		if rest == "" {
			return vol + string(filepath.Separator)
		}
		return filepath.Clean(vol + string(filepath.Separator) + filepath.FromSlash(rest))
	}

	// Relative path from the current browser folder
	root := m.parent.Data.Root()
	if root == "/" || (strings.Contains(root, "/") && !strings.Contains(root, `\`)) {
		return path.Clean(path.Join("/", m.parent.Data.Folder(), filepath.ToSlash(raw)))
	}
	return filepath.Clean(filepath.Join(m.currentDirPath(), filepath.FromSlash(raw)))
}

func (m *GotoDialogModel) subdirectoriesOf(baseDir string) []string {
	if m.hasCachedDir && m.cachedBase == baseDir {
		return m.cachedDirs
	}

	var dirs []string
	fsys, _, relFolder, useParentFS := m.resolveFSAndFolder(baseDir)
	if useParentFS && fsys != nil {
		rf := strings.TrimPrefix(relFolder, "/")
		if rf == "" {
			rf = "."
		}
		if entries, err := fs.ReadDir(fsys, rf); err == nil {
			for _, e := range entries {
				if isDirEntry(fsys, rf, e) {
					dirs = append(dirs, e.Name())
				}
			}
		}
	} else {
		if entries, err := os.ReadDir(baseDir); err == nil {
			osFS := os.DirFS(baseDir)
			for _, e := range entries {
				if isDirEntry(osFS, ".", e) {
					dirs = append(dirs, e.Name())
				}
			}
		}
	}

	slices.SortFunc(dirs, func(a, b string) int {
		if c := strings.Compare(strings.ToLower(a), strings.ToLower(b)); c != 0 {
			return c
		}
		return strings.Compare(a, b)
	})

	m.cachedBase = baseDir
	m.cachedDirs = dirs
	m.hasCachedDir = true
	return dirs
}

func isDirEntry(fsys fs.FS, dir string, e fs.DirEntry) bool {
	if e.IsDir() {
		return true
	}
	if e.Type()&fs.ModeSymlink != 0 && fsys != nil {
		if info, err := fs.Stat(fsys, path.Join(dir, e.Name())); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func (m *GotoDialogModel) resolveFSAndFolder(resolvedPath string) (fsys fs.FS, fsRoot, folder string, sameFS bool) {
	parentRoot := m.parent.Data.Root()

	// Check if parent uses slash-rooted FS (such as "/" on Unix or fstest.MapFS)
	if parentRoot == "/" {
		cleanSlash := path.Clean("/" + strings.TrimPrefix(filepath.ToSlash(resolvedPath), "/"))
		rf := strings.TrimPrefix(cleanSlash, "/")
		if rf == "" {
			rf = "."
		}
		if info, err := fs.Stat(m.parent.Data.FS(), rf); err == nil && info.IsDir() {
			return m.parent.Data.FS(), "/", cleanSlash, true
		}
	}

	// Check if resolvedPath belongs to the same volume/root as parentRoot
	vol := filepath.VolumeName(resolvedPath)
	parentVol := filepath.VolumeName(parentRoot)
	if vol != "" && strings.EqualFold(vol, parentVol) {
		rel := strings.TrimPrefix(resolvedPath[:len(vol)]+resolvedPath[len(vol):], resolvedPath[:len(vol)])
		slashFolder := filepath.ToSlash(rel)
		if slashFolder == "" {
			slashFolder = "/"
		}
		rf := strings.TrimPrefix(slashFolder, "/")
		if rf == "" {
			rf = "."
		}
		if info, err := fs.Stat(m.parent.Data.FS(), rf); err == nil && info.IsDir() {
			return m.parent.Data.FS(), parentRoot, slashFolder, true
		}
	}

	// Fall back to OS filesystem
	absPath, err := filepath.Abs(resolvedPath)
	if err != nil {
		absPath = resolvedPath
	}
	fsRoot = filepath.VolumeName(absPath)
	fsPath := strings.TrimPrefix(absPath, fsRoot)
	fsRoot += string(filepath.Separator)
	folder = filepath.ToSlash(fsPath)
	if folder == "" {
		folder = "/"
	}
	return os.DirFS(fsRoot), fsRoot, folder, false
}

func (m *GotoDialogModel) navigate(rawTarget string) (*Model, error) {
	resolved := m.resolveDirPath(rawTarget)
	fsys, fsRoot, folder, _ := m.resolveFSAndFolder(resolved)
	return New(fsys, fsRoot, folder)
}

func truncateMiddle(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return "..."[:maxWidth]
	}
	r := []rune(s)
	avail := maxWidth - 3
	front := (avail + 1) / 2
	back := avail - front
	return string(r[:front]) + "..." + string(r[len(r)-back:])
}
