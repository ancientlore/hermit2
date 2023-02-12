package scroller

// Viewer defines types that can view text data, including
// scrolling and pagination.
type Viewer interface {
	Len() int      // Length of data
	Pos() int      // Position of cursor
	Width() int    // Width of view
	Height() int   // Height of view
	SetWidth(int)  // Set the width of the view
	SetHeight(int) // Set the height of the view
	Home()         // Move cursor to first line
	End()          // Move cursor to last line
	PageUp()       // Move cursor up one page
	PageDown()     // Move cursor down one page
	Up()           // Move cursor up
	Down()         // Move cursor down
	View() string  // Return current page of text
}
