package scroller

// Viewer defines types that can view text data, including
// scrolling and pagination.
type Viewer interface {
	At(i int) string // Line of text at position i
	Len() int        // Length of data
}
