package views

import (
	"cmp"
	"path"
	"slices"
	"strings"
)

// SortType specifies the field used to sort directory entries.
type SortType int

const (
	SortByName SortType = iota
	SortByExt
	SortBySize
	SortByDate
)

// SortDirection specifies whether entries are sorted in forward or reverse order.
type SortDirection int

const (
	SortForward SortDirection = iota
	SortReverse
)

var (
	defaultSortType = SortByName
	defaultSortDir  = SortForward
)

// compareFunc is a generic comparison function following the slices.SortFunc signature.
type compareFunc[T any] func(a, b T) int

// reverseCompare inverts the ordering of a comparison function.
func reverseCompare[T any](fn compareFunc[T]) compareFunc[T] {
	return func(a, b T) int {
		return -fn(a, b)
	}
}

// chainCompare combines multiple comparison functions in priority order,
// returning the first non-zero result.
func chainCompare[T any](fns ...compareFunc[T]) compareFunc[T] {
	return func(a, b T) int {
		for _, fn := range fns {
			if v := fn(a, b); v != 0 {
				return v
			}
		}
		return 0
	}
}

// compareDir sorts directories before non-directories.
func compareDir(a, b fsEntry) int {
	ad, bd := a.isDir(), b.isDir()
	switch {
	case ad && !bd:
		return -1
	case !ad && bd:
		return 1
	default:
		return 0
	}
}

// compareSpecial sorts dot-prefixed entries before regular entries.
func compareSpecial(a, b fsEntry) int {
	pa := strings.HasPrefix(a.entry.Name(), ".")
	pb := strings.HasPrefix(b.entry.Name(), ".")
	switch {
	case pa && !pb:
		return -1
	case !pa && pb:
		return 1
	default:
		return 0
	}
}

// compareName compares entries by name.
func compareName(a, b fsEntry) int {
	return cmp.Compare(a.entry.Name(), b.entry.Name())
}

// compareExt compares entries by file extension (directories fall back to name).
func compareExt(a, b fsEntry) int {
	if a.isDir() && b.isDir() {
		return cmp.Compare(a.entry.Name(), b.entry.Name())
	}
	return cmp.Compare(path.Ext(a.entry.Name()), path.Ext(b.entry.Name()))
}

// compareSize compares entries by file size (directories sort equal by size).
func compareSize(a, b fsEntry) int {
	if a.isDir() && b.isDir() {
		return 0
	}
	switch {
	case a.info == nil && b.info != nil:
		return -1
	case a.info != nil && b.info == nil:
		return 1
	case a.info == nil && b.info == nil:
		return 0
	default:
		return cmp.Compare(a.info.Size(), b.info.Size())
	}
}

// compareDate compares entries by modification time.
func compareDate(a, b fsEntry) int {
	switch {
	case a.info == nil && b.info != nil:
		return -1
	case a.info != nil && b.info == nil:
		return 1
	case a.info == nil && b.info == nil:
		return 0
	default:
		return a.info.ModTime().Compare(b.info.ModTime())
	}
}

func entryComparator(sortType SortType, sortDir SortDirection) compareFunc[fsEntry] {
	withDir := func(fn compareFunc[fsEntry]) compareFunc[fsEntry] {
		if sortDir == SortReverse {
			return reverseCompare(fn)
		}
		return fn
	}

	switch sortType {
	case SortByName:
		return chainCompare(compareDir, compareSpecial, withDir(compareName), compareDate, compareSize, compareExt)
	case SortByExt:
		return chainCompare(compareDir, compareSpecial, withDir(compareExt), withDir(compareName), compareDate, compareSize)
	case SortBySize:
		return chainCompare(compareDir, compareSpecial, withDir(compareSize), compareName, compareDate, compareExt)
	case SortByDate:
		return chainCompare(compareDir, compareSpecial, withDir(compareDate), withDir(compareName), compareSize, compareExt)
	default:
		return chainCompare(compareDir, compareSpecial, withDir(compareName), compareDate, compareSize, compareExt)
	}
}

// SortType returns the current sort type of the view.
func (fsv FS) SortType() SortType {
	return fsv.sortType
}

// SortDirection returns the current sort direction of the view.
func (fsv FS) SortDirection() SortDirection {
	return fsv.sortDir
}

// Sort sorts the directory entries using the current sort type and direction.
func (fsv *FS) Sort() {
	slices.SortFunc(fsv.entries, entryComparator(fsv.sortType, fsv.sortDir))
}

// SetSort updates the active and default sort options and re-sorts the entries.
func (fsv *FS) SetSort(sortType SortType, sortDir SortDirection) {
	defaultSortType = sortType
	defaultSortDir = sortDir
	fsv.sortType = sortType
	fsv.sortDir = sortDir
	fsv.Sort()
}

// SyncSort ensures the view is sorted with the current default sort options.
func (fsv *FS) SyncSort() {
	if fsv.sortType != defaultSortType || fsv.sortDir != defaultSortDir {
		fsv.sortType = defaultSortType
		fsv.sortDir = defaultSortDir
		fsv.Sort()
	}
}
