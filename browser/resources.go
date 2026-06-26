package browser

import (
	_ "embed"
)

// In Go, we use embedding to handle static assets like JS files and Fonts.
// Ensure these files exist in the `embed` subdirectory or adjust paths accordingly.

//go:embed findVisibleInteractiveElements.js
var InteractiveElementsJSCode string

//go:embed fonts/OpenSans-Medium.ttf
var OpenSansFont []byte