# HERMIT Go Implementation Analysis & Architecture Recommendations

This document analyzes the initial Go implementation (`hermit2`), evaluates its architecture against Charmbracelet best practices and the original C++ specification, and details strategies for implementing advanced features (e.g., S3 browsers, macro systems, and Yaegi dynamic scripting).

---

## 1. Architectural Analysis of `hermit2`

The `hermit2` codebase is built on the modern Bubble Tea and Lip Gloss ecosystems. Its structural flow operates as follows:

```
[cmd/hermit/main.go]
         │
         ▼ (instantiates & runs)
[browser.Model] ──── embeds ───► [scroller.Model[views.FS]]
         │                                   │
         ├─► NewFileModel()                  ├─► Render() (Lip Gloss formatting)
         ├─► NewHelpModel()                  └─► Update() (PageUp/Dn, offset calculation)
         └─► NewFileInfoModel()
```

### Model Switching Architecture
Rather than nesting panels, `hermit2` employs a **Model Swap** strategy. To transition to a new screen (e.g. drilling down a folder, viewing file info, showing help):
1. The parent model creates a new instance of the child model.
2. It assigns itself to the child's `Prev` field (`newModel.Prev = m`).
3. It returns the child model to the Bubble Tea runtime.
4. When `Left Arrow` or `Esc` is pressed, the child model returns `m.Prev`, popping back to the previous screen.

### Codebase Strengths
* **Decoupled Scrolling:** The `scroller.Model[T Viewer]` is generic, allowing scrolling over any data type implementing the `scroller.Viewer` interface.
* **Syntax Highlighting:** The file viewer utilizes `github.com/alecthomas/chroma` to highlight code contents natively in the terminal.
* **Unix-Windows Build Splitting:** Uses Go build tags (e.g. `file_unix.go`, `file_windows.go`) to compile platform-specific filesystem operations.

---

## 2. Comparison and Optimization Opportunities

Comparing `hermit2` to the original C++ specification reveals several gaps and areas where Go code can be refactored:

### 1. Massive Sorting Boilerplate (`views/fssort.go`)
* **Current Issue:** The Go codebase implements `sort.Interface` (`Len`, `Less`, `Swap`) for eight separate sorting variations (`sortByName`, `sortByNameRev`, `sortByExt`, `sortByExtRev`, etc.). This creates over 300 lines of repetitive code.
* **Go Porting Strategy:** Refactor using Go 1.21 generics and the standard library `slices` package. A single sorting function can accept comparison closures dynamically, removing all boilerplate custom types:
  ```go
  import "slices"

  func (fsv *FS) Sort(sortType string, reverse bool) {
      slices.SortFunc(fsv.entries, func(a, b fs.DirEntry) int {
          // 1. Directories always sort first
          if a.IsDir() && !b.IsDir() { return -1 }
          if !a.IsDir() && b.IsDir() { return 1 }
          
          // 2. Resolve sort criteria
          var cmp int
          switch sortType {
          case "ext":
              cmp = strings.Compare(path.Ext(a.Name()), path.Ext(b.Name()))
          case "size":
              // ... compare sizes
          default:
              cmp = strings.Compare(a.Name(), b.Name())
          }
          if reverse { return -cmp }
          return cmp
      })
  }
  ```

### 2. Missing Dialog Overlays & Shell Macro System
* **Current Issue:** The C++ version relies heavily on modular prompts (`FilterDialog`, `RunDialog`, `SortDialog`, `BookmarkDialog`). `hermit2` currently lacks dialog layouts, printing errors directly to the footer, and lacks macro expansion entirely.
* **Go Porting Strategy:** Rather than creating full terminal screens for dialog inputs, implement transient modal sub-states within a single coordinator model.

---

## 3. Charmbracelet Best Practices & Recommendations

To prepare the codebase for complex features, restructure the application around modern Bubble Tea design patterns:

### 1. Transition to a Coordinator Model
Linked-list model switching (`Prev tea.Model`) can lead to memory leak traps (e.g., if child views keep handles open or large file arrays cached) and makes coordinating global state (like active filters, configuration variables, or bookmark mappings) difficult.

Instead, implement a **Coordinator Shell Model** that manages sub-components:

```go
type ApplicationState int

const (
    StateBrowsing ApplicationState = iota
    StateDialogInput
    StateViewer
)

type ShellModel struct {
    state        ApplicationState
    browser      browser.Model      // Main file grid panel
    viewer       viewport.Model     // Bubbles text viewer
    promptInput  textinput.Model    // Prompt overlay input
    activeDialog string             // "run", "goto", "filter", etc.
}
```
* **Event Routing:** The `ShellModel.Update()` checks the active state. If `StateDialogInput`, key events are forwarded exclusively to `promptInput.Update(msg)`. If `StateBrowsing`, events go to the `browser.Update(msg)`.
* **Resizing Propagation:** Ensures that when `tea.WindowSizeMsg` is received, the shell updates the widths/heights of all nested sub-models uniformly.

### 2. Dialog Overlay Rendering
Rather than clearing the console, render dialog box popups overlaid directly on top of the background file list in the `View()` function using Lip Gloss border layouts:

```go
func (m ShellModel) View() string {
    bgView := m.browser.View()
    
    if m.state == StateDialogInput {
        // Construct a modal box using Lip Gloss
        modal := lipgloss.NewStyle().
            Border(lipgloss.DoubleBorder()).
            BorderForeground(lipgloss.Color("#7D56F4")).
            Padding(1, 2).
            Width(40).
            Render(fmt.Sprintf("%s\n\n%s", m.activeDialogTitle, m.promptInput.View()))
            
        // Layer the modal on top of the background
        return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
    }
    
    return bgView
}
```

---

## 4. Expanding the Browser Abstraction (Multi-Provider Support)

To support filesystems, Amazon S3 buckets, databases, and system processes under a shared concept, design a generalized **`Browser`** and **`Item`** interface. This decouples the scroller UI grid from local OS directories:

```go
package browser

import "time"

// Item represents a single entry in any browseable resource.
type Item interface {
    Name() string
    IsDir() bool
    Size() int64
    ModTime() time.Time
    IsSelected() bool
    SetSelected(bool)
}

// Provider defines the interface for browsing and navigating a resource.
type Provider interface {
    Title() string                                     // Visual label shown in TitleBar
    List() ([]Item, error)                             // Queries list of items
    Enter(item Item) (Provider, error)                 // Navigates down (returns sub-provider)
    Execute(selected []Item, cmd string) (string, error)// Runs operations on target items
}
```

### Implementing Providers
1. **`LocalFSProvider`**: Implements filesystem browsing (wrapping `os.DirFS` and standard IO).
2. **`S3Provider`**: Implements S3 bucket browsing:
   * `List()` calls `ListObjectsV2` for a prefix and returns S3 keys formatted as `Item` folders/files.
   * `Enter()` drills into common prefixes (virtual subfolders).
   * `Execute()` can download items or edit them locally.
3. **`ProcessProvider`**:
   * `List()` returns active OS processes.
   * `Execute(selected, "kill")` terminates the selected processes.

---

## 5. Advanced Macro System & Yaegi Scripting

### 1. Robust Macro Engine
Replace the rigid C++ string parsing with a flexible Go-native macro parser:
* Support standard Go `text/template` tags (e.g. `{{range .Selected}}{{.Name}} {{end}}`).
* Implement interactive pre-scan parameters: scan for `!p`, trigger the shell textinput overlay, and swap the results before calling `os/exec`.

### 2. Yaegi Scripting Engine integration
Embedding **Yaegi** (`github.com/traefik/yaegi`) allows users to write extensions, custom commands, or custom `Browser` providers in **pure Go** and load them dynamically at startup without recompiling:

```go
import "github.com/traefik/yaegi/interp"

func LoadPlugins(pluginPath string) {
    i := interp.New(interp.Options{})
    
    // Read user plugin script (e.g., ~/.config/hermit/plugins/my_s3.go)
    _, err := i.EvalPath(pluginPath)
    if err != nil {
        log.Printf("Failed to load plugin: %v", err)
    }
}
```
Users can now share scripts that define new custom commands or S3 credential configurations, bringing advanced customization to the Go application.
