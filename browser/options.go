package browser

import (
	"github.com/ancientlore/hermit2/config"
	"github.com/ancientlore/hermit2/views"
)

// ApplyOptions applies the loaded configuration options to the browser's default
// sort, filter, and filter history state.
func ApplyOptions(opts config.Options) {
	views.SetDefaultSort(opts.SortBy, opts.SortDirection)
	views.SetDefaultFilter(opts.Filter)
	filterHistory.SetItems(opts.FilterHistory)
}

// CurrentOptions returns a snapshot of the browser's active sort, filter,
// and filter history state for persistence.
func CurrentOptions() config.Options {
	sortBy, sortDir := views.DefaultSort()
	return config.Options{
		SortBy:        sortBy,
		SortDirection: sortDir,
		Filter:        views.DefaultFilter(),
		FilterHistory: filterHistory.Items(),
	}
}

func saveOptions() error {
	return config.SaveActiveOptions(CurrentOptions())
}

