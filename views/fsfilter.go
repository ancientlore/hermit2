package views

import "strings"

var defaultFilter = ""

// DefaultFilter returns the package-level default wildcard filter string.
func DefaultFilter() string {
	return defaultFilter
}

// SetDefaultFilter sets the package-level default wildcard filter string.
func SetDefaultFilter(filter string) {
	defaultFilter = filter
}

// Filter returns the active wildcard filter string for this view.
func (fsv FS) Filter() string {
	return fsv.filter
}

// SetFilter updates the active and default filter strings, deselects any entries
// that are no longer visible, and rebuilds the sorted visible entry list.
func (fsv *FS) SetFilter(filter string) {
	defaultFilter = filter
	fsv.filter = filter
	fsv.applyFilter()
}

// SyncFilter ensures the view's filter matches the current default filter.
func (fsv *FS) SyncFilter() {
	if fsv.filter != defaultFilter {
		fsv.filter = defaultFilter
		fsv.applyFilter()
	}
}

// applyFilter filters allEntries into entries, deselecting any non-visible entries,
// and sorts the resulting visible entries.
func (fsv *FS) applyFilter() {
	visible := make([]fsEntry, 0, len(fsv.allEntries))
	for i := range fsv.allEntries {
		item := &fsv.allEntries[i]
		// Directories are always shown in the listing even when filtered,
		// matching the original C++ Directory::readDir behavior.
		if item.isDir() || matchFilter(fsv.filter, item.entry.Name()) {
			visible = append(visible, *item)
		} else {
			// Deselect items that are not visible when filtered.
			item.selected = false
		}
	}
	fsv.entries = visible
	fsv.Sort()
}

// matchFilter returns true if name matches the semicolon-separated wildcard filter.
// An empty filter (or "*.*" / "*") matches all files.
func matchFilter(filter, name string) bool {
	if filter == "" {
		return true
	}
	lowerName := strings.ToLower(name)
	hasPattern := false
	for _, pat := range strings.Split(filter, ";") {
		if pat == "" {
			continue
		}
		hasPattern = true
		if pat == "*.*" || pat == "*" {
			return true
		}
		if matchWildcard(strings.ToLower(pat), lowerName) {
			return true
		}
	}
	return !hasPattern
}

// matchWildcard matches name against pattern supporting '*' and '?' wildcards.
func matchWildcard(pattern, name string) bool {
	p := []rune(pattern)
	s := []rune(name)
	pi, si := 0, 0
	starIdx := -1
	matchIdx := 0

	for si < len(s) {
		if pi < len(p) && (p[pi] == '?' || p[pi] == s[si]) {
			pi++
			si++
		} else if pi < len(p) && p[pi] == '*' {
			starIdx = pi
			matchIdx = si
			pi++
		} else if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			si = matchIdx
		} else {
			return false
		}
	}

	for pi < len(p) && p[pi] == '*' {
		pi++
	}

	return pi == len(p)
}

