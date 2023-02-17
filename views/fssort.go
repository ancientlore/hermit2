package views

import (
	"path"
	"strings"
)

type sortByName FS

func (e sortByName) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	}
	// File/dir sort next
	return e.entries[i].Name() < e.entries[j].Name()
}

func (e sortByName) Len() int {
	return len(e.entries)
}

func (e sortByName) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortByNameRev FS

func (e sortByNameRev) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	}
	// File/dir sort next
	return e.entries[j].Name() < e.entries[i].Name()
}

func (e sortByNameRev) Len() int {
	return len(e.entries)
}

func (e sortByNameRev) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortByExt FS

func (e sortByExt) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	} else if e.entries[i].IsDir() && e.entries[j].IsDir() {
		return e.entries[i].Name() < e.entries[j].Name()
	}
	// Special files next
	pi := strings.HasPrefix(e.entries[i].Name(), ".")
	pj := strings.HasPrefix(e.entries[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Extensions next
	ei := path.Ext(e.entries[i].Name())
	ej := path.Ext(e.entries[j].Name())
	if ei < ej {
		return true
	} else if ei == ej {
		// Use name when both extensions match
		return e.entries[i].Name() < e.entries[j].Name()
	}
	return false
}

func (e sortByExt) Len() int {
	return len(e.entries)
}

func (e sortByExt) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortByExtRev FS

func (e sortByExtRev) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	}
	// Special files next
	pi := strings.HasPrefix(e.entries[i].Name(), ".")
	pj := strings.HasPrefix(e.entries[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Directory name next
	if e.entries[i].IsDir() && e.entries[j].IsDir() {
		return e.entries[j].Name() < e.entries[i].Name()
	}
	// Extensions next
	ei := path.Ext(e.entries[i].Name())
	ej := path.Ext(e.entries[j].Name())
	if ej < ei {
		return true
	} else if ei == ej {
		// Use name when both extensions match
		return e.entries[j].Name() < e.entries[i].Name()
	}
	return false
}

func (e sortByExtRev) Len() int {
	return len(e.entries)
}

func (e sortByExtRev) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortBySize FS

func (e sortBySize) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	} else if e.entries[i].IsDir() && e.entries[j].IsDir() {
		return e.entries[i].Name() < e.entries[j].Name()
	}
	// Size next
	infoi, _ := e.entries[i].Info()
	infoj, _ := e.entries[j].Info()
	if infoi.Size() < infoj.Size() {
		return true
	} else if infoi.Size() == infoj.Size() {
		// Use name when size is equal
		return e.entries[i].Name() < e.entries[j].Name()
	}
	return false
}

func (e sortBySize) Len() int {
	return len(e.entries)
}

func (e sortBySize) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortBySizeRev FS

func (e sortBySizeRev) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	} else if e.entries[i].IsDir() && e.entries[j].IsDir() {
		return e.entries[i].Name() < e.entries[j].Name()
	}
	// Size next
	infoi, _ := e.entries[i].Info()
	infoj, _ := e.entries[j].Info()
	if infoj.Size() < infoi.Size() {
		return true
	} else if infoi.Size() == infoj.Size() {
		// Use name when size is equal
		return e.entries[i].Name() < e.entries[j].Name()
	}
	return false
}

func (e sortBySizeRev) Len() int {
	return len(e.entries)
}

func (e sortBySizeRev) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortByDate FS

func (e sortByDate) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	}
	// Special files next
	pi := strings.HasPrefix(e.entries[i].Name(), ".")
	pj := strings.HasPrefix(e.entries[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Date next
	infoi, _ := e.entries[i].Info()
	infoj, _ := e.entries[j].Info()
	if infoi.ModTime().Before(infoj.ModTime()) {
		return true
	} else if infoi.ModTime() == infoj.ModTime() {
		// Use name when date is equal
		return e.entries[i].Name() < e.entries[j].Name()
	}
	return false
}

func (e sortByDate) Len() int {
	return len(e.entries)
}

func (e sortByDate) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}

type sortByDateRev FS

func (e sortByDateRev) Less(i, j int) bool {
	// Directories first
	if e.entries[i].IsDir() && !e.entries[j].IsDir() {
		return true
	} else if !e.entries[i].IsDir() && e.entries[j].IsDir() {
		return false
	}
	// Special files next
	pi := strings.HasPrefix(e.entries[i].Name(), ".")
	pj := strings.HasPrefix(e.entries[j].Name(), ".")
	if pi && !pj {
		return true
	} else if !pi && pj {
		return false
	}
	// Date next
	infoi, _ := e.entries[i].Info()
	infoj, _ := e.entries[j].Info()
	if infoj.ModTime().Before(infoi.ModTime()) {
		return true
	} else if infoi.ModTime() == infoj.ModTime() {
		// Use name when date is equal
		return e.entries[i].Name() < e.entries[j].Name()
	}
	return false
}

func (e sortByDateRev) Len() int {
	return len(e.entries)
}

func (e sortByDateRev) Swap(i, j int) {
	e.entries[i], e.entries[j] = e.entries[j], e.entries[i]
	e.selected[i], e.selected[j] = e.selected[j], e.selected[i]
}
